package eventdef

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

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
	ExistsByCondition(ctx context.Context, tenant string, conds repo.Conditions) (bool, error)
	GetByConditions(ctx context.Context, tenant string, conds repo.Conditions) (*model.EventDefinition, error)
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
	DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.FetchRequestReferenceObjectType, objectID string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type SpecService interface {
	CreateForAPI(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	CreateForEvent(ctx context.Context, bundleID string, in model.SpecInput) (string, error)
	ListForEvent(ctx context.Context, eventID string) ([]*model.Spec, error)
	Update(ctx context.Context, id string, in model.SpecInput) error
	Delete(ctx context.Context, id string) error
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string
}

//go:generate mockery -name=SpecConverter -output=automock -outpkg=automock -case=underscore
type SpecConverter interface {
	EventSpecInputFromSpec(spec *model.Spec, fr *model.FetchRequest) *model.EventSpecInput
}

type service struct {
	eventAPIRepo     EventAPIRepository
	fetchRequestRepo FetchRequestRepository
	fetchRequestSvc  FetchRequestService
	specSvc          SpecService
	specConverter    SpecConverter
	uidService       UIDService
	timestampGen     timestamp.Generator
}

func NewService(eventAPIRepo EventAPIRepository, fetchRequestRepo FetchRequestRepository, fetchRequestSvc FetchRequestService, specSvc SpecService, specConverter SpecConverter, uidService UIDService) *service {
	return &service{eventAPIRepo: eventAPIRepo,
		fetchRequestRepo: fetchRequestRepo,
		fetchRequestSvc:  fetchRequestSvc,
		uidService:       uidService,
		specSvc:          specSvc,
		specConverter:    specConverter,
		timestampGen:     timestamp.DefaultGenerator(),
	}
}

func (s *service) ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.EventDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
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

func (s *service) GetByConditions(ctx context.Context, conds repo.Conditions) (*model.EventDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	event, err := s.eventAPIRepo.GetByConditions(ctx, tnt, conds)
	if err != nil {
		return nil, err
	}

	return event, nil
}

func (s *service) ExistsByCondition(ctx context.Context, conds repo.Conditions) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, err
	}
	return s.eventAPIRepo.ExistsByCondition(ctx, tnt, conds)
}

func (s *service) CreateInBundle(ctx context.Context, bundleID string, in model.EventDefinitionInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	event := in.ToEventDefinitionWithinBundle(bundleID, tnt)

	specs := make([]*model.SpecInput, 0, 0)
	for _, spec := range in.Specs {
		specs = append(specs, spec.ToSpec())
	}
	event.Specs = nil

	err = s.eventAPIRepo.Create(ctx, event)
	if err != nil {
		return "", err
	}

	for _, spec := range specs {
		_, err = s.specSvc.CreateForEvent(ctx, in.ID, *spec)
		if err != nil {
			return "", errors.Wrapf(err, "error creating spec for event in bundle with id %s", bundleID)
		}
	}

	return in.ID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.EventDefinitionInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	event, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	specs, err := s.specSvc.ListForEvent(ctx, event.ID)
	if err != nil {
		return err
	}

	if len(in.Specs) > 0 {
		for _, spec := range specs {
			err = s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
			if err != nil {
				return errors.Wrapf(err, "while deleting FetchRequest for APIDefinition %s", id)
			}
			if err := s.specSvc.Delete(ctx, spec.ID); err != nil {
				return err
			}
		}
	} else {
		for _, spec := range specs {
			fr, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
			if err != nil {
				return err
			}
			if fr.Status.Condition == model.FetchRequestStatusConditionFailed {
				if err := s.fetchRequestRepo.DeleteByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID); err != nil {
					return err
				}
				in.Specs = append(in.Specs, s.specConverter.EventSpecInputFromSpec(spec, fr))
				if err := s.specSvc.Delete(ctx, spec.ID); err != nil {
					return err
				}
			}
		}
	}

	in.ID = id
	event = in.ToEventDefinitionWithinBundle(event.BundleID, tnt)

	newSpecs := make([]*model.SpecInput, 0, 0)
	for _, spec := range in.Specs {
		newSpecs = append(newSpecs, spec.ToSpec())
	}
	event.Specs = nil

	for _, spec := range newSpecs {
		_, err = s.specSvc.CreateForEvent(ctx, in.ID, *spec)
		if err != nil {
			return errors.Wrapf(err, "error creating spec for event with id %s", id)
		}
	}

	err = s.eventAPIRepo.Update(ctx, event)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with ID %s", id)
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
		return errors.Wrapf(err, "while deleting EventDefinition with ID %s", id)
	}

	return nil
}

func (s *service) RefetchAPISpecs(ctx context.Context, id string) ([]*model.EventSpec, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	api, err := s.eventAPIRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	specs, err := s.specSvc.ListForEvent(ctx, api.ID)
	if err != nil {
		return nil, err
	}

	for _, spec := range specs {
		fetchRequest, err := s.fetchRequestRepo.GetByReferenceObjectID(ctx, tnt, model.SpecFetchRequestReference, spec.ID)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting FetchRequest by API Definition ID %s", id)
		}

		if fetchRequest != nil {
			spec.Data = s.fetchRequestSvc.HandleAPISpec(ctx, fetchRequest)
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

	apiSpecs := make([]*model.EventSpec, 0, 0)
	for _, spec := range specs {
		apiSpecs = append(apiSpecs, spec.ToEventSpec())
	}

	return apiSpecs, nil
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
