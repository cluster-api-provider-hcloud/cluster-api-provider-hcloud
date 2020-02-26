package server

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/storage/names"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/utils"
)

type Service struct {
	scope *scope.MachineScope
}

func NewService(scope *scope.MachineScope) *Service {
	return &Service{
		scope: scope,
	}
}

var errNotImplemented = errors.New("Not implemented")

func (s *Service) genericLabels() map[string]string {
	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HetznerCluster.Name): string(infrav1.ResourceLifecycleOwned),
	}
}

func (s *Service) labels() map[string]string {
	m := s.genericLabels()
	m[infrav1.MachineNameTagKey] = s.scope.HetznerMachine.Name
	return m
}

func (s *Service) reconcileKubeadmConfig(ctx context.Context, volumes []*hcloud.Volume) (_ *ctrl.Result, err error) {

	// if kubeadmConfig is not ready yet
	if s.scope.KubeadmConfig == nil {
		return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	k := s.newKubeadmConfig(s.scope.KubeadmConfig)
	k.addKubeletConfigTLSBootstrap()

	// adding volumes
	for pos, volume := range volumes {
		if pos > 1 {
			return nil, fmt.Errorf("found more than a single, which is not expected volumes: %+#v", volumes)
		}
		mountPath := "/var/lib/etcd"
		k.addVolumeMount(int64(volume.ID), mountPath)
		k.addWaitForMount("kubelet.service.d/90-wait-for-mount.conf", mountPath)
	}

	// check if config was just updated
	if resourceVersionUpdated, err := k.update(ctx); err != nil {
		return nil, err
	} else if resourceVersionUpdated != nil {
		s.scope.HetznerMachine.Status.KubeadmConfigResourceVersionUpdated = resourceVersionUpdated
		return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	// ensure it resource version is bigger than at the time it had been last
	// updated
	if rvUpdatedStr :=
		s.scope.HetznerMachine.Status.KubeadmConfigResourceVersionUpdated; rvUpdatedStr != nil {
		rvUpdated, err := strconv.ParseInt(*rvUpdatedStr, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "error converting resourceVersionUpdated to int")
		}

		rvObserved, err := strconv.ParseInt(s.scope.KubeadmConfig.ResourceVersion, 10, 64)
		if err != nil {
			return nil, errors.Wrap(err, "error converting resourceVersionUpdated to int")
		}

		if rvObserved <= rvUpdated {
			k.s.scope.Info("observed resourceVersion of KubeadmConfig not bigger than resourceVersion when last updated")
			return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
	}

	// TODO: ensure user data is not nil

	return nil, errors.New("not implemented")

}

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {
	// gather image ID
	imageID, err := s.findImageIDBySpec(s.scope.Ctx, s.scope.HetznerMachine.Spec.Image)
	if err != nil {
		return nil, err
	}
	s.scope.HetznerMachine.Status.ImageID = imageID

	// gather volumes
	volumes, err := s.volumes(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find volumes for server")
	}

	// reconcile kubeadmConfig
	if res, err := s.reconcileKubeadmConfig(ctx, volumes); err != nil || res != nil {
		return res, err
	}

	// update current server
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	if len(actualServers) == 0 {
		return nil, nil
	}

	if s.scope.Machine.Spec.Bootstrap.Data == nil {
		s.scope.V(1).Info("user-data not available yet")
		return &reconcile.Result{Requeue: true}, nil
	}

	var myTrue = true
	var myFalse = false
	opts := hcloud.ServerCreateOpts{
		Name: names.SimpleNameGenerator.GenerateName(fmt.Sprintf(
			"%s-%s-",
			s.scope.HetznerCluster.Name,
			s.scope.Machine.Name,
		)),
		Labels: s.labels(),
		Image: &hcloud.Image{
			ID: int(*s.scope.HetznerMachine.Status.ImageID),
		},
		Location: &hcloud.Location{
			Name: string(s.scope.HetznerMachine.Status.Location),
		},
		SSHKeys: []*hcloud.SSHKey{
			{
				ID: 91895,
			},
		},
		ServerType: &hcloud.ServerType{
			Name: string(s.scope.HetznerMachine.Spec.Type),
		},
		Automount:        &myFalse,
		StartAfterCreate: &myTrue,
	}

	// setup network if available
	if net := s.scope.HetznerCluster.Status.Network; net != nil {
		opts.Networks = []*hcloud.Network{{
			ID: net.ID,
		}}
	}

	if _, _, err := s.scope.HetznerClient().CreateServer(s.scope.Ctx, opts); err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	return nil, nil
}

func (s *Service) Delete(ctx context.Context) (_ *ctrl.Result, err error) {
	// update current servers
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	var actionWait []*hcloud.Server
	var actionShutdown []*hcloud.Server
	var actionDelete []*hcloud.Server

	for _, server := range actualServers {
		switch status := server.Status; status {
		case hcloud.ServerStatusRunning:
			actionShutdown = append(actionShutdown, server)
		case hcloud.ServerStatusOff:
			actionDelete = append(actionDelete, server)
		default:
			actionWait = append(actionWait, server)
		}
	}

	// shutdown servers
	for _, server := range actionShutdown {
		if _, _, err := s.scope.HetznerClient().ShutdownServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to shutdown server")
		}
		actionWait = append(actionWait, server)
	}

	// delete servers that need delete
	for _, server := range actionDelete {
		if _, err := s.scope.HetznerClient().DeleteServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to delete server")
		}
	}

	var result *ctrl.Result
	if len(actionWait) > 0 {
		result = &ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}
	}

	return result, nil
}

func (s *Service) volumes(ctx context.Context) ([]*hcloud.Volume, error) {
	opts := hcloud.VolumeListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.genericLabels())
	volumes, err := s.scope.HetznerClient().ListVolumes(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	var volumesSelected []*hcloud.Volume
	for _, v := range volumes {
		if v.Name == fmt.Sprintf("%s-%s", s.scope.HetznerCluster.Name, s.scope.HetznerMachine.Name) {
			volumesSelected = append(volumesSelected, v)
		}
	}

	return volumesSelected, nil
}

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) ([]*hcloud.Server, error) {
	opts := hcloud.ServerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.labels())
	servers, err := s.scope.HetznerClient().ListServers(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil

}
