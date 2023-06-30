package bundle

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// BundleRepository missing godoc
//
//go:generate mockery --name=BundleRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleRepository interface {
	Create(ctx context.Context, tenant string, item *model.Bundle) error
	CreateGlobal(ctx context.Context, model *model.Bundle) error
	Update(ctx context.Context, tenant string, item *model.Bundle) error
	UpdateGlobal(ctx context.Context, model *model.Bundle) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Bundle, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error)
	ListByResourceIDNoPaging(ctx context.Context, tenantID, appID string, resourceType resource.Type) ([]*model.Bundle, error)
	ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	bndlRepo    BundleRepository
	apiSvc      APIService
	eventSvc    EventService
	documentSvc DocumentService
	biaSvc      BundleInstanceAuthService

	uidService UIDService
}

// NewService missing godoc
func NewService(bndlRepo BundleRepository, apiSvc APIService, eventSvc EventService, documentSvc DocumentService, biaSvc BundleInstanceAuthService, uidService UIDService) *service {
	return &service{
		bndlRepo:    bndlRepo,
		apiSvc:      apiSvc,
		eventSvc:    eventSvc,
		documentSvc: documentSvc,
		biaSvc:      biaSvc,
		uidService:  uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.BundleCreateInput) (string, error) {
	return s.CreateBundle(ctx, resourceType, resourceID, in, 0)
}

// CreateBundle Creates bundle for an application with given id
func (s *service) CreateBundle(ctx context.Context, resourceType resource.Type, resourceID string, in model.BundleCreateInput, bndlHash uint64) (string, error) {
	id := s.uidService.Generate()
	bndl := in.ToBundle(id, resourceType, resourceID, bndlHash)

	if err := s.createBundle(ctx, bndl, resourceType); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Bundle with id %s and name %s for %s with id %s", id, bndl.Name, resourceType, resourceID)
	}

	log.C(ctx).Infof("Successfully created a Bundle with id %s and name %s for %s with id %s", id, bndl.Name, resourceType, resourceID)

	log.C(ctx).Infof("Creating related resources in Bundle with id %s and name %s for %s with id %s", id, bndl.Name, resourceType, resourceID)
	if err := s.createRelatedResources(ctx, in, id, resourceType, resourceID); err != nil {
		return "", errors.Wrapf(err, "while creating related resources for %s with id %s", resourceType, resourceID)
	}

	return id, nil
}

// CreateMultiple missing godoc
func (s *service) CreateMultiple(ctx context.Context, resourceType resource.Type, resourceID string, in []*model.BundleCreateInput) error {
	if in == nil {
		return nil
	}

	for _, bndl := range in {
		if bndl == nil {
			continue
		}

		if _, err := s.Create(ctx, resourceType, resourceID, *bndl); err != nil {
			return errors.Wrapf(err, "while creating Bundle for %s with id %s", resourceType, resourceID)
		}
	}

	return nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.BundleUpdateInput) error {
	return s.UpdateBundle(ctx, resourceType, id, in, 0)
}

// UpdateBundle missing godoc
func (s *service) UpdateBundle(ctx context.Context, resourceType resource.Type, id string, in model.BundleUpdateInput, bndlHash uint64) error {
	bndl, err := s.getBundle(ctx, id, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while getting Bundle with id %s", id)
	}

	bndl.SetFromUpdateInput(in, bndlHash)

	if err = s.updateBundle(ctx, bndl, resourceType); err != nil {
		return errors.Wrapf(err, "while updating Bundle with id %s", id)
	}

	return nil
}

// Delete deletes a bundle by id
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteBundle(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting Bundle with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted a bundle with ID %s", id)

	return nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.bndlRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return exist, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	bndl, err := s.bndlRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bndl, nil
}

// GetForApplication missing godoc
func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndl, err := s.bndlRepo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bndl, nil
}

