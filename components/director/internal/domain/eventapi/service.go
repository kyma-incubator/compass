package eventapi

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"

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

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo       EventAPIRepository
	uidService UIDService
}

func NewService(repo EventAPIRepository, uidService api.UIDService) *service {
	return &service{repo: repo, uidService: uidService}
}

func (s *service) List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error) {
	return s.repo.ListByApplicationID(applicationID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.EventAPIDefinition, error) {
	eventAPI, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return eventAPI, nil
}

func (s *service) Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error) {
	id := s.uidService.Generate()
	eventAPI := in.ToEventAPIDefinition(id, applicationID)

	err := s.repo.Create(eventAPI)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error {
	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	eventAPI = in.ToEventAPIDefinition(id, eventAPI.ApplicationID)

	err = s.repo.Update(eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while updating EventAPIDefinition with Field %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while receiving EventAPIDefinition with Field %s", id)
	}

	err = s.repo.Delete(eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIDefinition with Field %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.EventAPISpec, error) {
	eventAPI, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return eventAPI.Spec, nil
}
