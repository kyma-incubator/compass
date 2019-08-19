package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	GetByID(id string) (*model.EventAPIDefinition, error)
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
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	eventAPIRepo     EventAPIRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
}

func NewService(eventAPIRepo EventAPIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{eventAPIRepo: eventAPIRepo, fetchRequestRepo: fetchRequestRepo, uidService: uidService}
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
	id := s.uidService.Generate()
	eventAPI := in.ToEventAPIDefinition(id, applicationID)

	err := s.eventAPIRepo.Create(eventAPI)
	if err != nil {
		return "", err
	}

	if eventAPI.Spec != nil && eventAPI.Spec.FetchRequest != nil {
		err := s.fetchRequestRepo.Create(ctx, eventAPI.Spec.FetchRequest)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for EventAPI '%s'", eventAPI.Name)
		}
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error {
	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	eventAPI = in.ToEventAPIDefinition(id, eventAPI.ApplicationID)

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

	// FetchRequest is deleted automatically because of cascading delete

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.EventAPISpec, error) {
	eventAPI, err := s.eventAPIRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return eventAPI.Spec, nil
}