// ListByApplicationIDNoPaging missing godoc
func (s *service) ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.bndlRepo.ListByResourceIDNoPaging(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionIDNoPaging lists bundles by Application Template Version ID without tenant isolation
func (s *service) ListByApplicationTemplateVersionIDNoPaging(ctx context.Context, appTemplateVersionID string) ([]*model.Bundle, error) {
	return s.bndlRepo.ListByResourceIDNoPaging(ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion)
}

// ListByApplicationIDs missing godoc
func (s *service) ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	bundlePages, err := s.bndlRepo.ListByApplicationIDs(ctx, tnt, applicationIDs, pageSize, cursor)
	if err != nil {
		return nil, err
	}

	// override the default instance auth only for runtime consumers
	if consumerInfo.ConsumerType != consumer.Runtime {
		return bundlePages, nil
	}

	bundleInstanceAuths, err := s.biaSvc.ListByRuntimeID(ctx, consumerInfo.ConsumerID)
	if err != nil {
		return nil, err
	}

	bundlesCount := 0
	for _, page := range bundlePages {
		bundlesCount += len(page.Data)
	}

	bundleIDToBundleInstanceAuths := make(map[string][]*model.BundleInstanceAuth, bundlesCount)
	for i, auth := range bundleInstanceAuths {
		if _, ok := bundleIDToBundleInstanceAuths[auth.BundleID]; !ok {
			bundleIDToBundleInstanceAuths[auth.BundleID] = []*model.BundleInstanceAuth{bundleInstanceAuths[i]}
		} else {
			bundleIDToBundleInstanceAuths[auth.BundleID] = append(bundleIDToBundleInstanceAuths[auth.BundleID], bundleInstanceAuths[i])
		}
	}

	for _, page := range bundlePages {
		for _, bundle := range page.Data {
			if auths, ok := bundleIDToBundleInstanceAuths[bundle.ID]; ok {
				log.C(ctx).Infof("Overrinding default instance auth for bundle with ID: %s", bundle.ID)
				bundle.DefaultInstanceAuth = auths[0].Auth
			}
		}
	}

	return bundlePages, nil
}

func (s *service) createRelatedResources(ctx context.Context, in model.BundleCreateInput, bundleID string, resourceType resource.Type, resourceID string) error {
	for i := range in.APIDefinitions {
		if _, err := s.apiSvc.CreateInBundle(ctx, resourceType, resourceID, bundleID, *in.APIDefinitions[i], in.APISpecs[i]); err != nil {
			return errors.Wrapf(err, "while creating APIs for bundle with id %q", bundleID)
		}
	}

	for i := range in.EventDefinitions {
		if _, err := s.eventSvc.CreateInBundle(ctx, resourceType, resourceID, bundleID, *in.EventDefinitions[i], in.EventSpecs[i]); err != nil {
			return errors.Wrapf(err, "while creating Event for bundle with id %q", bundleID)
		}
	}

	for _, document := range in.Documents {
		if _, err := s.documentSvc.CreateInBundle(ctx, resourceType, resourceID, bundleID, *document); err != nil {
			return errors.Wrapf(err, "while creating Document for bundle with id %q", bundleID)
		}
	}

	return nil
}

func (s *service) createBundle(ctx context.Context, bundle *model.Bundle, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.bndlRepo.CreateGlobal(ctx, bundle)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.bndlRepo.Create(ctx, tnt, bundle)
}

func (s *service) getBundle(ctx context.Context, bundleID string, resourceType resource.Type) (*model.Bundle, error) {
	if resourceType.IsTenantIgnorable() {
		return s.bndlRepo.GetByIDGlobal(ctx, bundleID)
	}

	return s.Get(ctx, bundleID)
}

func (s *service) updateBundle(ctx context.Context, bndl *model.Bundle, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.bndlRepo.UpdateGlobal(ctx, bndl)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.bndlRepo.Update(ctx, tnt, bndl)
}

func (s *service) deleteBundle(ctx context.Context, bundleID string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.bndlRepo.DeleteGlobal(ctx, bundleID)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.bndlRepo.Delete(ctx, tnt, bundleID)
}
