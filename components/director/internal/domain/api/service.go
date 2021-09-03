package api

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery --name=APIRepository --output=automock --outpkg=automock --case=underscore
type APIRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*model.APIDefinition, error)
	GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, counts map[string]int, pageSize int, cursor string) ([]*model.APIDefinitionPage, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.APIDefinition, error)
	CreateMany(ctx context.Context, item []*model.APIDefinition) error
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
	Delete(ctx context.Context, tenantID string, id string) error
	DeleteAllByBundleID(ctx context.Context, tenantID, bundleID string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error
	GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error)
	RefetchSpec(ctx context.Context, id string) (*model.Spec, error)
	GetFetchRequest(ctx context.Context, specID string) (*model.FetchRequest, error)
	ListByReferenceObjectIDs(ctx context.Context, objectType model.SpecReferenceObjectType, objectIDs []string) ([]*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string) ([]*model.FetchRequest, error)
}

//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore
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

func NewService(repo APIRepository, uidService UIDService, specService SpecService, bundleReferenceService BundleReferenceService) *service {
	return &service{
		repo:                   repo,
		uidService:             uidService,
		specService:            specService,
		bundleReferenceService: bundleReferenceService,
		timestampGen:           timestamp.DefaultGenerator,
	}
}

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

func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.repo.ListByApplicationID(ctx, tnt, appID)
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

func (s *service) CreateInBundle(ctx context.Context, appId, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error) {
	return s.Create(ctx, appId, &bundleID, nil, in, []*model.SpecInput{spec}, nil, 0)
}

func (s *service) Create(ctx context.Context, appId string, bundleID, packageID *string, in model.APIDefinitionInput, specs []*model.SpecInput, defaultTargetURLPerBundle map[string]string, apiHash uint64) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	api := in.ToAPIDefinition(id, appId, packageID, tnt, apiHash)

	err = s.repo.Create(ctx, api)
	if err != nil {
		return "", errors.Wrap(err, "while creating api")
	}

	for _, spec := range specs {
		if spec == nil {
			continue
		}
		_, err = s.specService.CreateByReferenceObjectID(ctx, *spec, model.APISpecReference, api.ID)
		if err != nil {
			return "", err
		}
	}

	if defaultTargetURLPerBundle == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetUrlFromJsonArray(in.TargetURLs)),
		}
		err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, bundleID)
		if err != nil {
			return "", err
		}
	} else {
		for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundle {
			bundleRefInput := &model.BundleReferenceInput{
				APIDefaultTargetURL: &defaultTargetURL,
			}
			err = s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, &crrBndlID)
			if err != nil {
				return "", err
			}
		}
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput) error {
	return s.UpdateInManyBundles(ctx, id, in, specIn, nil, nil, nil, 0)
}

func (s *service) UpdateInManyBundles(ctx context.Context, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundleForUpdate map[string]string, defaultTargetURLPerBundleForCreation map[string]string, bundleIDsForDeletion []string, apiHash uint64) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	api, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	api = in.ToAPIDefinition(id, api.ApplicationID, api.PackageID, tnt, apiHash)

	err = s.repo.Update(ctx, api)
	if err != nil {
		return errors.Wrapf(err, "while updating APIDefinition with id %s", id)
	}

	if defaultTargetURLPerBundleForUpdate == nil {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: str.Ptr(ExtractTargetUrlFromJsonArray(in.TargetURLs)),
		}
		err = s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, &api.ID, nil)
		if err != nil {
			return err
		}
	} else {
		err = s.updateBundleReferences(ctx, &api.ID, defaultTargetURLPerBundleForUpdate)
		if err != nil {
			return err
		}
	}

	err = s.createBundleReferences(ctx, &api.ID, defaultTargetURLPerBundleForCreation)
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

func (s *service) ListFetchRequests(ctx context.Context, specIDs []string) ([]*model.FetchRequest, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fetchRequests, err := s.specService.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDs)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	return fetchRequests, nil
}

func (s *service) updateBundleReferences(ctx context.Context, apiID *string, defaultTargetURLPerBundleForUpdate map[string]string) error {
	for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundleForUpdate {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: &defaultTargetURL,
		}
		err := s.bundleReferenceService.UpdateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, apiID, &crrBndlID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *service) createBundleReferences(ctx context.Context, apiID *string, defaultTargetURLPerBundleForCreation map[string]string) error {
	for crrBndlID, defaultTargetURL := range defaultTargetURLPerBundleForCreation {
		bundleRefInput := &model.BundleReferenceInput{
			APIDefaultTargetURL: &defaultTargetURL,
		}
		err := s.bundleReferenceService.CreateByReferenceObjectID(ctx, *bundleRefInput, model.BundleAPIReference, apiID, &crrBndlID)
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
