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
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/go-logr/logr"
	"github.com/hetznercloud/hcloud-go/hcloud"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/storage/names"
	ctrl "sigs.k8s.io/controller-runtime"
	controllerclient "sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha3 "sigs.k8s.io/cluster-api-provider-hetzner/api/v1alpha3"
)

// Clients collects all methods used by the controller in the hetzner cloud API
type Client interface {
	ListLocation(context.Context) ([]*hcloud.Location, error)
	CreateFloatingIP(context.Context, hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error)
	ListFloatingIPs(context.Context, hcloud.FloatingIPListOpts) ([]*hcloud.FloatingIP, error)
	ListImages(context.Context, hcloud.ImageListOpts) ([]*hcloud.Image, error)
}

var _ Client = &realClient{}

type realClient struct {
	client *hcloud.Client
}

func (c *realClient) ListLocation(ctx context.Context) ([]*hcloud.Location, error) {
	return c.client.Location.All(ctx)
}

func (c *realClient) CreateFloatingIP(ctx context.Context, opts hcloud.FloatingIPCreateOpts) (hcloud.FloatingIPCreateResult, *hcloud.Response, error) {
	return c.client.FloatingIP.Create(ctx, opts)
}

func (c *realClient) ListFloatingIPs(ctx context.Context, opts hcloud.FloatingIPListOpts) ([]*hcloud.FloatingIP, error) {
	return c.client.FloatingIP.AllWithOpts(ctx, opts)
}

func (c *realClient) ListImages(ctx context.Context, opts hcloud.ImageListOpts) ([]*hcloud.Image, error) {
	return c.client.Image.AllWithOpts(ctx, opts)
}

type ClientFactory func(token string) Client

// HetznerClusterReconciler reconciles a HetznerCluster object
type HetznerClusterReconciler struct {
	controllerclient.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	clientFactory ClientFactory
}

// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznerclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=hetznerclusters/status,verbs=get;update;patch

func (r *HetznerClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("hetznercluster", req.NamespacedName)

	// retrieve cluster object
	var cluster v1alpha3.HetznerCluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting hetznercluster/%s: %s", req.NamespacedName, err)
	}

	// retrieve token secret
	var tokenSecret corev1.Secret
	tokenSecretName := types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Spec.TokenRef.Name}
	if err := r.Get(ctx, tokenSecretName, &tokenSecret); err != nil {
		return ctrl.Result{}, fmt.Errorf("error getting referenced token secret/%s: %s", tokenSecretName, err)
	}

	tokenBytes, keyExists := tokenSecret.Data[cluster.Spec.TokenRef.Key]
	if !keyExists {
		return ctrl.Result{}, fmt.Errorf("error key %s does not exist in secret/%s", cluster.Spec.TokenRef.Key, tokenSecretName)
	}

	hc := r.clientFactory(string(tokenBytes))

	// run all reconcile tasks in order
	for _, task := range []reconcileTask{
		r.reconcileLocation,
		r.reconcileFloatingIPs,
		r.reconcileImage,
	} {
		if result, err := task(reconcileTaskInput{
			ctx:           ctx,
			hc:            hc,
			cluster:       &cluster,
			clusterBefore: cluster.DeepCopy(),
			log:           log,
		}); err != nil {
			return ctrl.Result{}, err
		} else if result != nil {
			return *result, err
		}
	}

	return ctrl.Result{}, nil
}

type reconcileTaskInput struct {
	ctx           context.Context
	hc            Client
	cluster       *v1alpha3.HetznerCluster
	clusterBefore *v1alpha3.HetznerCluster
	log           logr.Logger
}

func (r *reconcileTaskInput) statusChanged() bool {
	if reflect.DeepEqual(r.clusterBefore.Status, r.cluster.Status) {
		return false
	}
	return true
}

func (r *reconcileTaskInput) updateStatusIfChanged(c controllerclient.Client) (*ctrl.Result, error) {
	if !r.statusChanged() {
		return nil, nil
	}

	return &ctrl.Result{}, c.Status().Update(r.ctx, r.cluster)
}

