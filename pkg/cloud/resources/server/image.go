package server

import (
	"context"
	"fmt"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/pkg/errors"

	infrav1 "github.com/simonswine/cluster-api-provider-hcloud/api/v1alpha3"
)

func (s *Service) findImageIDBySpec(ctx context.Context, spec *infrav1.HcloudImageSpec) (*infrav1.HcloudImageID, error) {
	if spec == nil {
		return nil, errors.New("no image specified")
	}

	// TODO filter with label selector
	images, err := s.scope.HcloudClient().ListImages(ctx, hcloud.ImageListOpts{})
	if err != nil {
		return nil, fmt.Errorf("error listing images: %w", err)
	}

	for _, image := range images {
		imageID := infrav1.HcloudImageID(image.ID)

		// match by ID
		if specID := spec.ID; specID != nil && *specID == imageID {
			return &imageID, nil
		}

		// match by name
		if specName := spec.Name; specName != nil && *specName == image.Name {
			return &imageID, nil
		}

		// match by description
		if specName := spec.Name; specName != nil && *specName == image.Description {
			return &imageID, nil
		}

	}

	return nil, errors.New("no matching image found")
}
