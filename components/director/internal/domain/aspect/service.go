package aspect

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// AspectRepository is responsible for the repo-layer Aspect operations.
//
//go:generate mockery --name=AspectRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectRepository interface {
	Create(ctx context.Context, tenant string, item *model.Aspect) error
	DeleteByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyID string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal Aspect IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo         AspectRepository
	uidService   UIDService
	timestampGen timestamp.Generator
}

// NewService returns a new object responsible for service-layer Aspect operations.
func NewService(repo AspectRepository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// Create creates an Aspect for an Integration Dependency with given ID.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, in model.AspectInput) (string, error) {
	id := s.uidService.Generate()
	aspect := in.ToAspect(id, resourceType, resourceID, integrationDependencyID)

	if err := s.createAspect(ctx, aspect); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Aspect with id %s for integration dependency with id %s", id, integrationDependencyID)
	}

	log.C(ctx).Debugf("Successfully created a Aspect with id %s for integration dependency with id %s", id, integrationDependencyID)

	return id, nil
}

// DeleteByIntegrationDependencyID deletes Aspects for an Integration Dependency with given ID
func (s *service) DeleteByIntegrationDependencyID(ctx context.Context, integrationDependencyID string) error {
	if err := s.deleteAspectByIntegrationDependencyID(ctx, integrationDependencyID); err != nil {
		return errors.Wrapf(err, "while deleting Aspects for Integration Dependency with id %s", integrationDependencyID)
	}

	log.C(ctx).Infof("Successfully deleted Aspects for Integration Dependency with id %s", integrationDependencyID)

	return nil
}

func (s *service) createAspect(ctx context.Context, aspect *model.Aspect) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, aspect)
}

func (s *service) deleteAspectByIntegrationDependencyID(ctx context.Context, integrationDependencyID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteByIntegrationDependencyID(ctx, tnt, integrationDependencyID)
}
