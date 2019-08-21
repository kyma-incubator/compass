package api

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	GetByID(id string) (*model.APIDefinition, error)
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	CreateMany(item []*model.APIDefinition) error
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	Create(item *model.APIDefinition) error
	Update(item *model.APIDefinition) error
	Delete(item *model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, tenant string, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
	Delete(ctx context.Context, tenant, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo             APIRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
}

func NewService(repo APIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{repo: repo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
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

	fetchRequestID, err := s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
	if err != nil {
		return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
	}

	api := in.ToAPIDefinition(id, applicationID, fetchRequestID)

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

	if api.Spec.FetchRequestID != nil {
		err := s.fetchRequestRepo.Delete(ctx, tnt, *api.Spec.FetchRequestID)
		if err != nil {
			return errors.Wrapf(err, "while deleting FetchRequest for APIDefinition %s", id)
		}
	}

	fetchRequestID, err := s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
	if err != nil {
		return errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
	}

	api = in.ToAPIDefinition(id, api.ApplicationID, fetchRequestID)

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

	// FetchRequest is deleted automatically because of cascading delete

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

func (s *service) GetFetchRequest(ctx context.Context, apiDefID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.APIFetchRequestReference, apiDefID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", apiDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, parentObjectID string) (*string, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(time.Now(), id, model.APIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, tenant, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.APIFetchRequestReference, parentObjectID)
	}

	return &id, nil
}