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
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	infrav1 "sigs.k8s.io/cluster-api-provider-hetzner/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/resources/location"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/resources/server"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/scope"
)

// HetznerMachineReconciler reconciles a HetznerMachine object
type HetznerMachineReconciler struct {
	controllerclient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznermachines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznermachines/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machines;machines/status,verbs=get;list;watch

func (r *HetznerMachineReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hetznerMachine", req.Name)

	// Fetch the HetznerMachine instance
	hetznerMachine := &infrav1.HetznerMachine{}
	err := r.Get(ctx, req.NamespacedName, hetznerMachine)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Machine
	machine, err := util.GetOwnerMachine(ctx, r.Client, hetznerMachine.ObjectMeta)
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

	hetznerCluster := &infrav1.HetznerCluster{}

	hetznerClusterName := client.ObjectKey{
		Namespace: hetznerMachine.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hetznerClusterName, hetznerCluster); err != nil {
		log.Info("HetznerCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hetznerCluster", hetznerCluster.Name)

	// Create the scope.
	machineScope, err := scope.NewMachineScope(scope.MachineScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:            ctx,
			Client:         r.Client,
			Logger:         log,
			Cluster:        cluster,
			HetznerCluster: hetznerCluster,
		},
		Machine:        machine,
		HetznerMachine: hetznerMachine,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any AWSMachine changes.
	defer func() {
		if err := machineScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !hetznerMachine.DeletionTimestamp.IsZero() {
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

func (r *HetznerMachineReconciler) reconcileDelete(machineScope *scope.MachineScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling HetznerMachine delete")
	hetznerMachine := machineScope.HetznerMachine

	// delete servers
	if result, err, brk := breakReconcile(server.NewService(machineScope).Reconcile(machineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to reconcile server for HetznerMachine %s/%s", hetznerMachine.Namespace, hetznerMachine.Name)
	}

	if result, err, brk := breakReconcile(server.NewService(machineScope).Delete(machineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to delete servers for HetznerMachine %s/%s", hetznerMachine.Namespace, hetznerMachine.Name)
	}

	// Machine is deleted so remove the finalizer.
	machineScope.HetznerMachine.Finalizers = util.Filter(machineScope.HetznerMachine.Finalizers, infrav1.MachineFinalizer)

	return reconcile.Result{}, nil
}

func (r *HetznerMachineReconciler) reconcileNormal(machineScope *scope.MachineScope) (reconcile.Result, error) {
	machineScope.Info("Reconciling HetznerMachine")
	hetznerMachine := machineScope.HetznerMachine

	// If the HetznerMachine doesn't have our finalizer, add it.
	if !util.Contains(hetznerMachine.Finalizers, infrav1.MachineFinalizer) {
		hetznerMachine.Finalizers = append(hetznerMachine.Finalizers, infrav1.MachineFinalizer)
	}

	// ensure a valid location is set
	if err := location.NewService(machineScope).Reconcile(machineScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for HetznerMachine %s/%s", hetznerMachine.Namespace, hetznerMachine.Name)
	}

	// reconcile server
	if result, err, brk := breakReconcile(server.NewService(machineScope).Reconcile(machineScope.Ctx)); brk {
		return result, errors.Wrapf(err, "failed to reconcile server for HetznerMachine %s/%s", hetznerMachine.Namespace, hetznerMachine.Name)
	}

	return reconcile.Result{}, nil
}

func (r *HetznerMachineReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.HetznerMachine{}).
		Watches(
			&source.Kind{Type: &clusterv1.Machine{}},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: util.MachineToInfrastructureMapFunc(infrav1.GroupVersion.WithKind("HetznerMachine")),
			},
		).
		Watches(
			&source.Kind{Type: &infrav1.HetznerCluster{}},
			&handler.EnqueueRequestsFromMapFunc{ToRequests: handler.ToRequestsFunc(r.HetznerClusterToHetznerMachines)},
		).
		Complete(r)
}

// HetznerClusterToHetznerMachine is a handler.ToRequestsFunc to be used to
// enqeue requests for reconciliation of HetznerMachines.
func (r *HetznerMachineReconciler) HetznerClusterToHetznerMachines(o handler.MapObject) []ctrl.Request {
	result := []ctrl.Request{}

	c, ok := o.Object.(*infrav1.HetznerCluster)
	if !ok {
		r.Log.Error(errors.Errorf("expected a HetznerCluster but got a %T", o.Object), "failed to get HetznerMachine for HetznerCluster")
		return nil
	}
	log := r.Log.WithValues("HetznerCluster", c.Name, "Namespace", c.Namespace)

	cluster, err := util.GetOwnerCluster(context.TODO(), r.Client, c.ObjectMeta)
	switch {
	case apierrors.IsNotFound(err) || cluster == nil:
		return result
	case err != nil:
		log.Error(err, "failed to get owning cluster")
		return result
	}

	labels := map[string]string{clusterv1.MachineClusterLabelName: cluster.Name}
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
