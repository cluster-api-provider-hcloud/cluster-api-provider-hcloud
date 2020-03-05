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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha2"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/floatingip"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/location"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/network"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/manifests"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/packer"
)

// HetznerClusterReconciler reconciles a HetznerCluster object
type HetznerClusterReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznerclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznerclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

func (r *HetznerClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hetznerCluster", req.Name)

	// Fetch the HetznerCluster instance
	hetznerCluster := &infrav1.HetznerCluster{}
	err := r.Get(ctx, req.NamespacedName, hetznerCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster
	cluster, err := util.GetOwnerCluster(ctx, r.Client, hetznerCluster.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if cluster == nil {
		log.Info("Cluster Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("cluster", cluster.Name)

	// Create the scope.
	clusterScope, err := scope.NewClusterScope(scope.ClusterScopeParams{
		Ctx:            ctx,
		Client:         r.Client,
		Logger:         log,
		Cluster:        cluster,
		HetznerCluster: hetznerCluster,
		Packer:         r.Packer,
		Manifests:      r.Manifests,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any AWSMachine changes.
	defer func() {
		if err := clusterScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !hetznerCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(clusterScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(clusterScope)
}

func (r *HetznerClusterReconciler) reconcileDelete(clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling HetznerCluster delete")
	hetznerCluster := clusterScope.HetznerCluster
	ctx := context.TODO()

	// delete controlplane floating IPs
	if err := floatingip.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete floating IPs for HetznerCluster %s/%s", hetznerCluster.Namespace, hetznerCluster.Name)
	}

	// delete the network
	if err := network.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete network for HetznerCluster %s/%s", hetznerCluster.Namespace, hetznerCluster.Name)
	}

	// Cluster is deleted so remove the finalizer.
	clusterScope.HetznerCluster.Finalizers = util.Filter(clusterScope.HetznerCluster.Finalizers, infrav1.ClusterFinalizer)

	return reconcile.Result{}, nil
}

func (r *HetznerClusterReconciler) reconcileNormal(clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling HetznerCluster")
	hetznerCluster := clusterScope.HetznerCluster
	ctx := context.TODO()

	// If the AWSCluster doesn't have our finalizer, add it.
	if !util.Contains(hetznerCluster.Finalizers, infrav1.ClusterFinalizer) {
		hetznerCluster.Finalizers = append(hetznerCluster.Finalizers, infrav1.ClusterFinalizer)
	}

	// ensure a valid location is set
	if err := location.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for HetznerCluster %s/%s", hetznerCluster.Namespace, hetznerCluster.Name)
	}

	// reconcile the network
	if err := network.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile network for HetznerCluster %s/%s", hetznerCluster.Namespace, hetznerCluster.Name)
	}

	// reconcile the Controlplane Floating IPs
	if err := floatingip.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile floating IPs for HetznerCluster %s/%s", hetznerCluster.Namespace, hetznerCluster.Name)
	}

	// add the first controlplan floating IP to the status
	if len(hetznerCluster.Status.ControlPlaneFloatingIPs) > 0 {
		hetznerCluster.Status.APIEndpoints = []clusterv1.APIEndpoint{{
			Host: hetznerCluster.Status.ControlPlaneFloatingIPs[0].IP,
			Port: clusterScope.ControlPlaneAPIEndpointPort(),
		}}
	}

	hetznerCluster.Status.Ready = true
	return reconcile.Result{}, nil
}

func (r *HetznerClusterReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.HetznerCluster{}).
		Complete(r)
}
