package bastionhost

import (
	"context"
	"strings"
	"time"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/cloud/utils"
	packerapi "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Service struct {
	scope *scope.BastionHostScope
}

// Default location of bastion host is Falkenstein
const defaultLocation = "fsn1"

func NewService(scope *scope.BastionHostScope) *Service {
	return &Service{
		scope: scope,
	}
}

func (s *Service) createLabels() map[string]string {

	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HcloudCluster.Name): string(infrav1.ResourceLifecycleOwned),
		"machine_type": "bastion_host",
	}
}

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {

	// detect failure domain
	if s.scope.BastionHost.Spec.Location != nil {
		s.scope.BastionHost.Status.Location = *s.scope.BastionHost.Spec.Location
	} else {
		s.scope.BastionHost.Status.Location = defaultLocation
	}

	// gather image ID
	version := s.scope.BastionHost.Spec.Version
	imageID, err := s.scope.EnsureImage(ctx, &packerapi.PackerParameters{
		KubernetesVersion: strings.Trim(*version, "v"),
		Image:             s.scope.BastionHost.Spec.ImageName,
	})
	if err != nil {
		return nil, err
	}

	if imageID == nil {
		return &ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	s.scope.BastionHost.Status.ImageID = imageID

	var myTrue = true
	var myFalse = false

	opts := hcloud.ServerCreateOpts{
		Name:   s.scope.BastionHost.Name,
		Labels: s.createLabels(),
		Image: &hcloud.Image{
			ID: int(*s.scope.BastionHost.Status.ImageID),
		},
		Location: &hcloud.Location{
			Name: s.scope.BastionHost.Status.Location,
		},
		ServerType: &hcloud.ServerType{
			Name: *s.scope.BastionHost.Spec.ServerType,
		},
		Automount:        &myFalse,
		StartAfterCreate: &myTrue,
	}

	// setup SSH keys
	sshKeySpecs := s.scope.BastionHost.Spec.SSHKeys
	if len(sshKeySpecs) == 0 {
		sshKeySpecs = s.scope.HcloudCluster.Spec.SSHKeys
	}
	sshKeys, _, err := s.scope.HcloudClient().ListSSHKeys(ctx, hcloud.SSHKeyListOpts{})
	if err != nil {
	}
	for _, sshKey := range sshKeys {
		var match bool
		for _, sshKeySpec := range sshKeySpecs {
			if sshKeySpec.Name != nil && *sshKeySpec.Name == sshKey.Name {
				match = true
			}
			if sshKeySpec.ID != nil && *sshKeySpec.ID == sshKey.ID {
				match = true
			}
		}
		if match {
			opts.SSHKeys = append(opts.SSHKeys, sshKey)
		}
	}

	// setup network if available
	if net := s.scope.HcloudCluster.Status.Network; net != nil {
		opts.Networks = []*hcloud.Network{{
			ID: net.ID,
		}}
	}

	// update current server
	actualServers, err := s.actualStatus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to refresh bastion host status")
	}

	var bastionHost *hcloud.Server

	if len(actualServers) == 0 {

		if res, _, err := s.scope.HcloudClient().CreateServer(s.scope.Ctx, opts); err != nil {
			return nil, errors.Wrap(err, "failed to create bastion host")
		} else {
			record.Eventf(
				s.scope.BastionHost,
				"SuccessfulCreate",
				"Created new bastion host with id %d",
				res.Server.ID,
			)
			bastionHost = res.Server
		}
	} else if len(actualServers) == 1 {
		bastionHost = actualServers[0]
	} else {
		return nil, errors.New("found more than one actual bastion host")
	}

	// wait for server being running
	if bastionHost.Status != hcloud.ServerStatusRunning {
		s.scope.V(1).Info("bastion host not in running state", "bastion host", bastionHost.Name, "status", bastionHost.Status)
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, nil
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

	// shutdown bastion host
	for _, server := range actionShutdown {
		if _, _, err := s.scope.HcloudClient().ShutdownServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to shutdown bastion host")
		}
		actionWait = append(actionWait, server)
	}

	// delete servers that need delete
	for _, server := range actionDelete {
		if _, err := s.scope.HcloudClient().DeleteServer(ctx, server); err != nil {
			return nil, errors.Wrap(err, "failed to delete bastion host")
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

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) ([]*hcloud.Server, error) {
	opts := hcloud.ServerListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.createLabels())
	servers, err := s.scope.HcloudClient().ListServers(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
