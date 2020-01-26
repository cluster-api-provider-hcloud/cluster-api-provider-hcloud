package server

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	"k8s.io/apiserver/pkg/storage/names"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1 "sigs.k8s.io/cluster-api-provider-hetzner/api/v1alpha3"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/scope"
	"sigs.k8s.io/cluster-api-provider-hetzner/pkg/cloud/utils"
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

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {
	// gather image ID
	imageID, err := s.findImageIDBySpec(s.scope.Ctx, s.scope.HetznerMachine.Spec.Image)
	if err != nil {
		return nil, err
	}
	s.scope.HetznerMachine.Status.ImageID = imageID

	// update current server
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh server status")
	}

	if len(actualServers) > 0 {
		return nil, nil
	}

	if s.scope.Machine.Spec.Bootstrap.Data == nil {
		s.scope.V(1).Info("user-data not available yet")
		return &reconcile.Result{Requeue: true}, nil
	}

	userDataUpstream, err := base64.StdEncoding.DecodeString(*s.scope.Machine.Spec.Bootstrap.Data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode user data")
	}

	userDataObj, err := s.parseUserData([]byte(userDataUpstream))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse user data")
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

	// setup volumes if available
	if volumes, err := s.volumes(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to find volumes for server")
	} else {
		opts.Volumes = volumes

		if len(volumes) > 0 {
			mountPath := "/var/lib/etcd"
			if err := userDataObj.addVolumeMount(int64(volumes[0].ID), mountPath); err != nil {
				return nil, errors.Wrap(err, "failed to add mount unit")
			}
			if err := userDataObj.addWaitForMount("kubelet.service.d/90-wait-for-mount.conf", mountPath); err != nil {
				return nil, errors.Wrap(err, "failed to add a wait for unit")
			}
		}
	}

	userDataPatched, err := userDataObj.output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode user data")
	}

	opts.UserData = string(userDataPatched)

	if _, _, err := s.scope.HetznerClient().CreateServer(s.scope.Ctx, opts); err != nil {
		return nil, errors.Wrap(err, "failed to create server")
	}

	return nil, nil
}

func (s *Service) Delete(ctx context.Context) (err error) {
	// update current server
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh server status")
	}

	for _, server := range actualServers {
		// TODO: Shutdown first
		if _, err := s.scope.HetznerClient().DeleteServer(ctx, server); err != nil {
			return errors.Wrap(err, "failed to delete server")
		}
	}

	return nil
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
	opts.LabelSelector = utils.LabelsToLabelSelector(s.genericLabels())
	servers, err := s.scope.HetznerClient().ListServers(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil

}
