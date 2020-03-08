package volume

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/scope"
	"github.com/simonswine/cluster-api-provider-hcloud/pkg/cloud/utils"
)

type Service struct {
	scope *scope.VolumeScope
}

var volumeMinimumSize = resource.MustParse("10Gi")

func minimumSize() *resource.Quantity {
	test := volumeMinimumSize
	return &test
}

func NewService(scope *scope.VolumeScope) *Service {
	return &Service{
		scope: scope,
	}
}

var errNotImplemented = errors.New("Not implemented")

func apiToStatus(v *hcloud.Volume) *infrav1.HcloudVolumeStatus {
	volumeID := infrav1.HcloudVolumeID(v.ID)
	return &infrav1.HcloudVolumeStatus{
		Location: infrav1.HcloudLocation(v.Location.Name),
		VolumeID: &volumeID,
		Size:     resource.NewQuantity(int64(v.Size)*1024*1024*1024, resource.BinarySI),
	}
}

func (s *Service) name() string {
	return fmt.Sprintf("%s-%s", s.scope.HcloudCluster.Name, s.scope.HcloudVolume.Name)
}

func (s *Service) labels() map[string]string {
	return map[string]string{
		infrav1.ClusterTagKey(s.scope.HcloudCluster.Name): string(infrav1.ResourceLifecycleOwned),
	}
}

func (s *Service) Reconcile(ctx context.Context) (err error) {

	// ensure requsested volume size is set (default to 10GiB)
	// TODO: should be done through defaulting
	if s.scope.HcloudVolume.Spec.Size == nil {
		s.scope.HcloudVolume.Spec.Size = minimumSize()
	}

	// ensure requested size is bigger or equal than 10Gi
	// TODO: should be validation
	if s.scope.HcloudVolume.Spec.Size.Cmp(volumeMinimumSize) < 1 {
		s.scope.HcloudVolume.Spec.Size = minimumSize()
	}

	// ensure retain policy is by default to retain
	// TODO: should be done through validation and defaulting
	if s.scope.HcloudVolume.Spec.ReclaimPolicy == "" {
		s.scope.HcloudVolume.Spec.ReclaimPolicy = infrav1.HcloudVolumeReclaimRetain
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
		s.scope.HcloudVolume.Status = *actualStatus
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

	if s.scope.HcloudVolume.Spec.ReclaimPolicy != infrav1.HcloudVolumeReclaimDelete {
		s.scope.V(1).Info("Remove kubernetes volume, but retain HcloudVolume due to ReclaimPolicy", "volume_id", actualStatus.VolumeID)
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
		Location:  &hcloud.Location{Name: string(s.scope.HcloudVolume.Spec.Location)},
		Format:    &format,
		Automount: &automount,
		Size:      int(s.scope.HcloudVolume.Spec.Size.Value() / 1024 / 1024 / 1024),
	}

	v, _, err := s.scope.HcloudClient().CreateVolume(ctx, opts)
	if err != nil {
		return err
	}
	s.scope.HcloudVolume.Status = *apiToStatus(v.Volume)

	return nil
}

func (s *Service) deleteVolume(ctx context.Context, id infrav1.HcloudVolumeID) error {
	_, err := s.scope.HcloudClient().DeleteVolume(ctx, &hcloud.Volume{ID: int(id)})
	return err
}

// actualStatus gathers all matching server instances, matched by tag
func (s *Service) actualStatus(ctx context.Context) (*infrav1.HcloudVolumeStatus, error) {
	opts := hcloud.VolumeListOpts{}
	opts.LabelSelector = utils.LabelsToLabelSelector(s.labels())
	volumes, err := s.scope.HcloudClient().ListVolumes(s.scope.Ctx, opts)
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
