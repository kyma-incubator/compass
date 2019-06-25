package api

import (
	"context"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	GetByID(id string) (*model.APIDefinition, error)
	ListByApplicationID(applicationID string) ([]*model.APIDefinition, error)
	CreateMany(item []*model.APIDefinition) error
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	Create(item *model.APIDefinition) error
	Update(item *model.APIDefinition) error
	Delete(item *model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
}

type service struct {
	repo APIRepository
}

func NewService(repo APIRepository) *service{
	return &service{repo: repo}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.APIDefinitionPage, error) {
	return s.repo.List(filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.APIDefinition, error) {
	api, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (s *service) Create(ctx context.Context, in model.APIDefinitionInput) (string, error) {
	id := uuid.New().String()

	api := &model.APIDefinition{
		ID: id,
		Name: in.Name,
		Description: in.Description,
		Spec: in.Spec.ToAPISpec(),
	}

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

	api.Name = in.Name
	api.Description = in.Description
	api.TargetURL = in.TargetURL
	api.Group = in.Group
	api.Spec = in.Spec.ToAPISpec()
	api.DefaultAuth = in.DefaultAuth.ToAuth() //TODO: https://github.com/kyma-incubator/compass/issues/67
	api.Version = in.Version.ToVersion()

	err = s.repo.Update(api)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	api, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with ID %s", id)
	}

	return s.repo.Delete(api)
}

func (s *service) SetApiAuth(ctx context.Context, apiID string, runtimeID string, in model.AuthInput) error {
	api, err := s.Get(ctx,apiID)
	if err != nil {
		return err
	}

	for i, item := range api.Auths {
		if item.RuntimeID == runtimeID {
			//apiAuth = item
			api.Auth = item
			api.Auths[i] = item

			err = s.repo.Update(api)
			if err != nil {
				return err
			}

			return nil
		}
	}

	auth := &model.RuntimeAuth{
		RuntimeID: runtimeID,
		Auth:      in.ToAuth(),
	}
	api.Auth = auth
	api.Auths = append(api.Auths,auth)

	err = s.repo.Update(api)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) error{
	api, err := s.Get(ctx,apiID)
	if err != nil {
		return err
	}

	for i, item := range api.Auths {
		if item.RuntimeID == runtimeID {
			api.Auths = append(api.Auths[:i], api.Auths[i+1:]...)

			err := s.repo.Update(api)
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.APISpec,error) {
	api, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return api.Spec,nil
}