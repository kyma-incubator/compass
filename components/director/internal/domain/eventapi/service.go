package eventapi

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	GetByID(ctx context.Context, tenantID string, id string) (*model.EventAPIDefinition, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.EventAPIDefinition, error)
	Exists(ctx context.Context, tenantID, id string) (bool, error)
	ListByApplicationID(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.EventAPIDefinitionPage, error)
	Create(ctx context.Context, item *model.EventAPIDefinition) error
	CreateMany(ctx context.Context, items []*model.EventAPIDefinition) error
	Update(ctx context.Context, item *model.EventAPIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByApplicationID(ctx context.Context, tenantID string, appID string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
	DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	eventAPIRepo     EventAPIRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
	timestampGen     timestamp.Generator
}

func NewService(eventAPIRepo EventAPIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{eventAPIRepo: eventAPIRepo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
		timestampGen:     timestamp.DefaultGenerator(),
	}
}

func (s *service) List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.EventAPIDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("page size must be between 1 and 100")
	}

	return s.eventAPIRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.EventAPIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventAPI, err := s.eventAPIRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return eventAPI, nil
}

func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.EventAPIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventAPI, err := s.eventAPIRepo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API definition")
	}

	return eventAPI, nil
}

func (s *service) Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for EventAPIDefinition %s", id)
		}
	}
	eventAPI := in.ToEventAPIDefinition(id, applicationID, tnt)

	err = s.eventAPIRepo.Create(ctx, eventAPI)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.EventAPIFetchRequestReference, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting FetchRequest for EventAPIDefinition %s", id)
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for EventAPIDefinition %s", id)
		}
	}

	eventAPI = in.ToEventAPIDefinition(id, eventAPI.ApplicationID, tnt)

	err = s.eventAPIRepo.Update(ctx, eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while updating EventAPIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.eventAPIRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.EventAPISpec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	eventAPI, err := s.eventAPIRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return eventAPI.Spec, nil
}

func (s *service) GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.eventAPIRepo.Exists(ctx, tnt, eventAPIDefID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking if EventAPI Definition exists")
	}
	if !exists {
		return nil, fmt.Errorf("EventAPI Definition with ID %s doesn't exist", eventAPIDefID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.EventAPIFetchRequestReference, eventAPIDefID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by Event API Definition ID %s", eventAPIDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, parentObjectID string) (*string, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.EventAPIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.EventAPIFetchRequestReference, parentObjectID)
	}

	return &id, nil
}
