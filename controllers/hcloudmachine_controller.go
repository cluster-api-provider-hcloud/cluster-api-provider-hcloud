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

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/resources/server"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/packer"
)

// HcloudMachineReconciler reconciles a HcloudMachine object
type HcloudMachineReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudmachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.swine.dev,resources=hcloudmachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch

func (r *HcloudMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hcloudMachine", req.Name)

	// Fetch the HcloudMachine instance
	hcloudMachine := &infrav1.HcloudMachine{}
	err := r.Get(ctx, req.NamespacedName, hcloudMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Machine
	machine, err := util.GetOwnerMachine(ctx, r.Client, hcloudMachine.ObjectMeta)
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

	if util.IsPaused(cluster, hcloudMachine) {
		log.Info("HcloudMachine or linked Cluster is marked as paused. Won't reconcile")
		return ctrl.Result{}, nil
	}

	hcloudCluster := &infrav1.HcloudCluster{}

	hcloudClusterName := client.ObjectKey{
		Namespace: hcloudMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hcloudClusterName, hcloudCluster); err != nil {
		log.Info("HcloudCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hcloudCluster", hcloudCluster.Name)

	// Create the scope.
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:           ctx,
			Client:        r.Client,
			Logger:        log,
			Cluster:       cluster,
			HcloudCluster: hcloudCluster,
			Packer:        r.Packer,
			Manifests:     r.Manifests,
		},
		Machine:       machine,
		HcloudMachine: hcloudMachine,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any HcloudMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !hcloudMachine.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(machineScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(machineScope)
}

func breakReconcile(ctrl *reconcile.Result, err error) (reconcile.Result, error, bool) {
	c := reconcile.Result{}
	if ctrl != nil {
		c = *ctrl
	}
	return c, err, ctrl != nil || err != nil
}

func (r *HcloudMachineReconciler) reconcileDelete(machineScope *scope.MachineScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling HcloudMachine delete")
	hcloudMachine := machineScope.HcloudMachine

	// delete servers
	if result, err, brk := breakReconcile(server.NewService(machineScope).Delete(machineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to delete servers for HcloudMachine %s/%s", hcloudMachine.Namespace, hcloudMachine.Name)
	}

	// Machine is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(machineScope.HcloudMachine, infrav1.MachineFinalizer)

	return reconcile.Result{}, nil
}

func (r *HcloudMachineReconciler) reconcileNormal(machineScope *scope.MachineScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling HcloudMachine")
	hcloudMachine := machineScope.HcloudMachine

	// If the HcloudMachine doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(machineScope.HcloudMachine, infrav1.MachineFinalizer)

	// Register the finalizer immediately to avoid orphaning Hcloud resources
	// on delete
	if err := machineScope.PatchObject(machineScope.Ctx); err != nil {
		return ctrl.Result{}, err
	}

	// reconcile server
	if result, err, brk := breakReconcile(server.NewService(machineScope).Reconcile(machineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to reconcile server for HcloudMachine %s/%s", hcloudMachine.Namespace, hcloudMachine.Name)
	}

	return reconcile.Result{}, nil
}

func (r *HcloudMachineReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.HcloudMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("HcloudMachine")),
			},
		).
		Watches(
			&source.Kind{Type: &infrav1.HcloudCluster{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.HcloudClusterToHcloudMachines)},
		).
		Complete(r)
}

// HcloudClusterToHcloudMachine is a handler.ToRequestsFunc to be used to
// enqeue requests for reconciliation of HcloudMachines.
func (r *HcloudMachineReconciler) HcloudClusterToHcloudMachines(o handler.MapObject) []ctrl.Request {
	result := []ctrl.Request{}

	c, ok := o.Object.(*infrav1.HcloudCluster)
	if !ok {
		r.Log.Error(errors.Errorf("expected a HcloudCluster but got a %T", o.Object), "failed to get HcloudMachine for HcloudCluster")
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

	labels := map[string]string{clusterv1.ClusterLabelName: cluster.Name}
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
