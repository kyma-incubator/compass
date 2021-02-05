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
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListForBundle(ctx context.Context, tenantID, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	CreateMany(ctx context.Context, item []*model.APIDefinition) error
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleSpec(ctx context.Context, fr *model.FetchRequest) *string
}

//go:generate mockery -name=SpecService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
}

type service struct {
	repo         APIRepository
	uidService   UIDService
	specService  SpecService
	timestampGen timestamp.Generator
}

func NewService(repo APIRepository, uidService UIDService, specService SpecService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		specService:  specService,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.ListForBundle(ctx, tnt, bundleID, pageSize, cursor)
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

func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	apiDefinition, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API definition")
	}

	return apiDefinition, nil
}

func (s *service) CreateInBundle(ctx context.Context, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	api := in.ToAPIDefinitionWithinBundle(id, bundleID, tnt)

	err = s.repo.Create(ctx, api)
	if err != nil {
		return "", errors.Wrap(err, "while creating api")
	}

	if spec != nil {
		_, err = s.specService.CreateByReferenceObjectID(ctx, *spec, model.APISpecReference, api.ID)
		if err != nil {
			return "", err
		}
	}

	return id, nil
}
func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	api, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	api = in.ToAPIDefinition(id, api.BundleID, api.PackageID, tnt)

	err = s.repo.Update(ctx, api)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with id %s", id)
	}

	if specIn != nil {
		dbSpec, err := s.specService.GetByReferenceObjectID(ctx, model.APISpecReference, api.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting spec for APIDefinition with id %q", api.ID)
		}

		if dbSpec == nil {
			_, err = s.specService.CreateByReferenceObjectID(ctx, *specIn, model.APISpecReference, api.ID)
			return err
		}

		return s.specService.UpdateByReferenceObjectID(ctx, dbSpec.ID, *specIn, model.APISpecReference, api.ID)
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
		return errors.Wrapf(err, "while deleting APIDefinition with id %s", id)
	}

	return nil
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
		return nil, fmt.Errorf("API Definition with id %s doesn't exist", apiDefID)
	}

	spec, err := s.specService.GetByReferenceObjectID(ctx, model.APISpecReference, apiDefID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for APIDefinition with id %q", apiDefID)
	}

	var fetchRequest *model.FetchRequest
	if spec != nil {
		fetchRequest, err = s.specService.GetFetchRequest(ctx, spec.ID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return nil, nil
			}
			return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition with id %q", apiDefID)
		}
	}

	return fetchRequest, nil
}
