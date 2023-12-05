package aspecteventresource

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// AspectEventResourceRepository is responsible for the repo-layer Aspect Event Resource operations.
//
//go:generate mockery --name=AspectEventResourceRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectEventResourceRepository interface {
	Create(ctx context.Context, tenant string, item *model.AspectEventResource) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal Aspect Event Resource IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo         AspectEventResourceRepository
	uidService   UIDService
	timestampGen timestamp.Generator
}

// NewService returns a new object responsible for service-layer Aspect Event Resource operations.
func NewService(repo AspectEventResourceRepository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// Create creates an Aspect Event Resource for an Aspect with given ID.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, aspectID string, in model.AspectEventResourceInput) (string, error) {
	id := s.uidService.Generate()
	aspectEventResource := in.ToAspectEventResource(id, resourceType, resourceID, aspectID)

	if err := s.createAspectEventResource(ctx, aspectEventResource); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Aspect Event Resource with id %s for an Aspect with id %s", id, aspectID)
	}

	log.C(ctx).Infof("Successfully created an Aspect Event Resource with id %s for an Aspect with id %s", id, aspectID)

	return id, nil
}

func (s *service) createAspectEventResource(ctx context.Context, aspectEventResource *model.AspectEventResource) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, aspectEventResource)
}
