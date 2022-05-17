package bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// BundleRepository missing godoc
//go:generate mockery --name=BundleRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleRepository interface {
	Create(ctx context.Context, tenant string, item *model.Bundle) error
	Update(ctx context.Context, tenant string, item *model.Bundle) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error)
	ListByApplicationIDNoPaging(ctx context.Context, tenantID, appID string) ([]*model.Bundle, error)
	ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	bndlRepo    BundleRepository
	apiSvc      APIService
	eventSvc    EventService
	documentSvc DocumentService

	uidService UIDService
}

// NewService missing godoc
func NewService(bndlRepo BundleRepository, apiSvc APIService, eventSvc EventService, documentSvc DocumentService, uidService UIDService) *service {
	return &service{
		bndlRepo:    bndlRepo,
		apiSvc:      apiSvc,
		eventSvc:    eventSvc,
		documentSvc: documentSvc,
		uidService:  uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	bndl := in.ToBundle(id, applicationID)

	if err = s.bndlRepo.Create(ctx, tnt, bndl); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)
	}
	log.C(ctx).Infof("Successfully created a Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)

	log.C(ctx).Infof("Creating related resources in Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)
	err = s.createRelatedResources(ctx, in, id, applicationID)
	if err != nil {
		return "", errors.Wrapf(err, "while creating related resources for Application with id %s", applicationID)
	}

	return id, nil
}

// CreateMultiple missing godoc
func (s *service) CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleCreateInput) error {
	if in == nil {
		return nil
	}

	for _, bndl := range in {
		if bndl == nil {
			continue
		}

		_, err := s.Create(ctx, applicationID, *bndl)
		if err != nil {
			return errors.Wrapf(err, "while creating Bundle for Application with id %s", applicationID)
		}
	}

	return nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.BundleUpdateInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bndl, err := s.bndlRepo.GetByID(ctx, tnt, id)

	return updateBundle(ctx, id, in, err, bndl, s, tnt)
}

func updateBundle(ctx context.Context, id string, in model.BundleUpdateInput, err error, bndl *model.Bundle, s *service, tnt string) error {
	if err != nil {
		return errors.Wrapf(err, "while getting Bundle with id %s", id)
	}

	bndl.SetFromUpdateInput(in)

	if err = s.bndlRepo.Update(ctx, tnt, bndl); err != nil {
		return errors.Wrapf(err, "while updating Bundle with id %s", id)
	}
	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.bndlRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Bundle with id %s", id)
	}

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

	return s.bndlRepo.ListByApplicationIDNoPaging(ctx, tnt, appID)
}

// ListByApplicationIDs missing godoc
func (s *service) ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.bndlRepo.ListByApplicationIDs(ctx, tnt, applicationIDs, pageSize, cursor)
}

func (s *service) createRelatedResources(ctx context.Context, in model.BundleCreateInput, bundleID, appID string) error {
	for i := range in.APIDefinitions {
		if _, err := s.apiSvc.CreateInBundle(ctx, appID, bundleID, *in.APIDefinitions[i], in.APISpecs[i]); err != nil {
			return errors.Wrapf(err, "while creating APIs for bundle with id %q", bundleID)
		}
	}

	for i := range in.EventDefinitions {
		if _, err := s.eventSvc.CreateInBundle(ctx, appID, bundleID, *in.EventDefinitions[i], in.EventSpecs[i]); err != nil {
			return errors.Wrapf(err, "while creating Event for bundle with id %q", bundleID)
		}
	}

	for _, document := range in.Documents {
		if _, err := s.documentSvc.CreateInBundle(ctx, appID, bundleID, *document); err != nil {
			return errors.Wrapf(err, "while creating Document for bundle with id %q", bundleID)
		}
	}

	return nil
}
