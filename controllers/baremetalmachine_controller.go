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
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/baremetal"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer"
)

// BareMetalMachineReconciler reconciles a BareMetalMachine object
type BareMetalMachineReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=baremetalmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=baremetalmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch

func (r *BareMetalMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "bareMetalMachine", req.Name)

	// Fetch the BareMetalMachine instance
	bareMetalMachine := &infrav1.BareMetalMachine{}
	err := r.Get(ctx, req.NamespacedName, bareMetalMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Machine
	machine, err := util.GetOwnerMachine(ctx, r.Client, bareMetalMachine.ObjectMeta)
	if err != nil {
		return reconcile.Result{}, err
	}
	if machine == nil {
		log.Info("Machine Controller has not yet set OwnerRef")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("machine", machine.Name)

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, machine.ObjectMeta)
	if err != nil {
		log.Info("Machine is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)

	if util.IsPaused(cluster, bareMetalMachine) {
		log.Info("BareMetalMachine or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	hcloudCluster := &infrav1.HcloudCluster{}

	hcloudClusterName := client.ObjectKey{
		Namespace: bareMetalMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hcloudClusterName, hcloudCluster); err != nil {
		log.Info("HcloudCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hcloudCluster", hcloudCluster.Name)

	// Create the scope.
	bareMetalMachineScope, err := scope.NewBareMetalMachineScope(scope.BareMetalMachineScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:           ctx,
			Client:        r.Client,
			Logger:        log,
			Cluster:       cluster,
			HcloudCluster: hcloudCluster,
			Packer:        r.Packer,
			Manifests:     r.Manifests,
		},
		Machine:          machine,
		BareMetalMachine: bareMetalMachine,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any BareMetalMachine changes.
	defer func() {
		if err := bareMetalMachineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !bareMetalMachine.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(bareMetalMachineScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(bareMetalMachineScope)
}

func (r *BareMetalMachineReconciler) reconcileDelete(bareMetalMachineScope *scope.BareMetalMachineScope) (reconcile.Result, error) {
	bareMetalMachineScope.Info("Reconciling BareMetalMachine delete")
	bareMetalMachine := bareMetalMachineScope.BareMetalMachine

	// delete bare metal servers
	if result, err, brk := breakReconcile(baremetal.NewService(bareMetalMachineScope).Delete(bareMetalMachineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to delete servers for BareMetalMachine %s/%s", bareMetalMachine.Namespace, bareMetalMachine.Name)
	}

	// Machine is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(bareMetalMachineScope.BareMetalMachine, infrav1.MachineFinalizer)

	return reconcile.Result{}, nil
}

func (r *BareMetalMachineReconciler) reconcileNormal(bareMetalMachineScope *scope.BareMetalMachineScope) (reconcile.Result, error) {
	bareMetalMachineScope.Info("Reconciling BareMetalMachine")
	bareMetalMachine := bareMetalMachineScope.BareMetalMachine

	// If the BareMetalMachine doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(bareMetalMachineScope.BareMetalMachine, infrav1.MachineFinalizer)

	// Register the finalizer immediately to avoid orphaning Hcloud resources
	// on delete
	if err := bareMetalMachineScope.PatchObject(bareMetalMachineScope.Ctx); err != nil {
		return ctrl.Result{}, err
	}

	// reconcile bare metal server
	if result, err, brk := breakReconcile(baremetal.NewService(bareMetalMachineScope).Reconcile(bareMetalMachineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to reconcile server for BareMetalMachine %s/%s", bareMetalMachine.Namespace, bareMetalMachine.Name)
	}

	return reconcile.Result{}, nil
}

func (r *BareMetalMachineReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.BareMetalMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("BareMetalMachine")),
			},
		).
		Watches(
			&source.Kind{Type: &infrav1.HcloudCluster{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.HcloudClusterToBareMetalMachines)},
		).
		Complete(r)
}

// HcloudClusterToBareMetalMachine is a handler.ToRequestsFunc to be used to
// enqeue requests for reconciliation of BareMetalMachines.
func (r *BareMetalMachineReconciler) HcloudClusterToBareMetalMachines(o handler.MapObject) []ctrl.Request {
	result := []ctrl.Request{}

	c, ok := o.Object.(*infrav1.HcloudCluster)
	if !ok {
		r.Log.Error(errors.Errorf("expected a HcloudCluster but got a %T", o.Object), "failed to get BareMetalMachine for HcloudCluster")
		return nil
	}
	log := r.Log.WithValues("HcloudCluster", c.Name, "Namespace", c.Namespace)

	cluster, err := util.GetOwnerCluster(context.TODO(), r.Client, c.ObjectMeta)
	switch {
	case apierrors.IsNotFound(err) || cluster == nil:
		return result
	case err != nil:
		log.Error(err, "failed to get owning cluster")
		return result
	}

	labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name, "nodepool": "worker-bm"}
	machineList := &clusterv1.MachineList{}
	if err := r.List(context.TODO(), machineList, client.InNamespace(c.Namespace), client.MatchingLabels(labels)); err != nil {
		log.Error(err, "failed to list Machines")
		return nil
	}
	for _, m := range machineList.Items {
		if m.Spec.InfrastructureRef.Name == "" {
			continue
		}
		name := client.ObjectKey{Namespace: m.Namespace, Name: m.Spec.InfrastructureRef.Name}
		result = append(result, ctrl.Request{NamespacedName: name})
	}

	return result
}
