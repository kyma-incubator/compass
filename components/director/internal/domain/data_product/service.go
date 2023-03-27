package data_product

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/port"
	"github.com/kyma-incubator/compass/components/director/internal/domain/portapiref"
	"github.com/kyma-incubator/compass/components/director/internal/domain/porteventref"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

// DataProductRepository is responsible for the repo-layer Data Product operations.
//go:generate mockery --name=DataProductRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataProductRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.DataProduct, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.DataProduct, error)
	Create(ctx context.Context, tenant string, item *model.DataProduct) error
	Update(ctx context.Context, tenant string, item *model.DataProduct) error
}

// PortRepository is responsible for the repo-layer Port operations.
//go:generate mockery --name=PortRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type PortRepository interface {
	Create(ctx context.Context, tenant string, item *model.Port) error
}

// PortApiRefRepository is responsible for the repo-layer PortApiRef operations.
//go:generate mockery --name=PortApiRefRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type PortApiRefRepository interface {
	Create(ctx context.Context, tenant string, id, appID, portID, apiID string, minVersion *string) error
}

// PortEventRefRepository is responsible for the repo-layer PortEventRef operations.
//go:generate mockery --name=PortEventRefRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type PortEventRefRepository interface {
	Create(ctx context.Context, tenant string, id, appID, portID, eventID string, minVersion *string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal apiDefinition IDs when they are created.
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// SpecService is responsible for the service-layer Specification operations.
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
}

// BundleReferenceService is responsible for the service-layer BundleReference operations.
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceService interface {
	GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error)
	CreateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	UpdateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	DeleteByReferenceObjectID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error
	ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error)
}

type service struct {
	repo                   DataProductRepository
	uidService             UIDService
	bundleReferenceService BundleReferenceService
	timestampGen           timestamp.Generator
	portRepo               PortRepository
	portApiRefRepo         PortApiRefRepository
	portEventRefRepo       PortEventRefRepository
}

// NewService returns a new object responsible for service-layer APIDefinition operations.
func NewService(repo DataProductRepository, uidService UIDService, bundleReferenceService BundleReferenceService) *service {
	return &service{
		repo:                   repo,
		uidService:             uidService,
		bundleReferenceService: bundleReferenceService,
		timestampGen:           timestamp.DefaultGenerator,
		portRepo:               port.NewRepository(port.NewConverter()),
		portApiRefRepo:         portapiref.NewRepository(),
		portEventRefRepo:       porteventref.NewRepository(),
	}
}

// ListByApplicationID lists all DataProducts for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.DataProduct, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByApplicationID(ctx, tnt, appID)
}

// Create creates DataProduct/s.
func (s *service) Create(ctx context.Context, appID string, packageID *string, in model.DataProductInput, bundleIDs []string, defaultBundleID string, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	// create data product
	id := s.uidService.Generate()
	dataProduct := in.ToDataProduct(id, appID, packageID)

	if err = s.repo.Create(ctx, tnt, dataProduct); err != nil {
		return "", errors.Wrap(err, "while creating data product "+id)
	}

	// create ports
	for _, inputPort := range in.InputPorts {
		if err := s.createPort(ctx, inputPort, tnt, "input", id, appID, apisFromDB, eventsFromDB); err != nil {
			return "", err
		}
	}

	for _, outputPort := range in.OutputPorts {
		if err := s.createPort(ctx, outputPort, tnt, "output", id, appID, apisFromDB, eventsFromDB); err != nil {
			return "", err
		}
	}

	for _, bndlID := range bundleIDs {
		bundleRefInput := &model.BundleReferenceInput{}
		if defaultBundleID != "" && bndlID == defaultBundleID {
			isDefaultBundle := true
			bundleRefInput = &model.BundleReferenceInput{
				IsDefaultBundle: &isDefaultBundle,
			}
		}
		if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleDataProductReference, &dataProduct.ID, &bndlID); err != nil {
			return "", err
		}
	}

	return id, nil
}

func (s *service) createPort(ctx context.Context, portInput *model.PortInput, tenant, portType, dataProductID, appID string, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition) error {
	portID := s.uidService.Generate()
	port := portInput.ToPort(portID, dataProductID, appID, portType)
	if err := s.portRepo.Create(ctx, tenant, port); err != nil {
		return errors.Wrap(err, "while creating port")
	}

	for _, api := range portInput.ApiResources {
		var apiID *string
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &api.OrdID)
		}); found {
			apiID = &apisFromDB[i].ID
		}
		if apiID == nil {
			//TODO log
			continue
		}

		apiRefID := s.uidService.Generate()
		if err := s.portApiRefRepo.Create(ctx, tenant, apiRefID, appID, portID, *apiID, api.MinVersion); err != nil {
			return errors.Wrap(err, "while creating api port ref")
		}
	}

	for _, event := range portInput.EventResources {
		var eventID *string
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &event.OrdID)
		}); found {
			eventID = &eventsFromDB[i].ID
		}
		if eventID == nil {
			//TODO log
			continue
		}

		eventRefID := s.uidService.Generate()
		if err := s.portEventRefRepo.Create(ctx, tenant, eventRefID, appID, portID, *eventID, event.MinVersion); err != nil {
			return errors.Wrap(err, "while creating event port ref")
		}
	}
	return nil
}

// Get returns the DataProduct by its ID.
func (s *service) Get(ctx context.Context, id string) (*model.DataProduct, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	dataProduct, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, err
	}

	return dataProduct, nil
}

func (s *service) UpdateInManyBundles(ctx context.Context, id string, in model.DataProductInput, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion []string, defaultBundleID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	dataProduct, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	dataProduct = in.ToDataProduct(id, dataProduct.ApplicationID, dataProduct.OrdPackageID)

	if err = s.repo.Update(ctx, tnt, dataProduct); err != nil {
		return errors.Wrapf(err, "while updating DataProduct with id %s", id)
	}

	//TODO update ports

	for _, bundleID := range bundleIDsForCreation {
		createBundleRefInput := &model.BundleReferenceInput{}
		if defaultBundleID != "" && bundleID == defaultBundleID {
			isDefaultBundle := true
			createBundleRefInput = &model.BundleReferenceInput{IsDefaultBundle: &isDefaultBundle}
		}
		if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *createBundleRefInput, model.BundleDataProductReference, &dataProduct.ID, &bundleID); err != nil {
			return err
		}
	}

	for _, bundleID := range bundleIDsForDeletion {
		if err = s.bundleReferenceService.DeleteByReferenceObjectID(ctx, model.BundleDataProductReference, &dataProduct.ID, &bundleID); err != nil {
			return err
		}
	}

	for _, bundleID := range bundleIDsFromBundleReference {
		bundleRefInput := &model.BundleReferenceInput{}
		if defaultBundleID != "" && bundleID == defaultBundleID {
			isDefaultBundle := true
			bundleRefInput = &model.BundleReferenceInput{IsDefaultBundle: &isDefaultBundle}
		}
		if err := s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleDataProductReference, &dataProduct.ID, &bundleID); err != nil {
			return err
		}
	}

	return nil
}

func equalStrings(first, second *string) bool {
	return first != nil && second != nil && *first == *second
}

func searchInSlice(length int, f func(i int) bool) (int, bool) {
	for i := 0; i < length; i++ {
		if f(i) {
			return i, true
		}
	}
	return -1, false
}
