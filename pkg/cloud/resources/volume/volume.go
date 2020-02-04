package volume

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	infrav1 "github.com/simonswine/cluster-api-provider-hetzner/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hetzner/pkg/cloud/utils"
)

type Service struct {
	scope *scope.VolumeScope
}

var minumumSize = resource.MustParse("10Gi")

func NewService(scope *scope.VolumeScope) *Service {
	return &Service{
		scope: scope,
	}
}

var errNotImplemented = errors.New("Not implemented")

func apiToStatus(v *hcloud.Volume) *infrav1.HetznerVolumeStatus {
	volumeID := infrav1.HetznerVolumeID(v.ID)
	return &infrav1.HetznerVolumeStatus{
		Location: infrav1.HetznerLocation(v.Location.Name),
		VolumeID: &volumeID,
		Size:     resource.NewQuantity(int64(v.Size)*1024*1024*1024, resource.BinarySI),
	}
}

func (s *Service) name() string {
	return fmt.Sprintf("%s-%s", s.scope.HetznerCluster.Name, s.scope.HetznerVolume.Name)
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HetznerCluster.Name): string(infrav1.ResourceLifecycleOwned),
	}
}

func (s *Service) Reconcile(ctx context.Context) (err error) {

	// ensure requsested volume size is set (default to 10GiB)
	// TODO: should be done through defaulting
	if s.scope.HetznerVolume.Spec.Size == nil {
		s.scope.HetznerVolume.Spec.Size = minumumSize.Copy()
	}

	// ensure requested size is bigger or equal than 10Gi
	// TODO: should be validation
	if s.scope.HetznerVolume.Spec.Size.Cmp(minumumSize) < 1 {
		s.scope.HetznerVolume.Spec.Size = minumumSize.Copy()
	}

	// ensure retain policy is by default to retain
	// TODO: should be done through validation and defaulting
	if s.scope.HetznerVolume.Spec.ReclaimPolicy == "" {
		s.scope.HetznerVolume.Spec.ReclaimPolicy = infrav1.HetznerVolumeReclaimRetain
	}

	// update actual volume status
	actualStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh volume")
	}

	if actualStatus == nil {
		if err := s.createVolume(ctx); err != nil {
			return errors.Wrap(err, "failed to create volume")
		}
	} else {
		s.scope.HetznerVolume.Status = *actualStatus
	}
	return nil
}

func (s *Service) Delete(ctx context.Context) (err error) {
	// update actual volume status
	actualStatus, err := s.actualStatus(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to refresh volume")
	}

	if actualStatus == nil {
		return nil
	}

	if s.scope.HetznerVolume.Spec.ReclaimPolicy != infrav1.HetznerVolumeReclaimDelete {
		s.scope.V(1).Info("Remove kubernetes volume, but retain HetznerVolume due to ReclaimPolicy", "volume_id", actualStatus.VolumeID)
		return nil
	}

	return s.deleteVolume(ctx, *actualStatus.VolumeID)
}

func (s *Service) createVolume(ctx context.Context) error {
	var automount = false
	var format = "ext4"
	opts := hcloud.VolumeCreateOpts{
		Name:      s.name(),
		Labels:    s.labels(),
		Location:  &hcloud.Location{Name: string(s.scope.HetznerVolume.Spec.Location)},
		Format:    &format,
		Automount: &automount,
		Size:      int(s.scope.HetznerVolume.Spec.Size.Value() / 1024 / 1024 / 1024),
	}

	v, _, err := s.scope.HetznerClient().CreateVolume(ctx, opts)
	if err != nil {
		return err
	}
	s.scope.HetznerVolume.Status = *apiToStatus(v.Volume)

	return nil
}

func (s *Service) deleteVolume(ctx context.Context, id infrav1.HetznerVolumeID) error {
	_, err := s.scope.HetznerClient().DeleteVolume(ctx, &hcloud.Volume{ID: int(id)})
	return err
}

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) (*infrav1.HetznerVolumeStatus, error) {
	opts := hcloud.VolumeListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.labels())
	volumes, err := s.scope.HetznerClient().ListVolumes(s.scope.Ctx, opts)
	if err != nil {
		return nil, err
	}

	for _, v := range volumes {
		if v.Name == s.name() {
			return apiToStatus(v), nil
		}
	}

	return nil, nil

}
