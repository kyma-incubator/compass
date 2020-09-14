package api

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

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
	ExistsByCondition(ctx context.Context, tenant string, conds repo.Conditions) (bool, error)
	GetByConditions(ctx context.Context, tenant string, conds repo.Conditions) (*model.APIDefinition, error)
	ListForBundle(ctx context.Context, tenantID, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	CreateMany(ctx context.Context, item []*model.APIDefinition) error
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
	DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) error
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateForAPI(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	CreateForEvent(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	ListForAPI(ctx context.Context, apiID string) ([]*model.Spec, error)
	Update(ctx context.Context, id string, in model.SpecInput) error
	Delete(ctx context.Context, id string) error
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string
}

type service struct {
	repo                APIRepository
	fetchRequestRepo    FetchRequestRepository
	uidService          UIDService
	fetchRequestService FetchRequestService
	specSvc             SpecService
	timestampGen        timestamp.Generator
}

func NewService(repo APIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService, specSvc SpecService) *service {
	return &service{repo: repo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		specSvc:             specSvc,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
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

func (s *service) GetByConditions(ctx context.Context, conds repo.Conditions) (*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	api, err := s.repo.GetByConditions(ctx, tnt, conds)
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

func (s *service) CreateInBundle(ctx context.Context, bundleID string, in model.APIDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	api := in.ToAPIDefinitionWithinBundle(bundleID, tnt)

	specs := make([]*model.SpecInput, 0, 0)
	for _, spec := range in.Specs {
		specs = append(specs, spec.ToSpec())
	}
	api.Specs = nil

	err = s.repo.Create(ctx, api)
	if err != nil {
		return "", err
	}

	for _, spec := range specs {
		_, err = s.specSvc.CreateForAPI(ctx, in.ID, *spec)
		if err != nil {
			return "", errors.Wrapf(err, "error creating spec for api in bundle with id %s", bundleID)
		}
	}

	return in.ID, nil
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

	specs, err := s.specSvc.ListForAPI(ctx, api.ID)
	if err != nil {
		return err
	}

	for _, spec := range specs {
		err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
		if err != nil {
			return errors.Wrapf(err, "while deleting FetchRequest for APIDefinition %s", id)
		}
		if err := s.specSvc.Delete(ctx, spec.ID); err != nil {
			return err
		}
	}

	in.ID = id
	api = in.ToAPIDefinitionWithinBundle(api.BundleID, tnt)

	newSpecs := make([]*model.SpecInput, 0, 0)
	for _, spec := range in.Specs {
		newSpecs = append(newSpecs, spec.ToSpec())
	}
	api.Specs = nil

	for _, spec := range newSpecs {
		_, err = s.specSvc.CreateForAPI(ctx, in.ID, *spec)
		if err != nil {
			return errors.Wrapf(err, "error creating spec for api with id %s", id)
		}
	}

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

func (s *service) RefetchAPISpecs(ctx context.Context, id string) ([]*model.APISpec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	api, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	specs, err := s.specSvc.ListForAPI(ctx, api.ID)
	if err != nil {
		return nil, err
	}

	for _, spec := range specs {
		fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", id)
		}

		if fetchRequest != nil {
			spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fetchRequest)
		}
		err = s.specSvc.Update(ctx, spec.ID, model.SpecInput{
			ID:     spec.ID,
			Tenant: spec.Tenant,
			Data:   spec.Data,
			Format: spec.Format,
			Type:   spec.Type,
		})

		if err != nil {
			return nil, errors.Wrap(err, "while updating api spec")
		}
	}

	apiSpecs := make([]*model.APISpec, 0, 0)
	for _, spec := range specs {
		apiSpecs = append(apiSpecs, spec.ToAPISpec())
	}

	return apiSpecs, nil
}

func (s *service) GetFetchRequests(ctx context.Context, apiDefID string) ([]*model.FetchRequest, error) {
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

	specs, err := s.specSvc.ListForAPI(ctx, apiDefID)
	if err != nil {
		return nil, err
	}

	fetchRequests := make([]*model.FetchRequest, 0, 0)

	for _, spec := range specs {
		fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return nil, nil
			}
			return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", apiDefID)
		}
		fetchRequests = append(fetchRequests, fetchRequest)
	}

	return fetchRequests, nil
}

func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, err
	}
	return s.repo.Exists(ctx, tnt, id)
}

func (s *service) ExistsByCondition(ctx context.Context, conds repo.Conditions) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, err
	}
	return s.repo.ExistsByCondition(ctx, tnt, conds)
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in model.FetchRequestInput, parentObjectID string) (*model.FetchRequest, error) {
	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.APIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.APIFetchRequestReference, parentObjectID)
	}

	return fr, nil
}
