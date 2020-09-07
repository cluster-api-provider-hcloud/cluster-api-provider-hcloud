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
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	certificatesv1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	recorder "k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	loadbalancer "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/loadbalancer"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/location"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/resources/network"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/manifests"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer"
)

var errNoReadyAPIServer = errors.New("No ready API server was found")

// HcloudClusterReconciler reconciles a HcloudCluster object
type HcloudClusterReconciler struct {
	controllerclient.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Packer    *packer.Packer
	Manifests *manifests.Manifests
	recorder  record.EventRecorder

	targetClusterManagersStopCh map[types.NamespacedName]chan struct{}
	targetClusterManagersLock   sync.Mutex
}

// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster-api-provider-hcloud.capihc.com,resources=hcloudclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create

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

	// Initialize Packer
	if err := r.Packer.Initialize(hcloudCluster); err != nil {
		log.Error(err, "unable to initialise packer manager")
		os.Exit(1)
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

	if util.IsPaused(cluster, hcloudCluster) {
		log.Info("HcloudCluster or linked Cluster is marked as paused. Won't reconcile")
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

	// Always close the scope when exiting this function so we can persist any HcloudCluster changes.
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

	// stop targetClusterManager
	if err := r.reconcileTargetClusterManager(clusterScope); err != nil {
		return reconcile.Result{}, err
	}

	// wait for all hcloudMachines to be deleted
	if machines, _, err := clusterScope.ListMachines(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to list machines for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	} else if len(machines) > 0 {
		var names []string
		for _, m := range machines {
			names = append(names, fmt.Sprintf("machine/%s", m.Name))
		}
		r.recorder.Eventf(
			hcloudCluster,
			corev1.EventTypeNormal,
			"WaitingForMachineDeletion",
			"Machines %s still running, waiting with deletion of HcloudCluster",
			strings.Join(names, ", "),
		)
		// TODO: send delete for machines still running
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// delete load balancers
	if err := loadbalancer.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete load balancers for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// delete the network
	if err := network.NewService(clusterScope).Delete(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to delete network for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// Cluster is deleted so remove the finalizer.
	controllerutil.RemoveFinalizer(clusterScope.HcloudCluster, infrav1.ClusterFinalizer)

	return reconcile.Result{}, nil
}

func (r *HcloudClusterReconciler) reconcileNormal(clusterScope *scope.ClusterScope) (reconcile.Result, error) {
	clusterScope.Info("Reconciling HcloudCluster")
	hcloudCluster := clusterScope.HcloudCluster
	ctx := clusterScope.Ctx

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

	// reconcile the load balancers
	if err := loadbalancer.NewService(clusterScope).Reconcile(ctx); err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile load balancers for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// add the IPv4 of the first (the main) load balancer as host of API endpoint as control plane endpoint
	if len(hcloudCluster.Status.ControlPlaneLoadBalancers) > 0 {
		hcloudCluster.Spec.ControlPlaneEndpoint = clusterv1.APIEndpoint{
			Host: hcloudCluster.Status.ControlPlaneLoadBalancers[0].IPv4,
			Port: clusterScope.ControlPlaneAPIEndpointPort(),
		}
	}

	// set cluster infrastructure as ready
	hcloudCluster.Status.Ready = true

	// reconcile cluster manifests
	if err := r.reconcileManifests(clusterScope); err == errNoReadyAPIServer {
		r.recorder.Eventf(
			hcloudCluster,
			corev1.EventTypeNormal,
			"APIServerNotReady",
			"No ready API server available yet to reconcile: %s",
			err,
		)
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	} else if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "failed to reconcile manifests for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	// start targetClusterManager
	if err := r.reconcileTargetClusterManager(clusterScope); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *HcloudClusterReconciler) reconcileTargetClusterManager(clusterScope *scope.ClusterScope) error {
	hcloudCluster := clusterScope.HcloudCluster
	deleted := !hcloudCluster.DeletionTimestamp.IsZero()

	r.targetClusterManagersLock.Lock()
	defer r.targetClusterManagersLock.Unlock()
	key := types.NamespacedName{
		Namespace: hcloudCluster.Namespace,
		Name:      hcloudCluster.Name,
	}
	if stopCh, ok := r.targetClusterManagersStopCh[key]; !ok && !deleted {
		// create a new cluster manager
		m, err := r.newTargetClusterManager(clusterScope)
		if err != nil {
			return errors.Wrapf(err, "failed to create a clusterManager for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
		}
		r.targetClusterManagersStopCh[key] = make(chan struct{})

		go func() {
			if err := m.Start(r.targetClusterManagersStopCh[key]); err != nil {
				clusterScope.Error(err, "failed to start a targetClusterManager")
			} else {
				clusterScope.Info("stoppend targetClusterManager")
			}
			r.targetClusterManagersLock.Lock()
			defer r.targetClusterManagersLock.Unlock()
			delete(r.targetClusterManagersStopCh, key)
		}()
	} else if ok && deleted {
		close(stopCh)
		delete(r.targetClusterManagersStopCh, key)
	}

	return nil
}

func (r *HcloudClusterReconciler) newTargetClusterManager(clusterScope *scope.ClusterScope) (ctrl.Manager, error) {
	hcloudCluster := clusterScope.HcloudCluster

	clientConfig, err := clusterScope.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get a clientConfig for the API of HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get a restConfig for the API of HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	clientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get a clientSet for the API of HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
	}

	scheme := runtime.NewScheme()
	_ = certificatesv1.AddToScheme(scheme)

	clusterMgr, err := ctrl.NewManager(
		restConfig,
		ctrl.Options{
			Scheme:             scheme,
			MetricsBindAddress: "0",
			LeaderElection:     false,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to setup guest cluster manager")
	}

	gr := &GuestCSRReconciler{
		Client: clusterMgr.GetClient(),
		Log:    r.Log,
		mCluster: &managementCluster{
			Client:        r.Client,
			hcloudCluster: hcloudCluster,
			recorder:      r.recorder,
		},
		clientSet: clientSet,
	}

	if err := gr.SetupWithManager(clusterMgr, controller.Options{}); err != nil {
		errors.Wrapf(err, "failed to setup CSR controller")
	}

	return clusterMgr, nil
}

var _ ManagementCluster = &managementCluster{}

type managementCluster struct {
	controllerclient.Client
	hcloudCluster *infrav1.HcloudCluster
	recorder      recorder.EventRecorder
}

func (c *managementCluster) Namespace() string {
	return c.hcloudCluster.Namespace
}

func (c *managementCluster) Event(eventtype, reason, message string) {
	c.recorder.Event(c.hcloudCluster, eventtype, reason, message)
}

func (c *managementCluster) Eventf(eventtype, reason, message string, args ...interface{}) {
	c.recorder.Eventf(c.hcloudCluster, eventtype, reason, message, args...)
}

func (r *HcloudClusterReconciler) reconcileManifestsNetwork(clusterScope *scope.ClusterScope) error {
	hcloudCluster := clusterScope.HcloudCluster
	manifests := hcloudCluster.Spec.CNI

	// if nothing is set default to calico
	if manifests == nil {
		hcloudCluster.Spec.CNI = &infrav1.HcloudClusterSpecCNIManifests{
			Network: &infrav1.HcloudClusterSpecManifestsNetwork{
				Calico: &infrav1.HcloudClusterSpecManifestsNetworkCalico{},
			},
		}
		manifests = hcloudCluster.Spec.CNI
	}

	// if Cilium with IPSec is enabled ensure we have an existing PSK
	if manifests != nil && manifests.Network != nil && manifests.Network.Cilium != nil {

		// if IPSec is enabled ensure key ref is set
		cilium := manifests.Network.Cilium
		if cilium.IPSecKeysRef != nil {
			if cilium.IPSecKeysRef.Name == "" {
				cilium.IPSecKeysRef.Name = fmt.Sprintf("%s-cilium-ipsec-keys", hcloudCluster.Name)
			}
			if cilium.IPSecKeysRef.Key == "" {
				cilium.IPSecKeysRef.Key = "keys"
			}
		}
	}

	return nil
}

func (r *HcloudClusterReconciler) reconcileManifests(clusterScope *scope.ClusterScope) error {
	hcloudCluster := clusterScope.HcloudCluster

	if err := r.reconcileManifestsNetwork(clusterScope); err != nil {
		return err
	}

	// Check if manifests need to be applied or reapplied
	expectedHash, err := clusterScope.ManifestsHash()
	if err != nil {
		return err
	}

	applyManifests := func() error {
		machines, hcloudMachines, err := clusterScope.ListMachines(clusterScope.Ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to list machines for HcloudCluster %s/%s", hcloudCluster.Namespace, hcloudCluster.Name)
		}

		var clientConfig clientcmd.ClientConfig
		var readyErrors []error
	machines:
		for pos, m := range machines {
			if !util.IsControlPlaneMachine(m) {
				continue
			}

			if !m.Status.InfrastructureReady {
				continue
			}

			// find a ready clientconfig
			for _, address := range hcloudMachines[pos].Status.Addresses {
				if address.Type != corev1.NodeExternalIP && address.Type != corev1.NodeExternalDNS {
					continue
				}

				c, err := clusterScope.ClientConfigWithAPIEndpoint(clusterv1.APIEndpoint{
					Host: address.Address,
					Port: clusterScope.ControlPlaneAPIEndpointPort(),
				})
				if err != nil {
					return err
				}

				if err := scope.IsControlPlaneReady(clusterScope.Ctx, c); err != nil {
					readyErrors = append(readyErrors, fmt.Errorf("APIserver '%s': %w", hcloudMachines[pos].Name, err))
				}

				// APIserver ready
				clientConfig = c
				break machines
			}
		}

		if clientConfig == nil {
			if err := errorutil.NewAggregate(readyErrors); err != nil {
				r.recorder.Eventf(
					hcloudCluster,
					corev1.EventTypeWarning,
					"APIServerNotReady",
					"Health check for API servers failed: %s",
					err,
				)
				return errNoReadyAPIServer
			}
			return errNoReadyAPIServer
		}

		if err := clusterScope.ApplyManifestsWithClientConfig(clusterScope.Ctx, clientConfig); err != nil {
			return errors.Wrap(err, "error applying manifests to first API server")
		}

		if hcloudCluster.Status.Manifests == nil {
			hcloudCluster.Status.Manifests = &infrav1.HcloudClusterStatusManifests{}
		}
		var myTrue = true
		hcloudCluster.Status.Manifests.Initialized = &myTrue
		hcloudCluster.Status.Manifests.AppliedHash = &expectedHash
		return nil
	}

	m := hcloudCluster.Status.Manifests
	if m == nil ||
		m.Initialized == nil ||
		*m.Initialized == false ||
		m.AppliedHash == nil {
		if err := applyManifests(); err != nil {
			return err
		}
		r.recorder.Eventf(
			hcloudCluster,
			corev1.EventTypeNormal,
			"ManifestsApplied",
			"Latest Manifests (hash=%s) have been successfully applied to initialize the cluster",
			expectedHash,
		)
	} else if expectedHash != *m.AppliedHash {
		if err := applyManifests(); err != nil {
			return err
		}
		r.recorder.Eventf(
			hcloudCluster,
			corev1.EventTypeNormal,
			"ManifestsApplied",
			"Latest Manifests (hash=%s) have been successfully applied to update the exising (hash=%s)",
			expectedHash,
			*m.AppliedHash,
		)
	}

	return nil
}

func (r *HcloudClusterReconciler) SetupWithManager(mgr ctrl.Manager, options controller.Options) error {
	r.targetClusterManagersLock.Lock()
	defer r.targetClusterManagersLock.Unlock()
	if r.targetClusterManagersStopCh == nil {
		r.targetClusterManagersStopCh = make(map[types.NamespacedName]chan struct{})
	}
	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor("hcloud-cluster-reconciler")
	}
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
