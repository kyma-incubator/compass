package integrationdependency

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// IntegrationDependencyRepository is responsible for the repo-layer IntegrationDependency operations.
//
//go:generate mockery --name=IntegrationDependencyRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.IntegrationDependency, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.IntegrationDependency, error)
	ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.IntegrationDependency, error)
	Create(ctx context.Context, tenant string, item *model.IntegrationDependency) error
	CreateGlobal(ctx context.Context, item *model.IntegrationDependency) error
	Update(ctx context.Context, tenant string, item *model.IntegrationDependency) error
	UpdateGlobal(ctx context.Context, item *model.IntegrationDependency) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteGlobal(ctx context.Context, id string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal integration dependency IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo         IntegrationDependencyRepository
	uidService   UIDService
	timestampGen timestamp.Generator
}

// NewService returns a new object responsible for service-layer IntegrationDependency operations.
func NewService(repo IntegrationDependencyRepository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// Get returns an IntegrationDependency by given ID.
func (s *service) Get(ctx context.Context, id string) (*model.IntegrationDependency, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	integrationDependency, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Integration Dependency with ID: %q", id)
	}

	return integrationDependency, nil
}

// ListByApplicationID lists all integration dependencies for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByResourceID(ctx, tnt, resource.Application, appID)
}

// ListByApplicationTemplateVersionID lists all integration dependencies for a given application template version ID.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.IntegrationDependency, error) {
	return s.repo.ListByResourceID(ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID)
}

// Create creates integration dependency for a resource with given id.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) (string, error) {
	id := s.uidService.Generate()
	integrationDependency := in.ToIntegrationDependency(id, resourceType, resourceID, packageID, integrationDependencyHash)

	if err := s.createIntegrationDependency(ctx, resourceType, integrationDependency); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating an Integration Dependency with id %s for %s with id %s", id, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created a Integration Dependency with id %s for %s with id %s", id, resourceType, resourceID)

	return id, nil
}

// Update updates an existing Integration Dependency.
func (s *service) Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) error {
	integrationDependency, err := s.getIntegrationDependency(ctx, id, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while getting Integration Dependency with ID %s for %s", id, resourceType)
	}

	integrationDependency = in.ToIntegrationDependency(id, resourceType, resourceID, integrationDependency.PackageID, integrationDependencyHash)

	err = s.updateIntegrationDependency(ctx, integrationDependency, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while updating Integration Dependency with ID %s for %s", id, resourceType)
	}

	return nil
}

// Delete deletes the integration dependency by its ID.
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteIntegrationDependency(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Integration Dependency with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Integration Dependency with id %s", id)

	return nil
}

func (s *service) createIntegrationDependency(ctx context.Context, resourceType resource.Type, integrationDependency *model.IntegrationDependency) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.CreateGlobal(ctx, integrationDependency)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, integrationDependency)
}

func (s *service) getIntegrationDependency(ctx context.Context, id string, resourceType resource.Type) (*model.IntegrationDependency, error) {
	if resourceType.IsTenantIgnorable() {
		return s.repo.GetByIDGlobal(ctx, id)
	}

	return s.Get(ctx, id)
}

func (s *service) updateIntegrationDependency(ctx context.Context, integrationDependency *model.IntegrationDependency, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.UpdateGlobal(ctx, integrationDependency)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, tnt, integrationDependency)
}

func (s *service) deleteIntegrationDependency(ctx context.Context, integrationDependencyID string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.DeleteGlobal(ctx, integrationDependencyID)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, tnt, integrationDependencyID)
}
