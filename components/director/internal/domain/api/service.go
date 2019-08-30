package api

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	GetByID(id string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	CreateMany(item []*model.APIDefinition) error
	Create(item *model.APIDefinition) error
	Update(item *model.APIDefinition) error
	Delete(item *model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
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
	repo             APIRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
	timestampGen     timestamp.Generator
}

func NewService(repo APIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{repo: repo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
		timestampGen:     timestamp.DefaultGenerator(),
	}
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
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}
	}

	api := in.ToAPIDefinition(id, applicationID)

	err = s.repo.Create(api)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	api, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.APIFetchRequestReference, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting FetchRequest for APIDefinition %s", id)
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}
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

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error) {
	api, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return api.Spec, nil
}

func (s *service) GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.Exists(ctx, tnt, apiDefID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking if API Definition exists")
	}
	if !exists {
		return nil, fmt.Errorf("API Definition with ID %s doesn't exist", apiDefID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.APIFetchRequestReference, apiDefID)
	if err != nil {
		if repo.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", apiDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, parentObjectID string) (*string, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.APIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.APIFetchRequestReference, parentObjectID)
	}

	return &id, nil
}
