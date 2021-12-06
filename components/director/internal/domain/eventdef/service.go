package eventdef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// EventAPIRepository missing godoc
//go:generate mockery --name=EventAPIRepository --output=automock --outpkg=automock --case=underscore
type EventAPIRepository interface {
	GetByID(ctx context.Context, tenantID string, id string) (*model.EventDefinition, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.EventDefinition, error)
	ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, totalCounts map[string]int, pageSize int, cursor string) ([]*model.EventDefinitionPage, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.EventDefinition, error)
	Create(ctx context.Context, tenant string, item *model.EventDefinition) error
	Update(ctx context.Context, tenant string, item *model.EventDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

// SpecService missing godoc
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
}

// BundleReferenceService missing godoc
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore
type BundleReferenceService interface {
	GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error)
	CreateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	UpdateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	DeleteByReferenceObjectID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error)
}

type service struct {
	eventAPIRepo           EventAPIRepository
	uidService             UIDService
	specService            SpecService
	bundleReferenceService BundleReferenceService
	timestampGen           timestamp.Generator
}

// NewService missing godoc
func NewService(eventAPIRepo EventAPIRepository, uidService UIDService, specService SpecService, bundleReferenceService BundleReferenceService) *service {
	return &service{
		eventAPIRepo:           eventAPIRepo,
		uidService:             uidService,
		specService:            specService,
		bundleReferenceService: bundleReferenceService,
		timestampGen:           timestamp.DefaultGenerator,
	}
}

// ListByBundleIDs missing godoc
func (s *service) ListByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.EventDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	bundleRefs, counts, err := s.bundleReferenceService.ListByBundleIDs(ctx, model.BundleEventReference, bundleIDs, pageSize, cursor)
	if err != nil {
		return nil, err
	}

	return s.eventAPIRepo.ListByBundleIDs(ctx, tnt, bundleIDs, bundleRefs, counts, pageSize, cursor)
}

// ListByApplicationID missing godoc
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.eventAPIRepo.ListByApplicationID(ctx, tnt, appID)
}

// Get missing godoc
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

// GetForBundle missing godoc
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

// CreateInBundle missing godoc
func (s *service) CreateInBundle(ctx context.Context, appID, bundleID string, in model.EventDefinitionInput, spec *model.SpecInput) (string, error) {
	return s.Create(ctx, appID, &bundleID, nil, in, []*model.SpecInput{spec}, nil, 0)
}

// Create missing godoc
func (s *service) Create(ctx context.Context, appID string, bundleID, packageID *string, in model.EventDefinitionInput, specs []*model.SpecInput, bundleIDs []string, eventHash uint64) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()
	eventAPI := in.ToEventDefinition(id, appID, packageID, eventHash)

	if err = s.eventAPIRepo.Create(ctx, tnt, eventAPI); err != nil {
		return "", err
	}

	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if _, err = s.specService.CreateByReferenceObjectID(ctx, *spec, model.EventSpecReference, eventAPI.ID); err != nil {
			return "", err
		}
	}

	if bundleIDs == nil {
		if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, model.BundleReferenceInput{}, model.BundleEventReference, &eventAPI.ID, bundleID); err != nil {
			return "", err
		}
	} else {
		for _, bndlID := range bundleIDs {
			if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, model.BundleReferenceInput{}, model.BundleEventReference, &eventAPI.ID, &bndlID); err != nil {
				return "", err
			}
		}
	}

	return id, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput) error {
	return s.UpdateInManyBundles(ctx, id, in, specIn, nil, nil, 0)
}

// UpdateInManyBundles missing godoc
func (s *service) UpdateInManyBundles(ctx context.Context, id string, in model.EventDefinitionInput, specIn *model.SpecInput, bundleIDsForCreation []string, bundleIDsForDeletion []string, eventHash uint64) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	event, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	event = in.ToEventDefinition(id, event.ApplicationID, event.PackageID, eventHash)

	if err = s.eventAPIRepo.Update(ctx, tnt, event); err != nil {
		return errors.Wrapf(err, "while updating EventDefinition with id %s", id)
	}

	for _, bundleID := range bundleIDsForCreation {
		if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, model.BundleReferenceInput{}, model.BundleEventReference, &event.ID, &bundleID); err != nil {
			return err
		}
	}

	for _, bundleID := range bundleIDsForDeletion {
		if err = s.bundleReferenceService.DeleteByReferenceObjectID(ctx, model.BundleEventReference, &event.ID, &bundleID); err != nil {
			return err
		}
	}

	if specIn != nil {
		dbSpec, err := s.specService.GetByReferenceObjectID(ctx, model.EventSpecReference, event.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting spec for EventDefinition with id %q", event.ID)
		}

		if dbSpec == nil {
			_, err = s.specService.CreateByReferenceObjectID(ctx, *specIn, model.EventSpecReference, event.ID)
			return err
		}

		return s.specService.UpdateByReferenceObjectID(ctx, dbSpec.ID, *specIn, model.EventSpecReference, event.ID)
	}

	return nil
}

// Delete missing godoc
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

// DeleteAllByBundleID missing godoc
func (s *service) DeleteAllByBundleID(ctx context.Context, bundleID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	err = s.eventAPIRepo.DeleteAllByBundleID(ctx, tnt, bundleID)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventDefinitions for Bundle with id %q", bundleID)
	}

	return nil
}

// ListFetchRequests missing godoc
func (s *service) ListFetchRequests(ctx context.Context, specIDs []string) ([]*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fetchRequests, err := s.specService.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDs, model.EventSpecReference)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return fetchRequests, nil
}
