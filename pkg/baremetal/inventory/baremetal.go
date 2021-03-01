package baremetal

import (
	"context"
	"strings"
	"time"

	"github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/scope"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	hoursBeforeDeletion      time.Duration = 36
	rateLimitTimeOut         time.Duration = 660
	rateLimitTimeOutDeletion time.Duration = 120
)

type Service struct {
	scope *scope.BareMetalMachineScope
}

func NewService(scope *scope.BareMetalMachineScope) *Service {
	return &Service{
		scope: scope,
	}
}

func (s *Service) Reconcile(ctx context.Context) (_ *ctrl.Result, err error) {

	// If not token information has been given, the server cannot be successfully reconciled
	if s.scope.HcloudCluster.Spec.HrobotTokenRef == nil {
		s.scope.Recorder.Eventf(s.scope.BareMetalMachine, corev1.EventTypeWarning, "NoTokenFound", "No Hrobot token found")
		return nil, errors.Errorf("ERROR: No token for Hetzner Robot provided: Cannot reconcile server %s", s.scope.BareMetalMachine.Name)
	}

	// Get bare metal server object and Hetzner status
	status := s.scope.BareMetalMachine.Status
	server, err := s.scope.HrobotClient().GetBMServer(s.scope.IP())
	if err != nil {
		if checkRateLimitExceeded(err) {
			s.scope.Recorder.Eventf(s.scope.BareMetalMachine, corev1.EventTypeWarning, "HrobotRateLimitExceeded", "Hrobot rate limit exceeded. Wait for %v sec before trying again.", rateLimitTimeOut)
			return &reconcile.Result{RequeueAfter: rateLimitTimeOut * time.Second}, nil
		}
		s.scope.Recorder.Eventf(s.scope.BareMetalMachine, corev1.EventTypeWarning, "NoMatchingBareMetalMachineFound", "No matching bare metal machine found")
		return nil, errors.Wrap(err, "Failed to get bare metal server")
	}

	status.HetznerStatus = server.Status
	status.ID = server.ServerNumber
	status.Name = server.ServerName
	status.DataCenter = server.Dc
	status.PaidUntil = server.PaidUntil

	// Get status of object in BareMetalInventory of HcloudCluster.Status, where we update the status
	// in case of a server getting attached to a cluster etc.

	found := false
	bmInventory := s.scope.HcloudCluster.Status.BareMetalInventory
	for _, bmStatus := range bmInventory {
		if bmStatus.ID == status.ID {
			// Get the status (e.g. running, available, etc.) which has been updated in the cluster object
			status.Status = bmStatus.Status
			// Update the rest of the cluster object
			bmStatus = status
			found = true
		}
	}
	if found == false {
		bmInventory = append(bmInventory, status)
	}

	s.scope.HcloudCluster.Status.BareMetalInventory = bmInventory

	return nil, nil
}

func checkRateLimitExceeded(err error) bool {
	return strings.Contains(err.Error(), "rate limit exceeded")
}

func (s *Service) Delete(ctx context.Context) (_ *ctrl.Result, err error) {
	//TODO

	return nil, nil
}
