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
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/location"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/volume"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/scope"
)

// HcloudVolumeReconciler reconciles a HcloudVolume object
type HcloudVolumeReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
	Recorder  record.EventRecorder
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudvolumes/status,verbs=get;update;patch

func (r *HcloudVolumeReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.TODO()
	log := r.Log.WithValues("namespace", req.Namespace, "hcloudVolume", req.Name)

	// Fetch the HcloudVolume instance
	hcloudVolume := &infrav1.HcloudVolume{}
	err := r.Get(ctx, req.NamespacedName, hcloudVolume)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Fetch the Cluster.
	cluster, err := util.GetClusterFromMetadata(ctx, r.Client, hcloudVolume.ObjectMeta)
	if err != nil {
		log.Info("HcloudVolume is missing cluster label or cluster does not exist")
		return reconcile.Result{}, nil
	}
	log = log.WithValues("cluster", cluster.Name)

	hcloudCluster := &infrav1.HcloudCluster{}

	hcloudClusterName := client.ObjectKey{
		Namespace: hcloudVolume.Namespace,
		Name:      cluster.Spec.InfrastructureRef.Name,
	}
	if err := r.Client.Get(ctx, hcloudClusterName, hcloudCluster); err != nil {
		log.Info("HcloudCluster is not available yet")
		return reconcile.Result{}, nil
	}

	log = log.WithValues("hcloudCluster", hcloudCluster.Name)

	// Create the scope.
	volumeScope, err := scope.NewVolumeScope(scope.VolumeScopeParams{
		ClusterScopeParams: scope.ClusterScopeParams{
			Ctx:           ctx,
			Client:        r.Client,
			Logger:        log,
			Recorder:      r.Recorder,
			Cluster:       cluster,
			HcloudCluster: hcloudCluster,
			Packer:        r.Packer,
			Manifests:     r.Manifests,
		},
		HcloudVolume: hcloudVolume,
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
	if !hcloudVolume.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(volumeScope)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(volumeScope)
}

func (r *HcloudVolumeReconciler) reconcileDelete(volumeScope *scope.VolumeScope) (reconcile.Result, error) {
	volumeScope.Info("Reconciling HcloudVolume delete")
	hcloudVolume := volumeScope.HcloudVolume

	// delete servers
	if err := volume.NewService(volumeScope).Delete(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete volume for HcloudVolume %s/%s", hcloudVolume.Namespace, hcloudVolume.Name)
	}

	// delete controlplane loadbalancer
	// Volume is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(volumeScope.HcloudVolume, infrav1.VolumeFinalizer)

	return reconcile.Result{}, nil
}

func (r *HcloudVolumeReconciler) reconcileNormal(volumeScope *scope.VolumeScope) (reconcile.Result, error) {
	volumeScope.Info("Reconciling HcloudVolume")
	hcloudVolume := volumeScope.HcloudVolume

	// If the HcloudVolume doesn't have our finalizer, add it.
	controllerutil.AddFinalizer(hcloudVolume, infrav1.VolumeFinalizer)

	// ensure a valid location is set
	if err := location.NewService(volumeScope).Reconcile(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile location for HcloudVolume %s/%s", hcloudVolume.Namespace, hcloudVolume.Name)
	}

	// reconcile server
	if err := volume.NewService(volumeScope).Reconcile(volumeScope.Ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile volume for HcloudVolume %s/%s", hcloudVolume.Namespace, hcloudVolume.Name)
	}

	return reconcile.Result{}, nil
}

func (r *HcloudVolumeReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(options).
		For(&infrav1.HcloudVolume{}).
		Complete(r)
}
