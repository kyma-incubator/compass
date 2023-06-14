package api

import (
	"context"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

// APIRepository is responsible for the repo-layer APIDefinition operations.
//
//go:generate mockery --name=APIRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.APIDefinition, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.APIDefinition, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, counts map[string]int, pageSize int, cursor string) ([]*model.APIDefinitionPage, error)
	ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.APIDefinition, error)
	CreateMany(ctx context.Context, tenant string, item []*model.APIDefinition) error
	Create(ctx context.Context, tenant string, item *model.APIDefinition) error
	CreateGlobal(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, tenant string, item *model.APIDefinition) error
	UpdateGlobal(ctx context.Context, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error
	DeleteGlobal(ctx context.Context, id string) error
}

// UIDService is responsible for generating GUIDs, which will be used as internal apiDefinition IDs when they are created.
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// SpecService is responsible for the service-layer Specification operations.
//
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
}

// BundleReferenceService is responsible for the service-layer BundleReference operations.
//
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

	if pageSize < 1 || pageSize > 200 {
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

	return s.repo.ListByResourceID(ctx, tnt, resource.Application, appID)
}

// ListByApplicationTemplateVersionID lists all APIDefinitions for a given application ID.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.APIDefinition, error) {
	return s.repo.ListByResourceID(ctx, "", resource.ApplicationTemplateVersion, appTemplateVersionID)
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
func (s *service) CreateInBundle(ctx context.Context, resourceType resource.Type, resourceID string, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error) {
	return s.Create(ctx, resourceType, resourceID, &bundleID, nil, in, []*model.SpecInput{spec}, nil, 0, "")
}

// Create creates APIDefinition/s. This function is used both in the ORD scenario and is re-used in CreateInBundle but with "null" ORD specific arguments.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID, packageID *string, in model.APIDefinitionInput, specs []*model.SpecInput, defaultTargetURLPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error) {
	id := s.uidService.Generate()
	api := in.ToAPIDefinition(id, resourceType, resourceID, packageID, apiHash)

	enrichAPIProtocol(api, specs)

	if err := s.createAPI(ctx, api, resourceType); err != nil {
		return "", errors.Wrap(err, "while creating api")
	}

	if err := s.processSpecs(ctx, api.ID, specs, resourceType); err != nil {
		return "", errors.Wrap(err, "while processing specs")
	}

	if err := s.createBundleReferenceObject(ctx, api.ID, bundleID, defaultBundleID, api.TargetURLs, defaultTargetURLPerBundle); err != nil {
		return "", errors.Wrap(err, "while creating bundle reference object")
	}

	return id, nil
}

// Update updates an APIDefinition. This function is used in the graphQL flow.
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.APIDefinitionInput, specIn *model.SpecInput) error {
	return s.UpdateInManyBundles(ctx, resourceType, id, in, specIn, nil, nil, nil, 0, "")
}

// UpdateInManyBundles updates APIDefinition/s. This function is used both in the ORD scenario and is re-used in Update but with "null" ORD specific arguments.
func (s *service) UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundleForUpdate map[string]string, defaultTargetURLPerBundleForCreation map[string]string, bundleIDsForDeletion []string, apiHash uint64, defaultBundleID string) error {
	api, err := s.getAPI(ctx, id, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while getting API with ID %s for %s", id, resourceType)
	}

	_, resourceID := getParentResource(api)

	api = in.ToAPIDefinition(id, resourceType, resourceID, api.PackageID, apiHash)

	err = s.updateAPI(ctx, api, resourceType)
	if err != nil {
		return errors.Wrapf(err, "while updating API with ID %s for %s", id, resourceType)
	}

	if err = s.updateReferences(ctx, api, in.TargetURLs, defaultTargetURLPerBundleForUpdate, defaultBundleID); err != nil {
		return err
	}

	if err = s.createBundleReferences(ctx, api, defaultTargetURLPerBundleForCreation, defaultBundleID); err != nil {
		return err
	}

	if err = s.deleteBundleIDs(ctx, &api.ID, bundleIDsForDeletion); err != nil {
		return err
	}

	if specIn != nil {
		return s.handleSpecsInAPI(ctx, api.ID, specIn, resourceType)
	}

	return nil
}

