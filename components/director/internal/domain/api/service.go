package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	GetByID(id string) (*model.APIDefinition, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	CreateMany(item []*model.APIDefinition) error
	Create(item *model.APIDefinition) error
	Update(item *model.APIDefinition) error
	Delete(item *model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo       APIRepository
	uidService UIDService
}

func NewService(repo APIRepository, uidService UIDService) *service {
	return &service{repo: repo, uidService: uidService}
}

func (s *service) List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error) {
	return s.repo.ListByApplicationID(applicationID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.APIDefinition, error) {
	api, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (s *service) Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error) {
	id := s.uidService.Generate()
	api := in.ToAPIDefinition(id, applicationID)

	err := s.repo.Create(api)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput) error {
	api, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	api = in.ToAPIDefinition(id, api.ApplicationID)

	err = s.repo.Update(api)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	api, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while receiving APIDefinition with ID %s", id)
	}

	err = s.repo.Delete(api)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) (*model.RuntimeAuth, error) {
	api, err := s.Get(ctx, apiID)
	if err != nil {
		return nil, err
	}

	var runtimeAuth *model.RuntimeAuth
	var runtimeAuthIndex int
	for i, a := range api.Auths {
		if a.RuntimeID == runtimeID {
			runtimeAuth = a
			runtimeAuthIndex = i
			break
		}
	}

	newAuth := &model.RuntimeAuth{
		RuntimeID: runtimeID,
		Auth:      in.ToAuth(),
	}

	if runtimeAuth == nil {
		api.Auths = append(api.Auths, newAuth)
	} else {
		api.Auths[runtimeAuthIndex] = newAuth
	}

	err = s.repo.Update(api)
	if err != nil {
		return nil, err
	}

	return newAuth, nil
}

func (s *service) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*model.RuntimeAuth, error) {
	api, err := s.Get(ctx, apiID)
	if err != nil {
		return nil, err
	}

	var runtimeAuth *model.RuntimeAuth
	var runtimeAuthIndex int
	for i, a := range api.Auths {
		if a.RuntimeID == runtimeID {
			runtimeAuth = a
			runtimeAuthIndex = i
			break
		}
	}

	if runtimeAuth == nil {
		return nil, nil
	}

	api.Auths = append(api.Auths[:runtimeAuthIndex], api.Auths[runtimeAuthIndex+1:]...)

	err = s.repo.Update(api)
	if err != nil {
		return nil, err
	}

	return runtimeAuth, nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error) {
	api, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return api.Spec, nil
}
