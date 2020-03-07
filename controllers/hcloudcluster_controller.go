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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/resources/floatingip"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/resources/location"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/resources/network"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/packer"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/record"
)

// HcloudClusterReconciler reconciles a HcloudCluster object
type HcloudClusterReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

func (r *HcloudClusterReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hcloudCluster", req.Name)

	// Fetch the HcloudCluster instance
	hcloudCluster := &infrav1.HcloudCluster{}
	err := r.Get(ctx, req.NamespacedName, hcloudCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster
	cluster, err := util.GetOwnerCluster(ctx, r.Client, hcloudCluster.ObjectMeta)
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
		Ctx:           ctx,
		Client:        r.Client,
		Logger:        log,
		Cluster:       cluster,
		HcloudCluster: hcloudCluster,
		Packer:        r.Packer,
		Manifests:     r.Manifests,
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
	if !hcloudCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(clusterScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(clusterScope)
}

func (r *HcloudClusterReconciler) reconcileDelete(clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling HcloudCluster delete")
	hcloudCluster := clusterScope.HcloudCluster
	ctx := context.TODO()

	// delete controlplane floating IPs
	if err := floatingip.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete floating IPs for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// delete the network
	if err := network.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete network for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// wait for all hcloudMachines to be deleted
	if machines, _, err := clusterScope.ListMachines(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to list machines for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	} else if len(machines) > 0 {
		var names []string
		for _, m := range machines {
			names = append(names, fmt.Sprintf("machine/%s", m.Name))
		}
		record.Eventf(
			hcloudCluster,
			"WaitingForMachineDeletion",
			"Machines %s still running, waiting with deletion of HcloudCluster",
			strings.Join(names, ", "),
		)
		// TODO: send delete for machines still running
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.HcloudCluster, infrav1.ClusterFinalizer)

	return reconcile.Result{}, nil
}

func (r *HcloudClusterReconciler) reconcileNormal(clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling HcloudCluster")
	hcloudCluster := clusterScope.HcloudCluster
	ctx := context.TODO()

	// If the HcloudCluster doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(hcloudCluster, infrav1.ClusterFinalizer)

	// ensure a valid location is set
	if err := location.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// reconcile the network
	if err := network.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile network for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// reconcile the Controlplane Floating IPs
	if err := floatingip.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile floating IPs for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// add the first control plane  floating IP to the status
	if len(hcloudCluster.Status.ControlPlaneFloatingIPs) > 0 {
		hcloudCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
			Host: hcloudCluster.Status.ControlPlaneFloatingIPs[0].IP,
			Port: clusterScope.ControlPlaneAPIEndpointPort(),
		}
	}

	hcloudCluster.Status.Ready = true
	return reconcile.Result{}, nil
}

func (r *HcloudClusterReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	var (
		controlledType     = &infrav1.HcloudCluster{}
		controlledTypeName = reflect.TypeOf(controlledType).Elem().Name()
		controlledTypeGVK  = infrav1.GroupVersion.WithKind(controlledTypeName)
	)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(controlledType).
		// Watch the CAPI resource that owns this infrastructure resource.
		Watches(
			&source.Kind{Type: &clusterv1.Cluster{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.ClusterToInfrastructureMapFunc(controlledTypeGVK),
			},
		).
		Complete(r)
}
