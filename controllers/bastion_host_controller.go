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
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/bastionhost"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/location"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer"
)

// BastionHostReconciler reconciles a BastionHost object
type BastionHostReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=bastionhosts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=bastionhosts/status,verbs=get;update;patch
func (r *BastionHostReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "bastionHost", req.Name)

	// Fetch the BastionHost instance
	bastionHost := &infrav1.BastionHost{}
	err := r.Get(ctx, req.NamespacedName, bastionHost)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Machine
	machine, err := util.GetOwnerMachine(ctx, r.Client, bastionHost.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}
	log = log.WithValues("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		fmt.Println(err)
		log.Info("Bastion host is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)
	hcloudCluster := &infrav1.HcloudCluster{}

	hcloudClusterName := client.ObjectKey{
		Namespace: bastionHost.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hcloudClusterName, hcloudCluster); err != nil {
		log.Info("HcloudCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hcloudCluster", hcloudCluster.Name)
	// Create the scope.
	bastionHostScope, err := scope.NewBastionHostScope(scope.BastionHostScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:           ctx,
			Client:        r.Client,
			Logger:        log,
			Cluster:       cluster,
			HcloudCluster: hcloudCluster,
			Packer:        r.Packer,
			Manifests:     r.Manifests,
		},
		BastionHost: bastionHost,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any AWSVolume changes.
	defer func() {
		if err := bastionHostScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !bastionHost.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(bastionHostScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(bastionHostScope)
}

func (r *BastionHostReconciler) reconcileDelete(bastionHostScope *scope.BastionHostScope) (reconcile.Result, error) {
	bastionHostScope.Info("Reconciling BastionHost delete")
	bastionHost := bastionHostScope.BastionHost

	// delete bastionHost
	if result, err, brk := breakReconcile(bastionhost.NewService(bastionHostScope).Delete(bastionHostScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to delete bastion host %s/%s", bastionHost.Namespace, bastionHost.Name)
	}

	// Bastion host is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(bastionHostScope.BastionHost, infrav1.VolumeFinalizer)

	return reconcile.Result{}, nil
}

func (r *BastionHostReconciler) reconcileNormal(bastionHostScope *scope.BastionHostScope) (reconcile.Result, error) {
	bastionHostScope.Info("Reconciling BastionHost")
	bastionHost := bastionHostScope.BastionHost

	// If the Bastion Host doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(bastionHost, infrav1.VolumeFinalizer)

	// ensure a valid location is set
	if err := location.NewService(bastionHostScope).Reconcile(bastionHostScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for bastion host %s/%s", bastionHost.Namespace, bastionHost.Name)
	}

	// reconcile bastion host
	if result, err, brk := breakReconcile(bastionhost.NewService(bastionHostScope).Reconcile(bastionHostScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to reconcile bastion host %s/%s", bastionHost.Namespace, bastionHost.Name)
	}

	return reconcile.Result{}, nil
}

func (r *BastionHostReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {

	var (
		controlledType     = &infrav1.HcloudCluster{}
		controlledTypeName = reflect.TypeOf(controlledType).Elem().Name()
		controlledTypeGVK  = infrav1.GroupVersion.WithKind(controlledTypeName)
	)
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.BastionHost{}).
		// Watch the CAPI resource that owns this infrastructure resource.
		Watches(
			&source.Kind{Type: &clusterv1.Cluster{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.ClusterToInfrastructureMapFunc(controlledTypeGVK),
			},
		).Complete(r)
}
