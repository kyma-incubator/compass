package spec

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=SpecRepository -output=automock -outpkg=automock -case=underscore
type SpecRepository interface {
	Create(ctx context.Context, item *model.Spec) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.Spec, error)
	ListByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error)
	Delete(ctx context.Context, tenant, id string) error
	DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) error
	Update(ctx context.Context, item *model.Spec) error
	Exists(ctx context.Context, tenantID, id string) (bool, error)
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

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleSpec(ctx context.Context, fr *model.FetchRequest) *string
}

type service struct {
	repo                SpecRepository
	fetchRequestRepo    FetchRequestRepository
	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(repo SpecRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{
		repo:                repo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) ListByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByReferenceObjectID(ctx, tnt, objectType, objectID)
}

// Until now APIs and Events had embedded specification in them, we will model this behavior by relying that the first created spec is the one which GraphQL expects
func (s *service) GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	specs, err := s.repo.ListByReferenceObjectID(ctx, tnt, objectType, objectID)
	if err != nil {
		return nil, err
	}

	if len(specs) > 0 {
		return specs[0], nil
	}

	return nil, nil
}

func (s *service) CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	spec, err := in.ToSpec(id, tnt, objectType, objectID)
	if err != nil {
		return "", err
	}

	err = s.repo.Create(ctx, spec)
	if err != nil {
		return "", errors.Wrapf(err, "while creating spec for %q with id %q", objectType, objectID)
	}

	if in.Data == nil && in.FetchRequest != nil {
		fr, err := s.createFetchRequest(ctx, tnt, *in.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for %s Specification with id %q", objectType, id)
		}

		spec.Data = s.fetchRequestService.HandleSpec(ctx, fr)

		err = s.repo.Update(ctx, spec)
		if err != nil {
			return "", errors.Wrapf(err, "while updating %s Specification with id %q", objectType, id)
		}
	}

	return id, nil
}

func (s *service) UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	spec, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return err
	}

	err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting FetchRequest for Specification with id %q", id)
	}

	spec, err = in.ToSpec(id, tnt, objectType, objectID)
	if err != nil {
		return err
	}

	if in.Data == nil && in.FetchRequest != nil {
		fr, err := s.createFetchRequest(ctx, tnt, *in.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for %s Specification with id %q", objectType, id)
		}

		spec.Data = s.fetchRequestService.HandleSpec(ctx, fr)
	}

	err = s.repo.Update(ctx, spec)
	if err != nil {
		return errors.Wrapf(err, "while updating %s Specification with id %q", objectType, id)
	}

	return nil
}

func (s *service) DeleteByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteByReferenceObjectID(ctx, tnt, objectType, objectID)
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Specification with id %q", id)
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
		return nil, errors.Wrapf(err, "while getting FetchRequest for Specification with id %q", id)
	}

	if fetchRequest != nil {
		spec.Data = s.fetchRequestService.HandleSpec(ctx, fetchRequest)
	}

	err = s.repo.Update(ctx, spec)
	if err != nil {
		return nil, errors.Wrapf(err, "while updating Specification with id %q", id)
	}

	return spec, nil
}

func (s *service) GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.Exists(ctx, tnt, specID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking if Specification with id %q exists", specID)
	}
	if !exists {
		return nil, fmt.Errorf("specification with id %q doesn't exist", specID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, specID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting FetchRequest by Specification with id %q", specID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in model.FetchRequestInput, parentObjectID string) (*model.FetchRequest, error) {
	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.SpecFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %q with id %q", model.SpecFetchRequestReference, parentObjectID)
	}

	return fr, nil
}
