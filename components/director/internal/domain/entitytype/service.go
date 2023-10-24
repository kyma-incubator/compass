package entitytype

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// EntityTypeRepository missing godoc
//
//go:generate mockery --name=EntityTypeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeRepository interface {
	Create(ctx context.Context, tenant string, item *model.EntityType) error
	CreateGlobal(ctx context.Context, model *model.EntityType) error
	Update(ctx context.Context, tenant string, item *model.EntityType) error
	UpdateGlobal(ctx context.Context, model *model.EntityType) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.EntityType, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.EntityType, error)
	ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.EntityType, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	entityTypeRepo EntityTypeRepository
	uidService     UIDService
}

// NewService missing godoc
func NewService(entityTypeRepo EntityTypeRepository, uidService UIDService) *service {
	return &service{
		entityTypeRepo: entityTypeRepo,
		uidService:     uidService,
	}
}

// Create creates an Entity Type for a given resource.Type
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID string, in model.EntityTypeInput, entityTypeHash uint64) (string, error) {
	id := s.uidService.Generate()
	entityType := in.ToEntityType(id, resourceType, resourceID, packageID, entityTypeHash)

	if err := s.createEntityType(ctx, entityType, resourceType); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Entity Type with id %s and title %s for %s with id %s", id, entityType.Title, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created an Entity Type with id %s and title %s for %s with id %s", id, entityType.Title, resourceType, resourceID)

	return id, nil
}

// Update updates an Entity Type by ID for a given resource.Type
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.EntityTypeInput, entityTypeHash uint64) error {
	entityType, err := s.getEntityType(ctx, id, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while getting Entity Type with id %s", id)
	}

	entityType.SetFromUpdateInput(in, entityTypeHash)

	if err = s.updateEntityType(ctx, entityType, resourceType); err != nil {
		return errors.Wrapf(err, "while updating Entity Type with id %s", id)
	}

	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteEntityType(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Entity Type with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Entity Type with id %s", id)

	return nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.entityTypeRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Entity Type with ID: %q", id)
	}

	return exist, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.EntityType, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	entityType, err := s.entityTypeRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Entity Type with ID: %q", id)
	}

	return entityType, nil
}

// ListByApplicationID missing godoc
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.EntityType, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.entityTypeRepo.ListByResourceID(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionID lists entity types by Application Template Version ID without tenant isolation
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.EntityType, error) {
	return s.entityTypeRepo.ListByResourceID(ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion)
}

func (s *service) createEntityType(ctx context.Context, entityType *model.EntityType, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeRepo.CreateGlobal(ctx, entityType)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.entityTypeRepo.Create(ctx, tnt, entityType)
}

func (s *service) getEntityType(ctx context.Context, id string, resourceType resource.Type) (*model.EntityType, error) {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeRepo.GetByIDGlobal(ctx, id)
	}

	return s.Get(ctx, id)
}

func (s *service) updateEntityType(ctx context.Context, entityType *model.EntityType, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeRepo.UpdateGlobal(ctx, entityType)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.entityTypeRepo.Update(ctx, tnt, entityType)
}

func (s *service) deleteEntityType(ctx context.Context, id string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.entityTypeRepo.DeleteGlobal(ctx, id)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	return s.entityTypeRepo.Delete(ctx, tnt, id)
}
