package capability

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// CapabilityRepository is responsible for the repo-layer Capability operations.
//
//go:generate mockery --name=CapabilityRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type CapabilityRepository interface {
	ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.Capability, error)
	GetByID(ctx context.Context, tenantID, id string) (*model.Capability, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Capability, error)
	Create(ctx context.Context, tenant string, item *model.Capability) error
	CreateGlobal(ctx context.Context, item *model.Capability) error
	Update(ctx context.Context, tenant string, item *model.Capability) error
	UpdateGlobal(ctx context.Context, item *model.Capability) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteGlobal(ctx context.Context, id string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal capability IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// SpecService is responsible for the service-layer Specification operations.
//
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
}

type service struct {
	repo         CapabilityRepository
	uidService   UIDService
	specService  SpecService
	timestampGen timestamp.Generator
}

// NewService returns a new object responsible for service-layer Capability operations.
func NewService(repo CapabilityRepository, uidService UIDService, specService SpecService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		specService:  specService,
		timestampGen: timestamp.DefaultGenerator,
	}
}

// ListByApplicationID lists all Capabilities for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Capability, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByResourceID(ctx, tnt, resource.Application, appID)
}

// ListByApplicationTemplateVersionID lists all Capabilities for a given application template version ID.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Capability, error) {
	return s.repo.ListByResourceID(ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID)
}

// Get returns the Capability by its ID.
func (s *service) Get(ctx context.Context, id string) (*model.Capability, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	capability, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return capability, nil
}

// Create creates Capability.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.CapabilityInput, specs []*model.SpecInput, capabilityHash uint64) (string, error) {
	id := s.uidService.Generate()

	capability := in.ToCapability(id, resourceType, resourceID, packageID, capabilityHash)

	if err := s.createCapability(ctx, capability, resourceType); err != nil {
		return "", errors.Wrap(err, "while creating capability")
	}

	if err := s.processSpecs(ctx, capability.ID, specs, resourceType); err != nil {
		return "", errors.Wrap(err, "while processing specs")
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.CapabilityInput, capabilityHash uint64) error {
	capability, err := s.getCapability(ctx, id, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while getting Capability with ID %s for %s", id, resourceType)
	}

	resourceID := getParentResourceID(capability)
	capability = in.ToCapability(id, resourceType, resourceID, capability.PackageID, capabilityHash)

	err = s.updateCapability(ctx, capability, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while updating Capability with ID %s for %s", id, resourceType)
	}

	return nil
}

// Delete deletes the Capability by its ID.
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteCapability(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Capability with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Capability with id %s", id)

	return nil
}

func (s *service) createCapability(ctx context.Context, capability *model.Capability, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.CreateGlobal(ctx, capability)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, capability)
}

func (s *service) getCapability(ctx context.Context, id string, resourceType resource.Type) (*model.Capability, error) {
	if resourceType.IsTenantIgnorable() {
		return s.repo.GetByIDGlobal(ctx, id)
	}
	return s.Get(ctx, id)
}

func (s *service) updateCapability(ctx context.Context, api *model.Capability, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.UpdateGlobal(ctx, api)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, tnt, api)
}

func (s *service) processSpecs(ctx context.Context, capabilityID string, specs []*model.SpecInput, resourceType resource.Type) error {
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		if _, err := s.specService.CreateByReferenceObjectID(ctx, *spec, resourceType, model.CapabilitySpecReference, capabilityID); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) deleteCapability(ctx context.Context, capabilityID string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.DeleteGlobal(ctx, capabilityID)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, tnt, capabilityID)
}

func getParentResourceID(capability *model.Capability) string {
	if capability.ApplicationTemplateVersionID != nil {
		return *capability.ApplicationTemplateVersionID
	} else if capability.ApplicationID != nil {
		return *capability.ApplicationID
	}

	return ""
}
