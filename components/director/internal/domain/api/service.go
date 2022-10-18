package api

import (
	"context"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

// APIRepository is responsible for the repo-layer APIDefinition operations.
//go:generate mockery --name=APIRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.APIDefinition, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, counts map[string]int, pageSize int, cursor string) ([]*model.APIDefinitionPage, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.APIDefinition, error)
	CreateMany(ctx context.Context, tenant string, item []*model.APIDefinition) error
	Create(ctx context.Context, tenant string, item *model.APIDefinition) error
	Update(ctx context.Context, tenant string, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error
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
	repo                   APIRepository
	uidService             UIDService
	specService            SpecService
	bundleReferenceService BundleReferenceService
	timestampGen           timestamp.Generator
}

// NewService returns a new object responsible for service-layer APIDefinition operations.
func NewService(repo APIRepository, uidService UIDService, specService SpecService, bundleReferenceService BundleReferenceService) *service {
	return &service{
		repo:                   repo,
		uidService:             uidService,
		specService:            specService,
		bundleReferenceService: bundleReferenceService,
		timestampGen:           timestamp.DefaultGenerator,
	}
}

// ListByBundleIDs lists all APIDefinitions in pages for a given array of bundle IDs.
func (s *service) ListByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.APIDefinitionPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 600 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	bundleRefs, counts, err := s.bundleReferenceService.ListByBundleIDs(ctx, model.BundleAPIReference, bundleIDs, pageSize, cursor)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByBundleIDs(ctx, tnt, bundleIDs, bundleRefs, counts, pageSize, cursor)
}

// ListByApplicationID lists all APIDefinitions for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByApplicationID(ctx, tnt, appID)
}

// Get returns the APIDefinition by its ID.
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

// GetForBundle returns an APIDefinition by its ID and a bundle ID.
func (s *service) GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	apiDefinition, err := s.repo.GetForBundle(ctx, tnt, id, bundleID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting API definition with id %q", id)
	}

	return apiDefinition, nil
}

// CreateInBundle creates an APIDefinition. This function is used in the graphQL flow.
func (s *service) CreateInBundle(ctx context.Context, appID, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error) {
	return s.Create(ctx, appID, &bundleID, nil, in, []*model.SpecInput{spec}, nil, 0, "")
}

// Create creates APIDefinition/s. This function is used both in the ORD scenario and is re-used in CreateInBundle but with "null" ORD specific arguments.
func (s *service) Create(ctx context.Context, appID string, bundleID, packageID *string, in model.APIDefinitionInput, specs []*model.SpecInput, defaultTargetURLPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	api := in.ToAPIDefinition(id, appID, packageID, apiHash)

	if len(specs) > 0 && specs[0] != nil && specs[0].APIType != nil {
		switch *specs[0].APIType {
		case model.APISpecTypeOdata:
			protocol := ord.APIProtocolODataV2
			api.APIProtocol = &protocol
		case model.APISpecTypeOpenAPI:
			protocol := ord.APIProtocolRest
			api.APIProtocol = &protocol
		}
	}

	if err = s.repo.Create(ctx, tnt, api); err != nil {
		return "", errors.Wrap(err, "while creating api")
	}

	for _, spec := range specs {
		if spec == nil {
			continue
		}

		if _, err = s.specService.CreateByReferenceObjectID(ctx, *spec, model.APISpecReference, api.ID); err != nil {
			return "", err
		}
	}

	// when defaultTargetURLPerBundle == nil we are in the graphQL flow
	if defaultTargetURLPerBundle == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetURLFromJSONArray(in.TargetURLs)),
		}
		if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, bundleID); err != nil {
			return "", err
		}
	} else {
		for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundle {
			bundleRefInput := &model.BundleReferenceInput{
				APIDefaultTargetURL: &defaultTargetURL,
			}
			if defaultBundleID != "" && crrBndlID == defaultBundleID {
				isDefaultBundle := true
				bundleRefInput.IsDefaultBundle = &isDefaultBundle
			}
			if err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, &crrBndlID); err != nil {
				return "", err
			}
		}
	}

	return id, nil
}

// Update updates an APIDefinition. This function is used in the graphQL flow.
func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput) error {
	return s.UpdateInManyBundles(ctx, id, in, specIn, nil, nil, nil, 0, "")
}

// UpdateInManyBundles updates APIDefinition/s. This function is used both in the ORD scenario and is re-used in Update but with "null" ORD specific arguments.
func (s *service) UpdateInManyBundles(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundleForUpdate map[string]string, defaultTargetURLPerBundleForCreation map[string]string, bundleIDsForDeletion []string, apiHash uint64, defaultBundleID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	api, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	api = in.ToAPIDefinition(id, api.ApplicationID, api.PackageID, apiHash)

	if err = s.repo.Update(ctx, tnt, api); err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with id %s", id)
	}

	// when defaultTargetURLPerBundle == nil we are in the graphQL flow
	if defaultTargetURLPerBundleForUpdate == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetURLFromJSONArray(in.TargetURLs)),
		}
		err = s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, nil)
		if err != nil {
			return err
		}
	} else {
		err = s.updateBundleReferences(ctx, api, defaultTargetURLPerBundleForUpdate, defaultBundleID)
		if err != nil {
			return err
		}
	}

	err = s.createBundleReferences(ctx, api, defaultTargetURLPerBundleForCreation, defaultBundleID)
	if err != nil {
		return err
	}

	err = s.deleteBundleIDs(ctx, &api.ID, bundleIDsForDeletion)
	if err != nil {
		return err
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

// Delete deletes the APIDefinition by its ID.
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

// DeleteAllByBundleID deletes all APIDefinitions for a given bundle ID
func (s *service) DeleteAllByBundleID(ctx context.Context, bundleID string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.DeleteAllByBundleID(ctx, tnt, bundleID)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIDefinitions for Bundle with id %q", bundleID)
	}

	return nil
}

// ListFetchRequests lists all FetchRequests for given specification IDs
func (s *service) ListFetchRequests(ctx context.Context, specIDs []string) ([]*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fetchRequests, err := s.specService.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDs, model.APISpecReference)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return fetchRequests, nil
}

func (s *service) updateBundleReferences(ctx context.Context, api *model.APIDefinition, defaultTargetURLPerBundleForUpdate map[string]string, defaultBundleID string) error {
	for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundleForUpdate {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: &defaultTargetURL,
		}
		if defaultBundleID != "" && defaultBundleID == crrBndlID {
			isDefaultBundle := true
			bundleRefInput.IsDefaultBundle = &isDefaultBundle
		}

		err := s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, &crrBndlID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) createBundleReferences(ctx context.Context, api *model.APIDefinition, defaultTargetURLPerBundleForCreation map[string]string, defaultBundleID string) error {
	for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundleForCreation {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: &defaultTargetURL,
		}
		if defaultBundleID != "" && crrBndlID == defaultBundleID {
			isDefaultBundle := true
			bundleRefInput.IsDefaultBundle = &isDefaultBundle
		}

		err := s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, &crrBndlID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) deleteBundleIDs(ctx context.Context, apiID *string, bundleIDsForDeletion []string) error {
	for _, bundleID := range bundleIDsForDeletion {
		err := s.bundleReferenceService.DeleteByReferenceObjectID(ctx, model.BundleAPIReference, apiID, &bundleID)
		if err != nil {
			return err
		}
	}
	return nil
}
