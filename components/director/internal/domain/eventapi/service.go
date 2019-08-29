package eventapi

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	GetByID(id string) (*model.EventAPIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error)
	Create(item *model.EventAPIDefinition) error
	CreateMany(items []*model.EventAPIDefinition) error
	Update(item *model.EventAPIDefinition) error
	Delete(item *model.EventAPIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
	Delete(ctx context.Context, tenant, id string) error
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

func (s *service) List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
	return s.eventAPIRepo.ListByApplicationID(applicationID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.EventAPIDefinition, error) {
	eventAPI, err := s.eventAPIRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return eventAPI, nil
}

func (s *service) Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()

	var fetchRequestID *string
	if in.Spec != nil && in.Spec.FetchRequest != nil {
		fetchRequestID, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for EventAPIDefinition %s", id)
		}
	}
	eventAPI := in.ToEventAPIDefinition(id, applicationID, fetchRequestID)

	err = s.eventAPIRepo.Create(eventAPI)
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

	if eventAPI.Spec != nil && eventAPI.Spec.FetchRequestID != nil {
		err := s.fetchRequestRepo.Delete(ctx, tnt, *eventAPI.Spec.FetchRequestID)
		if err != nil {
			return errors.Wrapf(err, "while deleting FetchRequest for EventAPIDefinition %s", id)
		}
	}

	var fetchRequestID *string
	if in.Spec != nil && in.Spec.FetchRequest != nil {
		fetchRequestID, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for EventAPIDefinition %s", id)
		}
	}

	eventAPI = in.ToEventAPIDefinition(id, eventAPI.ApplicationID, fetchRequestID)

	err = s.eventAPIRepo.Update(eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while updating EventAPIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while receiving EventAPIDefinition with ID %s", id)
	}

	err = s.eventAPIRepo.Delete(eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.EventAPISpec, error) {
	eventAPI, err := s.eventAPIRepo.GetByID(id)
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
		if repo.IsNotFoundError(err) {
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
