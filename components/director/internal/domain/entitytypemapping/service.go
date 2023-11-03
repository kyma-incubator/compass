package entitytypemapping

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// EntityTypeMappingRepository missing godoc
//
//go:generate mockery --name=EntityTypeMappingRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeMappingRepository interface {
	Create(ctx context.Context, tenant string, item *model.EntityTypeMapping) error
	CreateGlobal(ctx context.Context, model *model.EntityTypeMapping) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
	GetByID(ctx context.Context, tenant, id string) (*model.EntityTypeMapping, error)
	ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.EntityTypeMapping, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	entityTypeMappingRepo EntityTypeMappingRepository
	uidService            UIDService
}

// NewService returns a new service instance
func NewService(entityTypeMappingRepo EntityTypeMappingRepository, uidService UIDService) *service {
	return &service{
		entityTypeMappingRepo: entityTypeMappingRepo,
		uidService:            uidService,
	}
}

// Create creates an Entity Type Mapping for a given resource.Type
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in *model.EntityTypeMappingInput) (string, error) {
	id := s.uidService.Generate()
	entityTypeMapping := in.ToEntityTypeMapping(id, resourceType, resourceID)

	if err := s.createEntityTypeMapping(ctx, entityTypeMapping, resourceType); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Entity Type Mapping with id %s for %s with id %s", id, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created an Entity Type Mapping with id %s for %s with id %s", id, resourceType, resourceID)

	return id, nil
}

// Delete deletes an Entity Type Mapping by ID
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteEntityTypeMapping(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Entity Type Mapping with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Entity Type Mapping with id %s", id)

	return nil
}

// Get returns an Entity Type Mapping by ID
func (s *service) Get(ctx context.Context, id string) (*model.EntityTypeMapping, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	entityTypeMapping, err := s.entityTypeMappingRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Entity Type Mapping with ID: %q", id)
	}

	return entityTypeMapping, nil
}

func (s *service) ListByOwnerResourceID(ctx context.Context, resourceID string, resourceType resource.Type) ([]*model.EntityTypeMapping, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.entityTypeMappingRepo.ListByResourceID(ctx, tnt, resourceID, resourceType)
}

func (s *service) createEntityTypeMapping(ctx context.Context, entityTypeMapping *model.EntityTypeMapping, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeMappingRepo.CreateGlobal(ctx, entityTypeMapping)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.entityTypeMappingRepo.Create(ctx, tnt, entityTypeMapping)
}

func (s *service) deleteEntityTypeMapping(ctx context.Context, id string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeMappingRepo.DeleteGlobal(ctx, id)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	return s.entityTypeMappingRepo.Delete(ctx, tnt, id)
}
