package eventdef

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	GetByID(ctx context.Context, tenantID string, id string) (*model.EventDefinition, error)
	GetForPackage(ctx context.Context, tenant string, id string, packageID string) (*model.EventDefinition, error)
	Exists(ctx context.Context, tenantID, id string) (bool, error)
	ListForPackage(ctx context.Context, tenantID string, packageID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	Create(ctx context.Context, item *model.EventDefinition) error
	CreateMany(ctx context.Context, items []*model.EventDefinition) error
	Update(ctx context.Context, item *model.EventDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
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
	eventAPIRepo     EventAPIRepository
	fetchRequestRepo FetchRequestRepository
	uidService       UIDService
	timestampGen     timestamp.Generator
}

func NewService(eventAPIRepo EventAPIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{eventAPIRepo: eventAPIRepo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
		timestampGen:     timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.eventAPIRepo.ListForPackage(ctx, tnt, packageID, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.EventDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventAPI, err := s.eventAPIRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return eventAPI, nil
}

func (s *service) GetForPackage(ctx context.Context, id string, packageID string) (*model.EventDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventAPI, err := s.eventAPIRepo.GetForPackage(ctx, tnt, id, packageID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API definition")
	}

	return eventAPI, nil
}

func (s *service) CreateInPackage(ctx context.Context, packageID string, in model.EventDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()

	eventAPI := in.ToEventDefinitionWithinPackage(id, packageID, tnt)

	err = s.eventAPIRepo.Create(ctx, eventAPI)
	if err != nil {
		return "", err
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return "", errors.Wrapf(err, "while creating FetchRequest for EventDefinition with id %s", id)
		}
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventDefinitionInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	eventAPI, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.EventAPIFetchRequestReference, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting FetchRequest for EventDefinition with id %s", id)
	}

	if in.Spec != nil && in.Spec.FetchRequest != nil {
		_, err = s.createFetchRequest(ctx, tnt, in.Spec.FetchRequest, id)
		if err != nil {
			return errors.Wrapf(err, "while creating FetchRequest for EventDefinition with id %s", id)
		}
	}

	eventAPI = in.ToEventDefinitionWithinPackage(id, eventAPI.PackageID, tnt)

	err = s.eventAPIRepo.Update(ctx, eventAPI)
	if err != nil {
		return errors.Wrapf(err, "while updating EventDefinition with id %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.eventAPIRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventDefinition with id %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpec(ctx context.Context, id string) (*model.EventSpec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	eventAPI, err := s.eventAPIRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return eventAPI.Spec, nil
}

func (s *service) GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	exists, err := s.eventAPIRepo.Exists(ctx, tnt, eventAPIDefID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking if Event Definition exists")
	}
	if !exists {
		return nil, fmt.Errorf("Event Definition with ID %s doesn't exist", eventAPIDefID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.EventAPIFetchRequestReference, eventAPIDefID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by Event Definition ID %s", eventAPIDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, parentObjectID string) (*string, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.EventAPIFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.EventAPIFetchRequestReference, parentObjectID)
	}

	return &id, nil
}
