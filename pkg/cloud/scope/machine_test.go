package scope

import (
	"testing"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
)

func newFakeMachineScope() *MachineScope {
	clusterScope := newFakeClusterScope()
	return &MachineScope{
		Machine:       &clusterv1.Machine{},
		HcloudMachine: &infrav1.HcloudMachine{},
		ClusterScope:  *clusterScope,
	}
}

func newFakeClusterScope() *ClusterScope {
	return &ClusterScope{
		Cluster:       &clusterv1.Cluster{},
		HcloudCluster: &infrav1.HcloudCluster{},
	}
}

func TestMachineScope_GetFailureDomain(t *testing.T) {
	location := "loc1"
	s := newFakeMachineScope()
	s.Machine.Spec.FailureDomain = &location

	// base case failureDomain on the machine
	l, err := s.GetFailureDomain()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if exp, act := location, l; exp != act {
		t.Errorf("unexpected location: %s (expected %s)", act, exp)
	}

	// error without failureDomain on cluster
	s = newFakeMachineScope()
	if _, err := s.GetFailureDomain(); err != ErrFailureDomainNotFound {
		t.Errorf("unexpected error %s (expected %s)", err, ErrBootstrapDataNotReady)
	}

	// expect a consistently hashed name
	s = newFakeMachineScope()
	s.HcloudMachine.ObjectMeta.Name = "my-name"
	s.Cluster.Status.FailureDomains = map[string]clusterv1.FailureDomainSpec{
		"loc2": {},
		"loc3": {},
		"loc4": {},
	}
	l, err = s.GetFailureDomain()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if exp, act := "loc3", l; exp != act {
		t.Errorf("unexpected location: %s (expected %s)", act, exp)
	}
	s.Machine.ObjectMeta.Labels = map[string]string{
		clusterv1.MachineControlPlaneLabelName: "true",
	}
	if _, err := s.GetFailureDomain(); err != ErrFailureDomainNotFound {
		t.Errorf("unexpected error %s (expected %s)", err, ErrBootstrapDataNotReady)
	}

	// expect a consistently hashed name with control plane filter
	s = newFakeMachineScope()
	s.Machine.ObjectMeta.Labels = map[string]string{
		clusterv1.MachineControlPlaneLabelName: "true",
	}
	s.Cluster.Status.FailureDomains = map[string]clusterv1.FailureDomainSpec{
		"loc2": {},
		"loc3": {},
		"loc4": {ControlPlane: true},
	}
	l, err = s.GetFailureDomain()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if exp, act := "loc4", l; exp != act {
		t.Errorf("unexpected location: %s (expected %s)", act, exp)
	}
}