type reconcileTask func(i reconcileTaskInput) (*ctrl.Result, error)

// reconcileLocation makes sure the cluster location is correct
func (r *HetznerClusterReconciler) reconcileLocation(i reconcileTaskInput) (*ctrl.Result, error) {
	locations, err := i.hc.ListLocation(i.ctx)
	if err != nil {
		return nil, err
	}

	for _, location := range locations {
		if location.Name == string(i.cluster.Spec.Location) {
			i.cluster.Status.Location = i.cluster.Spec.Location
			i.cluster.Status.NetworkZone = v1alpha3.HetznerNetworkZone(location.NetworkZone)
			return i.updateStatusIfChanged(r)
		}
	}
	return nil, fmt.Errorf("error location '%s' cannot be found", i.cluster.Spec.Location)
}

func matchFloatingIPSpecStatus(spec v1alpha3.HetznerFloatingIPSpec, status v1alpha3.HetznerFloatingIPStatus) bool {
	return spec.Type == status.Type
}

func floatingIPAPIToStatus(ip *hcloud.FloatingIP) (*v1alpha3.HetznerFloatingIPStatus, error) {
	network := fmt.Sprintf("%s/32", ip.IP.String())
	if ip.Network != nil {
		network = ip.Network.String()
	}

	var ipType v1alpha3.HetznerFloatingIPType

	if ip.Type == hcloud.FloatingIPTypeIPv4 {
		ipType = v1alpha3.HetznerFloatingIPTypeIPv4
	} else if ip.Type == hcloud.FloatingIPTypeIPv6 {
		ipType = v1alpha3.HetznerFloatingIPTypeIPv6
	} else {
		return nil, fmt.Errorf("Unknown floating IP type: %s", ip.Type)
	}

	status := &v1alpha3.HetznerFloatingIPStatus{
		ID:      ip.ID,
		Name:    ip.Name,
		Network: network,
		Type:    ipType,
	}
	return status, nil
}

type intSlice []int

