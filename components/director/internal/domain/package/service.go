package mp_bundle

import (
	"context"

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
	HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string
}

type service struct {
	bundleRepo       BundleRepository
	apiRepo          APIRepository
	eventAPIRepo     EventAPIRepository
	documentRepo     DocumentRepository
	fetchRequestRepo FetchRequestRepository

	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(bundleRepo BundleRepository, apiRepo APIRepository, eventAPIRepo EventAPIRepository, documentRepo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{
		bundleRepo:          bundleRepo,
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
	bundle := in.ToBundle(id, applicationID, tnt)

	err = s.bundleRepo.Create(ctx, bundle)
	if err != nil {
		return "", err
	}

	err = s.createRelatedResources(ctx, in, tnt, id)
	if err != nil {
		return "", errors.Wrap(err, "while creating related Application resources")
	}

	return id, nil
}

func (s *service) CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleCreateInput) error {
	if in == nil {
		return nil
	}

	for _, bundle := range in {
		if bundle == nil {
			continue
		}

		_, err := s.Create(ctx, applicationID, *bundle)
		if err != nil {
			return errors.Wrap(err, "while creating Bundle for Application")
		}
	}

	return nil
}

func (s *service) Update(ctx context.Context, id string, in model.BundleUpdateInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bundle, err := s.bundleRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	bundle.SetFromUpdateInput(in)

	err = s.bundleRepo.Update(ctx, bundle)
	if err != nil {
		return errors.Wrapf(err, "while updating Bundle with ID: [%s]", id)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.bundleRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Bundle with ID: [%s]", id)
	}

	return nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.bundleRepo.Exists(ctx, tnt, id)
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

	bundle, err := s.bundleRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bundle, nil
}

func (s *service) GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bundle, err := s.bundleRepo.GetByInstanceAuthID(ctx, tnt, instanceAuthID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle by Instance Auth ID: [%s]", instanceAuthID)
	}

	return bundle, nil
}

func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bundle, err := s.bundleRepo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bundle, nil
}

func (s *service) ListByApplicationID(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
	}

	return s.bundleRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) createRelatedResources(ctx context.Context, in model.BundleCreateInput, tenant string, bundleID string) error {
	err := s.createAPIs(ctx, bundleID, tenant, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.createEvents(ctx, bundleID, tenant, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	err = s.createDocuments(ctx, bundleID, tenant, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) createAPIs(ctx context.Context, bundleID, tenant string, apis []*model.APIDefinitionInput) error {
	var err error
	for _, item := range apis {
		apiDefID := s.uidService.Generate()

		api := item.ToAPIDefinitionWithinBundle(apiDefID, bundleID, tenant)

		err = s.apiRepo.Create(ctx, api)
		if err != nil {
			return errors.Wrap(err, "while creating API for application")
		}

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			fr, err := s.createFetchRequest(ctx, tenant, item.Spec.FetchRequest, model.APIFetchRequestReference, apiDefID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}

			api.Spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fr)
			err = s.apiRepo.Update(ctx, api)
			if err != nil {
				return errors.Wrap(err, "while updating api with api spec")
			}
		}

	}

	return nil
}

func (s *service) createEvents(ctx context.Context, bundleID, tenant string, events []*model.EventDefinitionInput) error {
	var err error
	for _, item := range events {
		eventID := s.uidService.Generate()
		err = s.eventAPIRepo.Create(ctx, item.ToEventDefinitionWithinBundle(eventID, bundleID, tenant))
		if err != nil {
			return errors.Wrap(err, "while creating EventDefinitions for application")
		}

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.Spec.FetchRequest, model.EventAPIFetchRequestReference, eventID)
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
			return errors.Wrapf(err, "while creating Document for application")
		}

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
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", objectType, objectID)
	}

	return fr, nil
}
