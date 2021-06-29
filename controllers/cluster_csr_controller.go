/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	certificatesv1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha4"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/csr"
)

type ManagementCluster interface {
	controllerclient.Client
	Eventf(eventtype, reason, message string, args ...interface{})
	Event(eventtype, reason, message string)
	Namespace() string
}

type GuestCSRReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	mCluster  ManagementCluster
	clientSet *kubernetes.Clientset
}

func (r *GuestCSRReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {

	log := r.Log.WithValues("csr", req.Name)

	// Fetch the HcloudCluster instance
	certificateSigningRequest := &certificatesv1.CertificateSigningRequest{}
	err := r.Get(ctx, req.NamespacedName, certificateSigningRequest)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// skip CSR that have already been decided
	if len(certificateSigningRequest.Status.Conditions) > 0 {
		return reconcile.Result{}, nil
	}
	nodePrefix := "system:node:"
	// skip CSR from non-nodes
	if !strings.HasPrefix(certificateSigningRequest.Spec.Username, nodePrefix) {
		return reconcile.Result{}, nil
	}

	// find matching HcloudMachine object
	var hcloudMachine infrav1.HcloudMachine
	if err := r.mCluster.Get(ctx, types.NamespacedName{
		Namespace: r.mCluster.Namespace(),
		Name:      strings.TrimPrefix(certificateSigningRequest.Spec.Username, nodePrefix),
	}, &hcloudMachine); err != nil {
		return reconcile.Result{}, err
	}

	csrBlock, _ := pem.Decode(certificateSigningRequest.Spec.Request)

	csrRequest, err := x509.ParseCertificateRequest(csrBlock.Bytes)
	if err != nil {
		r.mCluster.Eventf(
			corev1.EventTypeWarning,
			"CSRParsingError",
			"Error parsing CertificateSigningRequest %s: %s",
			req.Name,
			err,
		)
		return reconcile.Result{}, err
	}

	var condition = certificatesv1.CertificateSigningRequestCondition{
		LastUpdateTime: metav1.Time{Time: time.Now()},
	}
	if err := csr.ValidateKubeletCSR(csrRequest, &hcloudMachine); err != nil {
		condition.Type = certificatesv1.CertificateDenied
		condition.Reason = "CSRValidationFailed"
		condition.Message = fmt.Sprintf("Validation by cluster-api-provider-hcloud failed: %s", err)
	} else {
		condition.Type = certificatesv1.CertificateApproved
		condition.Reason = "CSRValidationSucceed"
		condition.Message = "Validation by cluster-api-provider-hcloud was successful"
	}

	certificateSigningRequest.Status.Conditions = append(
		certificateSigningRequest.Status.Conditions,
		condition,
	)

	if _, err := r.clientSet.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(ctx, certificateSigningRequest, metav1.UpdateOptions{}); err != nil {
		log.Error(err, "updating approval of csr failed", "username", certificateSigningRequest.Spec.Username, "csr")
	}

	return reconcile.Result{}, nil
}

func (r *GuestCSRReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	var (
		controlledType = &certificatesv1.CertificateSigningRequest{}
		// controlledTypeName = reflect.TypeOf(controlledType).Elem().Name()
		// controlledTypeGVK  = certificatesv1.GroupVersion.WithKind(controlledTypeName)
	)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(controlledType).
		Complete(r)
}
