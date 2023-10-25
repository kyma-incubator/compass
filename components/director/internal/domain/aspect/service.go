package aspect

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// AspectRepository is responsible for the repo-layer Aspect operations.
type AspectRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.Aspect, error)
	ListByResourceID(ctx context.Context, tenantID string, integrationDependencyID string) ([]*model.Aspect, error)
	Create(ctx context.Context, tenant string, item *model.Aspect) error
	Delete(ctx context.Context, tenantID string, id string) error
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

// Get returns an Aspect by given ID.
func (s *service) Get(ctx context.Context, id string) (*model.Aspect, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	aspect, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Aspect with ID: %q", id)
	}

	return aspect, nil
}

// ListByIntegrationDependencyID returns all Aspects by given Integration Dependency ID.
func (s *service) ListByIntegrationDependencyID(ctx context.Context, integrationDependencyId string) ([]*model.Aspect, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByResourceID(ctx, tnt, integrationDependencyId)
}

// Create creates aspect for an Integration Dependency with given ID.
func (s *service) Create(ctx context.Context, integrationDependencyId string, in model.AspectInput) (string, error) {
	id := s.uidService.Generate()
	aspect := in.ToAspect(id, integrationDependencyId)

	if err := s.createAspect(ctx, aspect); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Aspect with id %s for integration dependency with id %s", id, integrationDependencyId)
	}

	log.C(ctx).Debugf("Successfully created a Aspect with id %s for integration dependency with id %s", id, integrationDependencyId)

	return id, nil
}

// Delete deletes the Aspect by its ID.
func (s *service) Delete(ctx context.Context, id string) error {
	if err := s.deleteAspect(ctx, id); err != nil {
		return errors.Wrapf(err, "while deleting Aspect with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Aspect with id %s", id)

	return nil
}

// DeleteByIntegrationDependencyID deletes Aspects for an Integration Dependency with given ID
func (s *service) DeleteByIntegrationDependencyID(ctx context.Context, integrationDependencyId string) error {
	if err := s.deleteAspectByIntegrationDependencyID(ctx, integrationDependencyId); err != nil {
		return errors.Wrapf(err, "while deleting Aspects for Integration Dependency with id %s", integrationDependencyId)
	}

	log.C(ctx).Infof("Successfully deleted Aspects for Integration Dependency with id %s", integrationDependencyId)

	return nil
}

func (s *service) createAspect(ctx context.Context, aspect *model.Aspect) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, aspect)
}

func (s *service) deleteAspect(ctx context.Context, aspectId string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, tnt, aspectId)
}

func (s *service) deleteAspectByIntegrationDependencyID(ctx context.Context, integrationDependencyId string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteByIntegrationDependencyID(ctx, tnt, integrationDependencyId)
}
