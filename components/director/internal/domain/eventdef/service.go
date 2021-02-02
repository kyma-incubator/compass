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
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.EventDefinition, error)
	Exists(ctx context.Context, tenantID, id string) (bool, error)
	ListForBundle(ctx context.Context, tenantID string, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	Create(ctx context.Context, item *model.EventDefinition) error
	CreateMany(ctx context.Context, items []*model.EventDefinition) error
	Update(ctx context.Context, item *model.EventDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
	GetByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=SpecService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	ListByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
}

type service struct {
	eventAPIRepo        EventAPIRepository
	fetchRequestRepo    FetchRequestRepository
	uidService          UIDService
	specService         SpecService
	timestampGen        timestamp.Generator
}

func NewService(eventAPIRepo EventAPIRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, specService SpecService) *service {
	return &service{
		eventAPIRepo:        eventAPIRepo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		specService:         specService,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.eventAPIRepo.ListForBundle(ctx, tnt, bundleID, pageSize, cursor)
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

func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.EventDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventAPI, err := s.eventAPIRepo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting API definition")
	}

	return eventAPI, nil
}

func (s *service) CreateInBundle(ctx context.Context, bundleID string, in model.EventDefinitionInput, spec model.SpecInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()
	eventAPI := in.ToEventDefinitionWithinBundle(id, bundleID, tnt)

	err = s.eventAPIRepo.Create(ctx, eventAPI)
	if err != nil {
		return "", err
	}

	_, err = s.specService.CreateByReferenceObjectID(ctx, spec, model.EventSpecReference, eventAPI.ID)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventDefinitionInput, spec model.SpecInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	event, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	event = in.ToEventDefinitionWithinBundle(id, event.BundleID, tnt)

	err = s.eventAPIRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrapf(err, "while updating EventDefinition with id %s", id)
	}

	specs, err := s.specService.ListByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
	if err != nil {
		return errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
	}

	err = s.specService.UpdateByReferenceObjectID(ctx, specs[0].ID, spec, model.EventSpecReference, event.ID)
	if err != nil {
		return err
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
		return nil, fmt.Errorf("event definition with id %s doesn't exist", eventAPIDefID)
	}

	specs, err := s.specService.ListByReferenceObjectID(ctx, model.EventSpecReference, eventAPIDefID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting spec for EventDefinition with id %q", eventAPIDefID)
	}

	fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, specs[0].ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while getting FetchRequest by Event Definition with id %q", eventAPIDefID)
	}

	return fetchRequest, nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, parentObjectID string) (*model.FetchRequest, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, model.SpecFetchRequestReference, parentObjectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", model.SpecFetchRequestReference, parentObjectID)
	}

	return fr, nil
}
