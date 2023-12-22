package dataproduct

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// DataProductRepository is responsible for the repo-layer Data Product operations.
//
//go:generate mockery --name=DataProductRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataProductRepository interface {
	ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.DataProduct, error)
	Create(ctx context.Context, tenant string, item *model.DataProduct) error
	CreateGlobal(ctx context.Context, item *model.DataProduct) error
	GetByID(ctx context.Context, tenantID, id string) (*model.DataProduct, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.DataProduct, error)
	Update(ctx context.Context, tenant string, item *model.DataProduct) error
	UpdateGlobal(ctx context.Context, item *model.DataProduct) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteGlobal(ctx context.Context, id string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal data product IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo         DataProductRepository
	uidService   UIDService
	timestampGen timestamp.Generator
}

// NewService returns a new object responsible for service-layer Data Product operations.
func NewService(repo DataProductRepository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// ListByApplicationID lists all Data Products for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.DataProduct, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByResourceID(ctx, tnt, resource.Application, appID)
}

// ListByApplicationTemplateVersionID lists all Data Products for a given application template version ID.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.DataProduct, error) {
	return s.repo.ListByResourceID(ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID)
}

// Create creates Data Product for a resource with given id.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.DataProductInput, dataProductHash uint64) (string, error) {
	id := s.uidService.Generate()
	dataProduct := in.ToDataProduct(id, resourceType, resourceID, packageID, dataProductHash)

	if err := s.createDataProduct(ctx, resourceType, dataProduct); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Data Product with id %s for %s with id %s", id, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created a Data Product with id %s for %s with id %s", id, resourceType, resourceID)

	return id, nil
}

// Update updates an existing Data Product.
func (s *service) Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, packageID *string, in model.DataProductInput, dataProductHash uint64) error {
	dataProduct := in.ToDataProduct(id, resourceType, resourceID, packageID, dataProductHash)

	err := s.updateDataProduct(ctx, dataProduct, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while updating Data Product with ID %s for %s", id, resourceType)
	}

	return nil
}

// Get returns a Data Product by given ID.
func (s *service) Get(ctx context.Context, id string) (*model.DataProduct, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	dataProduct, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Data Product with ID: %q", id)
	}

	return dataProduct, nil
}

// Delete deletes the Data Product by its ID.
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteDataProduct(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Data Product with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Data Product with id %s", id)

	return nil
}

func (s *service) createDataProduct(ctx context.Context, resourceType resource.Type, dataProduct *model.DataProduct) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.CreateGlobal(ctx, dataProduct)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, dataProduct)
}

func (s *service) getDataProduct(ctx context.Context, id string, resourceType resource.Type) (*model.DataProduct, error) {
	if resourceType.IsTenantIgnorable() {
		return s.repo.GetByIDGlobal(ctx, id)
	}

	return s.Get(ctx, id)
}

func (s *service) updateDataProduct(ctx context.Context, dataProduct *model.DataProduct, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.UpdateGlobal(ctx, dataProduct)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, tnt, dataProduct)
}

func (s *service) deleteDataProduct(ctx context.Context, dataProductID string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.DeleteGlobal(ctx, dataProductID)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, tnt, dataProductID)
}
