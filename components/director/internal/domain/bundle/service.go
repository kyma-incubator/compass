package mp_bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=BundleRepository -output=automock -outpkg=automock -case=underscore
type BundleRepository interface {
	Create(ctx context.Context, item *model.Bundle) error
	Update(ctx context.Context, item *model.Bundle) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error)
	GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Bundle, error)
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.BundlePage, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
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

func NewService(bndlRepo BundleRepository, apiSvc APIService, eventSvc EventService, documentSvc DocumentService, uidService UIDService) *service {
	return &service{
		bndlRepo:    bndlRepo,
		apiSvc:      apiSvc,
		eventSvc:    eventSvc,
		documentSvc: documentSvc,
		uidService:  uidService,
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.BundleCreateInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	bndl := in.ToBundle(id, applicationID, tnt)

	err = s.bndlRepo.Create(ctx, bndl)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)
	}
	log.C(ctx).Infof("Successfully created a Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)

	log.C(ctx).Infof("Creating related resources in Bundle with id %s and name %s for Application with id %s", id, bndl.Name, applicationID)
	err = s.createRelatedResources(ctx, in, id)
	if err != nil {
		return "", errors.Wrapf(err, "while creating related resources for Application with id %s", applicationID)
	}

	return id, nil
}

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

func (s *service) Update(ctx context.Context, id string, in model.BundleUpdateInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bndl, err := s.bndlRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Bundle with id %s", id)
	}

	bndl.SetFromUpdateInput(in)

	err = s.bndlRepo.Update(ctx, bndl)
	if err != nil {
		return errors.Wrapf(err, "while updating Bundle with id %s", id)
	}
	return nil
}

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

func (s *service) GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bndl, err := s.bndlRepo.GetByInstanceAuthID(ctx, tnt, instanceAuthID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle by BundleInstanceAuth with id %s", instanceAuthID)
	}

	return bndl, nil
}

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

func (s *service) ListByApplicationID(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.bndlRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) createRelatedResources(ctx context.Context, in model.BundleCreateInput, bundleID string) error {
	for i := range in.APIDefinitions {
		_, err := s.apiSvc.CreateInBundle(ctx, bundleID, *in.APIDefinitions[i], in.APISpecs[i])
		if err != nil {
			return errors.Wrapf(err, "while creating APIs for bundle with id %q", bundleID)
		}
	}

	for i := range in.EventDefinitions {
		_, err := s.eventSvc.CreateInBundle(ctx, bundleID, *in.EventDefinitions[i], in.EventSpecs[i])
		if err != nil {
			return errors.Wrapf(err, "while creating Event for bundle with id %q", bundleID)
		}
	}

	for _, document := range in.Documents {
		_, err := s.documentSvc.CreateInBundle(ctx, bundleID, *document)
		if err != nil {
			return errors.Wrapf(err, "while creating Document for bundle with id %q", bundleID)
		}
	}

	return nil
}