func (s intSlice) contains(e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

// reconcileFloatingIPs ensures the correct base image is available
func (r *HetznerClusterReconciler) reconcileImage(i reconcileTaskInput) (*ctrl.Result, error) {
	if i.cluster.Spec.Image == nil {
		return nil, errors.New("error no image specified")
	}

	images, err := i.hc.ListImages(i.ctx, hcloud.ImageListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing images: %w", err)
	}

	for _, image := range images {
		imageID := fmt.Sprintf("%d", image.ID)

		// match by ID
		if specID := i.cluster.Spec.Image.ID; specID != nil && *specID == imageID {
			i.cluster.Status.ImageID = imageID
			return i.updateStatusIfChanged(r)
		}

		// match by name
		if specName := i.cluster.Spec.Image.Name; specName != nil && *specName == image.Name {
			i.cluster.Status.ImageID = imageID
			return i.updateStatusIfChanged(r)
		}

		// match by description
		if specName := i.cluster.Spec.Image.Name; specName != nil && *specName == image.Description {
			i.cluster.Status.ImageID = imageID
			return i.updateStatusIfChanged(r)
		}

	}

	return nil, errors.New("error no matching image found")

}

// reconcileFloatingIPs makes sure the cluster's floating IPs are existing
func (r *HetznerClusterReconciler) reconcileFloatingIPs(i reconcileTaskInput) (*ctrl.Result, error) {
	// index existing status entries
	var ids intSlice
	ipStatusByID := make(map[int]*v1alpha3.HetznerFloatingIPStatus)
	for pos := range i.cluster.Status.ControlPlaneFloatingIPs {
		ipStatus := &i.cluster.Status.ControlPlaneFloatingIPs[pos]
		ipStatusByID[ipStatus.ID] = ipStatus
		ids = append(ids, ipStatus.ID)
	}
	for _, ipSpec := range i.cluster.Spec.ControlPlaneFloatingIPs {
		if ipSpec.ID != nil {
			ids = append(ids, *ipSpec.ID)
		}
	}

	// refresh existing floating IPs
	clusterTagKey := v1alpha3.ClusterTagKey(i.cluster.Name)
	ipStatuses, err := i.hc.ListFloatingIPs(i.ctx, hcloud.FloatingIPListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing floating IPs: %w", err)
	}
	for _, ipStatus := range ipStatuses {
		_, ok := ipStatus.Labels[clusterTagKey]
		if !ok && !ids.contains(ipStatus.ID) {
			continue
		}

		apiStatus, err := floatingIPAPIToStatus(ipStatus)
		if err != nil {
			return nil, fmt.Errorf("error converting floatingIP API to status: %w", err)
		}
		ipStatusByID[apiStatus.ID] = apiStatus
	}

	ids = []int{}
	for id, _ := range ipStatusByID {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	i.cluster.Status.ControlPlaneFloatingIPs = []v1alpha3.HetznerFloatingIPStatus{}
	for _, id := range ids {
		status := ipStatusByID[id]
		i.cluster.Status.ControlPlaneFloatingIPs = append(
			i.cluster.Status.ControlPlaneFloatingIPs,
			*status,
		)
	}

	// update object if changed
	if i.statusChanged() {
		return i.updateStatusIfChanged(r)
	}

	var matchedSpecToStatusMap = make(map[int]*v1alpha3.HetznerFloatingIPStatus)
	var matchedStatusToSpecMap = make(map[int]*v1alpha3.HetznerFloatingIPSpec)

	for posSpec, ipSpec := range i.cluster.Spec.ControlPlaneFloatingIPs {
		for posStatus, ipStatus := range i.cluster.Status.ControlPlaneFloatingIPs {
			if _, ok := matchedStatusToSpecMap[posStatus]; ok {
				continue
			}

			if matchFloatingIPSpecStatus(ipSpec, ipStatus) {
				matchedStatusToSpecMap[posStatus] = &ipSpec
				matchedSpecToStatusMap[posSpec] = &ipStatus
			}
		}
	}

	// TODO: floating IPs to remove

	// floating IPs to create
	for pos, spec := range i.cluster.Spec.ControlPlaneFloatingIPs {
		if _, ok := matchedSpecToStatusMap[pos]; ok {
			continue
		}

		var ipType hcloud.FloatingIPType
		if spec.Type == v1alpha3.HetznerFloatingIPTypeIPv4 {
			ipType = hcloud.FloatingIPTypeIPv4
		} else if spec.Type == v1alpha3.HetznerFloatingIPTypeIPv6 {
			ipType = hcloud.FloatingIPTypeIPv6
		} else {
			return nil, fmt.Errorf("error invalid floating IP type: %s", spec.Type)
		}

		homeLocation := &hcloud.Location{Name: string(i.cluster.Status.Location)}
		name := names.SimpleNameGenerator.GenerateName(i.cluster.Name + "-control-plane-")
		description := fmt.Sprintf("Kubernetes control plane %s", i.cluster.Name)

		opts := hcloud.FloatingIPCreateOpts{
			Type:         ipType,
			Name:         &name,
			Description:  &description,
			HomeLocation: homeLocation,
			Labels: map[string]string{
				clusterTagKey: string(v1alpha3.ResourceLifecycleOwned),
			},
		}

		ip, _, err := i.hc.CreateFloatingIP(i.ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("error creating floating IP: %s", err)
		}

		status, err := floatingIPAPIToStatus(ip.FloatingIP)
		if err != nil {
			return nil, fmt.Errorf("error converting floating IP: %s", err)
		}

		i.cluster.Status.ControlPlaneFloatingIPs = append(
			i.cluster.Status.ControlPlaneFloatingIPs,
			*status,
		)

		return i.updateStatusIfChanged(r)
	}
	return nil, nil
}

func (r *HetznerClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.clientFactory == nil {
		r.clientFactory = func(token string) Client {
			return &realClient{client: hcloud.NewClient(hcloud.WithToken(token))}
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha3.HetznerCluster{}).
		Complete(r)
}
