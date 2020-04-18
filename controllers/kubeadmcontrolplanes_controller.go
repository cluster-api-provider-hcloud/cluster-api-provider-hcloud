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
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
)

// KubeadmControlPlaneReconciler TODO
type KubeadmControlPlaneReconciler struct {
	controllerclient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *KubeadmControlPlaneReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	log := r.Log.WithValues("namespace", req.Namespace, "kubeadmControlPlane", req.Name)

	// Fetch the KubeadmControlPlane instance.
	kcp := &controlplanev1.KubeadmControlPlane{}
	if err := r.Client.Get(ctx, req.NamespacedName, kcp); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Fetch the Cluster
	cluster, err := util.GetOwnerCluster(ctx, r.Client, kcp.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)

	// end here if we are not using an HcloudCluster as infrastructure
	if cluster.Spec.InfrastructureRef.Kind != "HcloudCluster" || strings.Split(cluster.Spec.InfrastructureRef.APIVersion, "/")[0] != infrav1.GroupVersion.Group {
		return reconcile.Result{}, nil
	}

	// Handle deleted KubeadmControlPlanes
	if !kcp.DeletionTimestamp.IsZero() {
		// Cluster is deleted so remove the finalizer.
		controllerutil.RemoveFinalizer(kcp, infrav1.KubeadmControlPlaneFinalizer)
		return reconcile.Result{}, nil
	}

	// Handle non-deleted KubeadmControlPlanes
	patchHelper, err := patch.NewHelper(kcp, r.Client)
	if err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to init patch helper")
	}
	// Set finalizer
	controllerutil.AddFinalizer(kcp, infrav1.KubeadmControlPlaneFinalizer)
	// Register the finalizer immediately to avoid orphaning resources on delete
	if err := patchHelper.Patch(ctx, kcp); err != nil {
		return reconcile.Result{}, err
	}

	// Fetch the HcloudCluster instance
	hcloudCluster := &infrav1.HcloudCluster{}
	hcloudNamespacedName := types.NamespacedName{Namespace: req.Namespace, Name: cluster.Spec.InfrastructureRef.Name}
	err = r.Get(ctx, hcloudNamespacedName, hcloudCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	log = log.WithValues("hcloudCluster", cluster.Name)

	return reconcile.Result{}, nil
}

func (r *KubeadmControlPlaneReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&controlplanev1.KubeadmControlPlane{}).
		Complete(r)
}