// Delete deletes the APIDefinition by its ID.
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	if err := s.deleteAPI(ctx, id, resourceType); err != nil {
		return errors.Wrapf(err, "while deleting APIDefinition with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted APIDefinition with id %s", id)

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

func (s *service) handleSpecsInAPI(ctx context.Context, id string, specIn *model.SpecInput, resourceType resource.Type) error {
	dbSpec, err := s.specService.GetByReferenceObjectID(ctx, resourceType, model.APISpecReference, id)
	if err != nil {
		return errors.Wrapf(err, "while getting spec for APIDefinition with id %q", id)
	}

	if dbSpec == nil {
		_, err = s.specService.CreateByReferenceObjectID(ctx, *specIn, resourceType, model.APISpecReference, id)
		return err
	}

	return s.specService.UpdateByReferenceObjectID(ctx, dbSpec.ID, *specIn, resourceType, model.APISpecReference, id)
}

func (s *service) updateReferences(ctx context.Context, api *model.APIDefinition, targetURLs json.RawMessage, defaultTargetURLPerBundleForUpdate map[string]string, defaultBundleID string) error {
	// when defaultTargetURLPerBundle == nil we are in the graphQL flow
	if defaultTargetURLPerBundleForUpdate == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetURLFromJSONArray(targetURLs)),
		}
		return s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, nil)
	}

	return s.updateBundleReferences(ctx, api, defaultTargetURLPerBundleForUpdate, defaultBundleID)
}

func (s *service) processSpecs(ctx context.Context, apiID string, specs []*model.SpecInput, resourceType resource.Type) error {
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		if _, err := s.specService.CreateByReferenceObjectID(ctx, *spec, resourceType, model.APISpecReference, apiID); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) createBundleReferenceObject(ctx context.Context, apiID string, bundleID *string, defaultBundleID string, targetURLs json.RawMessage, defaultTargetURLPerBundle map[string]string) error {
	// when defaultTargetURLPerBundle == nil we are in the graphQL flow
	if defaultTargetURLPerBundle == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetURLFromJSONArray(targetURLs)),
		}
		return s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &apiID, bundleID)
	}

	for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundle {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: &defaultTargetURL,
		}
		if defaultBundleID != "" && crrBndlID == defaultBundleID {
			isDefaultBundle := true
			bundleRefInput.IsDefaultBundle = &isDefaultBundle
		}
		if err := s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &apiID, &crrBndlID); err != nil {
			return err
		}
	}

	return nil
}

func (s *service) getAPI(ctx context.Context, id string, resourceType resource.Type) (*model.APIDefinition, error) {
	if resourceType.IsTenantIgnorable() {
		return s.repo.GetByIDGlobal(ctx, id)
	}
	return s.Get(ctx, id)
}

func (s *service) updateAPI(ctx context.Context, api *model.APIDefinition, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.UpdateGlobal(ctx, api)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Update(ctx, tnt, api)
}

func (s *service) createAPI(ctx context.Context, api *model.APIDefinition, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.CreateGlobal(ctx, api)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, tnt, api)
}

func (s *service) deleteAPI(ctx context.Context, apiID string, resourceType resource.Type) error {
	if resourceType.IsTenantIgnorable() {
		return s.repo.DeleteGlobal(ctx, apiID)
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}
	return s.repo.Delete(ctx, tnt, apiID)
}

func getParentResource(api *model.APIDefinition) (resource.Type, string) {
	if api.ApplicationTemplateVersionID != nil {
		return resource.ApplicationTemplateVersion, *api.ApplicationTemplateVersionID
	} else if api.ApplicationID != nil {
		return resource.Application, *api.ApplicationID
	}

	return "", ""
}

func enrichAPIProtocol(api *model.APIDefinition, specs []*model.SpecInput) {
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
}
