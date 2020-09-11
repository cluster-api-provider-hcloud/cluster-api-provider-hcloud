package server

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	infrav1 "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/api/v1alpha3"
	packerapi "github.com/cluster-api-provider-hcloud/cluster-api-provider-hcloud/pkg/packer/api"
)

func (s *Service) findImageIDBySpec(ctx context.Context, spec string) (*infrav1.HcloudImageID, error) {
	if spec == "" {
		version := s.scope.Machine.Spec.Version
		if version == nil {
			return nil, errors.New("Machine has no Version set")
		}
		return s.scope.EnsureImage(ctx, &packerapi.PackerParameters{
			KubernetesVersion: strings.Trim(*version, "v"),
			Image:             s.scope.HcloudMachine.Spec.ImageName,
		})
	}
	/*
		// specific image specified
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
	*/
	return nil, errors.New("no matching image found")
}
