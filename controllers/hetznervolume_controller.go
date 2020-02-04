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
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/location"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/resources/volume"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
)

// HetznerVolumeReconciler reconciles a HetznerVolume object
type HetznerVolumeReconciler struct {
	controllerclient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznervolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznervolumes/status,verbs=get;update;patch

func (r *HetznerVolumeReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hetznerVolume", req.Name)

	// Fetch the HetznerVolume instance
	hetznerVolume := &infrav1.HetznerVolume{}
	err := r.Get(ctx, req.NamespacedName, hetznerVolume)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, hetznerVolume.ObjectMeta)
	if err != nil {
		log.Info("HetznerVolume is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)

	hetznerCluster := &infrav1.HetznerCluster{}

	hetznerClusterName := client.ObjectKey{
		Namespace: hetznerVolume.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hetznerClusterName, hetznerCluster); err != nil {
		log.Info("HetznerCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hetznerCluster", hetznerCluster.Name)

	// Create the scope.
	volumeScope, err := scope.NewVolumeScope(scope.VolumeScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:            ctx,
			Client:         r.Client,
			Logger:         log,
			Cluster:        cluster,
			HetznerCluster: hetznerCluster,
		},
		HetznerVolume: hetznerVolume,
	})
	if err != nil {
		return reconcile.Result{}, errors.Errorf("failed to create scope: %+v", err)
	}

	// Always close the scope when exiting this function so we can persist any AWSVolume changes.
	defer func() {
		if err := volumeScope.Close(); err != nil && reterr == nil {
			reterr = err
		}
	}()

	// Handle deleted clusters
	if !hetznerVolume.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(volumeScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(volumeScope)
}

func (r *HetznerVolumeReconciler) reconcileDelete(volumeScope *scope.VolumeScope) (reconcile.Result, error) {
	volumeScope.Info("Reconciling HetznerVolume delete")
	hetznerVolume := volumeScope.HetznerVolume

	// delete servers
	if err := volume.NewService(volumeScope).Delete(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete volume for HetznerVolume %s/%s", hetznerVolume.Namespace, hetznerVolume.Name)
	}

	// delete controlplane floating IPs
	// Volume is deleted so remove the finalizer.
	volumeScope.HetznerVolume.Finalizers = util.Filter(volumeScope.HetznerVolume.Finalizers, infrav1.VolumeFinalizer)

	return reconcile.Result{}, nil
}

func (r *HetznerVolumeReconciler) reconcileNormal(volumeScope *scope.VolumeScope) (reconcile.Result, error) {
	volumeScope.Info("Reconciling HetznerVolume")
	hetznerVolume := volumeScope.HetznerVolume

	// If the HetznerVolume doesn't have our finalizer, add it.
	if !util.Contains(hetznerVolume.Finalizers, infrav1.VolumeFinalizer) {
		hetznerVolume.Finalizers = append(hetznerVolume.Finalizers, infrav1.VolumeFinalizer)
	}

	// ensure a valid location is set
	if err := location.NewService(volumeScope).Reconcile(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for HetznerVolume %s/%s", hetznerVolume.Namespace, hetznerVolume.Name)
	}

	// reconcile server
	if err := volume.NewService(volumeScope).Reconcile(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile volume for HetznerVolume %s/%s", hetznerVolume.Namespace, hetznerVolume.Name)
	}

	return reconcile.Result{}, nil
}

func (r *HetznerVolumeReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.HetznerVolume{}).
		Complete(r)
}
