package spec

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateForAPI(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	CreateForEvent(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	ListForAPI(ctx context.Context, apiID string) ([]*model.Spec, error)
	ListForEvent(ctx context.Context, eventID string) ([]*model.Spec, error)
	Update(ctx context.Context, id string, in model.SpecInput) error
	Delete(ctx context.Context, id string) error
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=SpecRepository -output=automock -outpkg=automock -case=underscore
type SpecRepository interface {
	ListForAPI(ctx context.Context, tenantID, apiID string) ([]*model.Spec, error)
	ListForEvent(ctx context.Context, tenantID, eventID string) ([]*model.Spec, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenantID, id string) (*model.Spec, error)
	CreateMany(ctx context.Context, item []*model.Spec) error
	Create(ctx context.Context, item *model.Spec) error
	Update(ctx context.Context, item *model.Spec) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string
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
	repo                SpecRepository
	fetchRequestRepo    FetchRequestRepository
	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(repo SpecRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{repo: repo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForAPI(ctx context.Context, apiID string) ([]*model.Spec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListForAPI(ctx, tnt, apiID)
}

func (s *service) ListForEvent(ctx context.Context, eventID string) ([]*model.Spec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListForEvent(ctx, tnt, eventID)
}

func (s *service) CreateForAPI(ctx context.Context, apiID string, in model.SpecInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	spec := in.ToSpecWithinAPI(apiID, tnt)

	err = s.repo.Create(ctx, spec)
	if err != nil {
		return "", err
	}

	if in.FetchRequest != nil {
		fr, err := s.createFetchRequest(ctx, tnt, *in.FetchRequest, in.ID)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", in.ID)
		}

		spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fr)

		err = s.repo.Update(ctx, spec)
		if err != nil {
			return "", errors.Wrap(err, "while updating spec with spec spec")
		}
	}

	return in.ID, nil
}

func (s *service) CreateForEvent(ctx context.Context, eventID string, in model.SpecInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	spec := in.ToSpecWithinEvent(eventID, tnt)

	err = s.repo.Create(ctx, spec)
	if err != nil {
		return "", err
	}

	if in.FetchRequest != nil {
		fr, err := s.createFetchRequest(ctx, tnt, *in.FetchRequest, in.ID)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", in.ID)
		}

		spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fr)

		err = s.repo.Update(ctx, spec)
		if err != nil {
			return "", errors.Wrap(err, "while updating spec with spec spec")
		}
	}

	return in.ID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.SpecInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	spec, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return err
	}

	err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, id) // TODO
	if err != nil {
		return errors.Wrapf(err, "while deleting FetchRequest for APIDefinition %s", id)
	}

	in.ID = id
	if spec.APIDefinitionID != nil && len(*spec.APIDefinitionID) > 0 {
		spec = in.ToSpecWithinAPI(*spec.APIDefinitionID, tnt)
	} else {
		spec = in.ToSpecWithinEvent(*spec.EventDefinitionID, tnt)
	}

	if in.Data != nil && in.FetchRequest != nil {
		fr, err := s.createFetchRequest(ctx, tnt, *in.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for APIDefinition %s", id)
		}

		spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fr)
	}

	err = s.repo.Update(ctx, spec)
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

func (s *service) RefetchSpec(ctx context.Context, id string) (*model.Spec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	spec, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, id)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", id)
	}

	if fetchRequest != nil {
		spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fetchRequest)
	}

	err = s.repo.Update(ctx, spec)
	if err != nil {
		return nil, errors.Wrap(err, "while updating spec with spec spec")
	}

	return spec, nil
}

func (s *service) GetFetchRequest(ctx context.Context, spec string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.Exists(ctx, tnt, spec)
	if err != nil {
		return nil, errors.Wrap(err, "while checking if API Definition exists")
	}
	if !exists {
		return nil, fmt.Errorf("API Definition with ID %s doesn't exist", spec)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", spec)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in model.FetchRequestInput, parentObjectID string) (*model.FetchRequest, error) {
	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.SpecFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.SpecFetchRequestReference, parentObjectID)
	}

	return fr, nil
}
