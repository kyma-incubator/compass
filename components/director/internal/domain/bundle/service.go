package mp_bundle

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
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

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
}

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	Create(ctx context.Context, items *model.EventDefinition) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	Create(ctx context.Context, item *model.Document) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleSpec(ctx context.Context, fr *model.FetchRequest) *string
}

type service struct {
	bndlRepo         BundleRepository
	apiRepo          APIRepository
	eventAPIRepo     EventAPIRepository
	documentRepo     DocumentRepository
	fetchRequestRepo FetchRequestRepository

	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(bndlRepo BundleRepository, apiRepo APIRepository, eventAPIRepo EventAPIRepository, documentRepo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{
		bndlRepo:            bndlRepo,
		apiRepo:             apiRepo,
		eventAPIRepo:        eventAPIRepo,
		documentRepo:        documentRepo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		timestampGen:        timestamp.DefaultGenerator(),
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
	err = s.createRelatedResources(ctx, in, tnt, id)
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

func (s *service) createRelatedResources(ctx context.Context, in model.BundleCreateInput, tenantID string, bundleID string) error {
	err := s.createAPIs(ctx, bundleID, tenantID, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.createEvents(ctx, bundleID, tenantID, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	err = s.createDocuments(ctx, bundleID, tenantID, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) createAPIs(ctx context.Context, bundleID, tenantID string, apis []*model.APIDefinitionInput) error {
	var err error
	for _, item := range apis {
		apiDefID := s.uidService.Generate()

		api := item.ToAPIDefinitionWithinBundle(apiDefID, bundleID, tenantID)

		err = s.apiRepo.Create(ctx, api)
		if err != nil {
			return errors.Wrapf(err, "while creating APIDefinition with id %s within Bundle with id %s", apiDefID, bundleID)
		}
		log.C(ctx).Infof("Successfully created APIDefinition with id %s within Bundle with id %s", apiDefID, bundleID)

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			fr, err := s.createFetchRequest(ctx, tenantID, item.Spec.FetchRequest, model.APIFetchRequestReference, apiDefID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}

			api.Spec.Data = s.fetchRequestService.HandleSpec(ctx, fr)
			err = s.apiRepo.Update(ctx, api)
			if err != nil {
				return errors.Wrap(err, "while updating api with api spec")
			}
		}

	}

	return nil
}

func (s *service) createEvents(ctx context.Context, bundleID, tenantID string, events []*model.EventDefinitionInput) error {
	var err error
	for _, item := range events {
		eventID := s.uidService.Generate()

		err = s.eventAPIRepo.Create(ctx, item.ToEventDefinitionWithinBundle(eventID, bundleID, tenantID))
		if err != nil {
			return errors.Wrapf(err, "while creating EventDefinition with id %s in Bundle with id %s", eventID, bundleID)
		}
		log.C(ctx).Infof("Successfully created EventDefinition with id %s in Bundle with id %s", eventID, bundleID)

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenantID, item.Spec.FetchRequest, model.EventAPIFetchRequestReference, eventID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}
		}
	}
	return nil
}

func (s *service) createDocuments(ctx context.Context, bundleID, tenant string, events []*model.DocumentInput) error {
	var err error
	for _, item := range events {
		documentID := s.uidService.Generate()

		err = s.documentRepo.Create(ctx, item.ToDocumentWithinBundle(documentID, tenant, bundleID))
		if err != nil {
			return errors.Wrapf(err, "while creating Document with id %s in Bundle with id %s", documentID, bundleID)
		}
		log.C(ctx).Infof("Successfully created Document with id %s in Bundle with id %s", documentID, bundleID)

		if item.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.FetchRequest, model.DocumentFetchRequestReference, documentID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, objectType, objectID)

	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest with id %s for %s with id %s", id, objectType, objectID)
	}
	log.C(ctx).Infof("Successfully created FetchRequest with id %s for type %s with id %s", id, objectType, objectID)
	return fr, nil
}
