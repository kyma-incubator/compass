package api

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.APIDefinition, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	CreateMany(ctx context.Context, item []*model.APIDefinition) error
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByApplicationID(ctx context.Context, tenantID string, id string) error
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

func (s *service) List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("page size must be between 1 and 100")
	}

	return s.repo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	api, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return api, nil
}

func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	apiDefinition, err := s.repo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API definition")
	}

	return apiDefinition, nil
}

func (s *service) Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()

	api := in.ToAPIDefinition(id, &applicationID, tnt)
	err = s.repo.Create(ctx, api)
	if err != nil {
		return "", err
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, *in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}
	}

	return id, nil
}

func (s *service) CreateToPackage(ctx context.Context, packageID string, in model.APIDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()

	api := in.ToAPIDefinitionWithPackage(id, packageID, tnt)
	err = s.repo.Create(ctx, api)
	if err != nil {
		return "", err
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, *in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}
	}

	return id, nil
}
func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
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
		_, err = s.createFetchRequest(ctx, tnt, *in.Spec.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}
	}

	api = in.ToAPIDefinition(id, api.ApplicationID, tnt)

	err = s.repo.Update(ctx, api)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIDefinition with ID %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.APISpec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	api, err := s.repo.GetByID(ctx, tnt, id)
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
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", apiDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in model.FetchRequestInput, parentObjectID string) (*string, error) {
	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.APIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.APIFetchRequestReference, parentObjectID)
	}

	return &id, nil
}
