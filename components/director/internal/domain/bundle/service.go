package mp_bundle

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery -name=BundleRepository -output=automock -outpkg=automock -case=underscore
type BundleRepository interface {
	Create(ctx context.Context, item *model.Bundle) error
	Update(ctx context.Context, item *model.Bundle) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error)
	ExistsByCondition(ctx context.Context, tenant string, conds repo.Conditions) (bool, error)
	GetByConditions(ctx context.Context, tenant string, conds repo.Conditions) (*model.Bundle, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error)
	GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Bundle, error)
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.BundlePage, error)
	GetForPackage(ctx context.Context, tenantID, id string, packageID string) (*model.Bundle, error)
	ListByPackageID(ctx context.Context, tenantID, packageID string, pageSize int, cursor string) (*model.BundlePage, error)
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	CreateInBundle(ctx context.Context, bundleID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	Delete(ctx context.Context, id string) error
	ExistsByCondition(ctx context.Context, conds repo.Conditions) (bool, error)
	Exists(ctx context.Context, id string) (bool, error)
	GetByConditions(ctx context.Context, conds repo.Conditions) (*model.APIDefinition, error)
	GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error)
	ListForBundle(ctx context.Context, bundleID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	RefetchAPISpecs(ctx context.Context, id string) ([]*model.APISpec, error)
}

//go:generate mockery -name=EventDefService -output=automock -outpkg=automock -case=underscore
type EventDefService interface {
	CreateInBundle(ctx context.Context, bundleID string, in model.EventDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput) error
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	Delete(ctx context.Context, id string) error
	RefetchAPISpec(ctx context.Context, id string) (*model.EventSpec, error)
	GetFetchRequest(ctx context.Context, eventAPIDefID string) (*model.FetchRequest, error)
	GetByConditions(ctx context.Context, conds repo.Conditions) (*model.EventDefinition, error)
	ExistsByCondition(ctx context.Context, conds repo.Conditions) (bool, error)
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	Create(ctx context.Context, item *model.Document) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery -name=FetchRequestService -output=automock -outpkg=automock -case=underscore
type FetchRequestService interface {
	HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string
}

type service struct {
	bundleRepo       BundleRepository
	apiSvc           APIService
	eventSvc         EventDefService
	documentRepo     DocumentRepository
	fetchRequestRepo FetchRequestRepository

	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(bundleRepo BundleRepository, apiSvc APIService, eventSvc EventDefService, documentRepo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{
		bundleRepo:          bundleRepo,
		apiSvc:              apiSvc,
		eventSvc:            eventSvc,
		documentRepo:        documentRepo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.BundleInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	bundle := in.ToBundle(applicationID, tnt)

	err = s.bundleRepo.Create(ctx, bundle)
	if err != nil {
		return "", err
	}

	err = s.createRelatedResources(ctx, in, tnt, in.ID)
	if err != nil {
		return "", errors.Wrap(err, "while creating related Bundle resources")
	}

	return in.ID, nil
}

func (s *service) CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleInput) error {
	if in == nil {
		return nil
	}

	for _, bundle := range in {
		if bundle == nil {
			continue
		}

		_, err := s.Create(ctx, applicationID, *bundle)
		if err != nil {
			return errors.Wrap(err, "while creating Bundle for Application")
		}
	}

	return nil
}

func (s *service) Update(ctx context.Context, id string, in model.BundleInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bundle, err := s.bundleRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	bundle.SetFromUpdateInput(in)

	err = s.bundleRepo.Update(ctx, bundle)
	if err != nil {
		return errors.Wrapf(err, "while updating Bundle with ID: [%s]", id)
	}
	err = s.updateRelatedResources(ctx, in, tnt, id)
	if err != nil {
		return errors.Wrap(err, "while updating related Bundle resources")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.bundleRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Bundle with ID: [%s]", id)
	}

	return nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.bundleRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return exist, nil
}

func (s *service) CreateOrUpdate(ctx context.Context, appID, openDiscoveryID string, in model.BundleInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "while loading tenant from context")
	}
	exists, err := s.bundleRepo.ExistsByCondition(ctx, tnt, repo.Conditions{
		repo.NewEqualCondition("od_id", openDiscoveryID),
		repo.NewEqualCondition("app_id", appID),
	})
	if err != nil {
		return "", err
	}

	if !exists {
		if len(in.ID) == 0 {
			in.ID = s.uidService.Generate()
		}
		bundle := in.ToBundle(appID, tnt)

		err = s.bundleRepo.Create(ctx, bundle)
		if err != nil {
			return "", err
		}
	} else {
		bundle, err := s.bundleRepo.GetByConditions(ctx, tnt, repo.Conditions{
			repo.NewEqualCondition("od_id", openDiscoveryID),
			repo.NewEqualCondition("app_id", appID),
		})
		if err != nil {
			return "", err
		}
		in.ID = bundle.ID
		bundle.SetFromUpdateInput(in)

		err = s.bundleRepo.Update(ctx, bundle)
		if err != nil {
			return "", errors.Wrapf(err, "while updating Bundle with ID: [%s]", bundle.ID)
		}
	}
	if err := s.createOrUpdateRelatedResources(ctx, in, tnt, in.ID); err != nil {
		return "", err
	}
	return in.ID, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	bundle, err := s.bundleRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bundle, nil
}

func (s *service) GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bundle, err := s.bundleRepo.GetByInstanceAuthID(ctx, tnt, instanceAuthID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle by Instance Auth ID: [%s]", instanceAuthID)
	}

	return bundle, nil
}

func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bundle, err := s.bundleRepo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Bundle with ID: [%s]", id)
	}

	return bundle, nil
}

func (s *service) ListByApplicationID(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
	}

	return s.bundleRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) createRelatedResources(ctx context.Context, in model.BundleInput, tenant string, bundleID string) error {
	err := s.createAPIs(ctx, bundleID, tenant, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.createEvents(ctx, bundleID, tenant, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	err = s.createDocuments(ctx, bundleID, tenant, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) createOrUpdateRelatedResources(ctx context.Context, in model.BundleInput, tenant string, bundleID string) error {
	err := s.createOrUpdateAPIs(ctx, bundleID, tenant, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.createOrUpdateEvents(ctx, bundleID, tenant, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	err = s.createOrUpdateDocuments(ctx, bundleID, tenant, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) updateRelatedResources(ctx context.Context, in model.BundleInput, tenant string, bundleID string) error {
	err := s.updateAPIs(ctx, bundleID, tenant, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.updateEvents(ctx, bundleID, tenant, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	/*err = s.updateDocuments(ctx, bundleID, tenant, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}*/ // TODO: Documents does not support update right now

	return nil
}

func (s *service) createAPIs(ctx context.Context, bundleID, tenant string, apis []*model.APIDefinitionInput) error {
	var err error
	for _, item := range apis {
		_, err = s.apiSvc.CreateInBundle(ctx, bundleID, *item)
		if err != nil {
			return errors.Wrapf(err, "while creating API for bundle with id %s", bundleID)
		}
	}

	return nil
}

func (s *service) updateAPIs(ctx context.Context, bundleID, tenant string, apis []*model.APIDefinitionInput) error {
	for _, item := range apis {
		if len(item.ID) == 0 {
			return fmt.Errorf("id is mandatory when updating APIs")
		}
		err := s.apiSvc.Update(ctx, item.ID, *item)
		if err != nil {
			return errors.Wrapf(err, "while updating API for bundle with id %s", bundleID)
		}
	}
	return nil
}

func (s *service) createOrUpdateAPIs(ctx context.Context, bundleID, tenant string, apis []*model.APIDefinitionInput) error {
	toUpdate := make([]*model.APIDefinitionInput, 0, 0)
	toCreate := make([]*model.APIDefinitionInput, 0, 0)
	for i := range apis {
		exists, err := s.apiSvc.ExistsByCondition(ctx, repo.Conditions{
			repo.NewEqualCondition("od_id", apis[i].OpenDiscoveryID),
			repo.NewEqualCondition("bundle_id", bundleID),
		})
		if err != nil {
			return err
		}
		if exists {
			api, err := s.apiSvc.GetByConditions(ctx, repo.Conditions{
				repo.NewEqualCondition("od_id", apis[i].OpenDiscoveryID),
				repo.NewEqualCondition("bundle_id", bundleID),
			})
			if err != nil {
				return err
			}
			apis[i].ID = api.ID
			toUpdate = append(toUpdate, apis[i])
		} else {
			toCreate = append(toCreate, apis[i])
		}
	}
	if err := s.createAPIs(ctx, bundleID, tenant, toCreate); err != nil {
		return err
	}
	if err := s.updateAPIs(ctx, bundleID, tenant, toUpdate); err != nil {
		return err
	}
	return nil
}

func (s *service) createEvents(ctx context.Context, bundleID, tenant string, events []*model.EventDefinitionInput) error {
	var err error
	for _, item := range events {
		_, err = s.eventSvc.CreateInBundle(ctx, bundleID, *item)
		if err != nil {
			return errors.Wrapf(err, "while creating Event for bundle with id %s", bundleID)
		}
	}
	return nil
}

func (s *service) updateEvents(ctx context.Context, bundleID, tenant string, events []*model.EventDefinitionInput) error {
	for _, item := range events {
		if len(item.ID) == 0 {
			return fmt.Errorf("id is mandatory when updating Events")
		}
		err := s.eventSvc.Update(ctx, item.ID, *item)
		if err != nil {
			return errors.Wrapf(err, "while updating EventDefinitions for bundle with id %s", bundleID)
		}
	}
	return nil
}

func (s *service) createOrUpdateEvents(ctx context.Context, bundleID, tenant string, events []*model.EventDefinitionInput) error {
	toUpdate := make([]*model.EventDefinitionInput, 0, 0)
	toCreate := make([]*model.EventDefinitionInput, 0, 0)
	for i := range events {
		exists, err := s.eventSvc.ExistsByCondition(ctx, repo.Conditions{
			repo.NewEqualCondition("od_id", events[i].OpenDiscoveryID),
			repo.NewEqualCondition("bundle_id", bundleID),
		})
		if err != nil {
			return err
		}
		if exists {
			event, err := s.eventSvc.GetByConditions(ctx, repo.Conditions{
				repo.NewEqualCondition("od_id", events[i].OpenDiscoveryID),
				repo.NewEqualCondition("bundle_id", bundleID),
			})
			if err != nil {
				return err
			}
			events[i].ID = event.ID
			toUpdate = append(toUpdate, events[i])
		} else {
			toCreate = append(toCreate, events[i])
		}
	}
	if err := s.createEvents(ctx, bundleID, tenant, toCreate); err != nil {
		return err
	}
	if err := s.updateEvents(ctx, bundleID, tenant, toUpdate); err != nil {
		return err
	}
	return nil
}

func (s *service) createDocuments(ctx context.Context, bundleID, tenant string, documents []*model.DocumentInput) error {
	var err error
	for _, item := range documents {
		if len(item.ID) == 0 {
			item.ID = s.uidService.Generate()
		}
		err = s.documentRepo.Create(ctx, item.ToDocumentWithinBundle(tenant, bundleID))
		if err != nil {
			return errors.Wrapf(err, "while creating Document for application")
		}

		if item.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.FetchRequest, model.DocumentFetchRequestReference, item.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) createOrUpdateDocuments(ctx context.Context, bundleID, tenant string, documens []*model.DocumentInput) error {
	toUpdate := make([]*model.DocumentInput, 0, 0)
	toCreate := make([]*model.DocumentInput, 0, 0)
	for i := range documens {
		exists, err := s.documentRepo.Exists(ctx, tenant, documens[i].ID)
		if err != nil {
			return err
		}
		if exists {
			toUpdate = append(toUpdate, documens[i])
		} else {
			toCreate = append(toCreate, documens[i])
		}
	}
	if err := s.createDocuments(ctx, bundleID, tenant, toCreate); err != nil {
		return err
	}
	/*if err := s.updateDocuments(ctx, bundleID, tenant, toUpdate); err != nil {
		return err
	}*/ // TODO: Documents does not support update
	return nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, objectType model.FetchRequestReferenceObjectType, objectID string) (*model.FetchRequest, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, objectType, objectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", objectType, objectID)
	}

	return fr, nil
}

func (s *service) ListForPackage(ctx context.Context, packageID string, pageSize int, cursor string) (*model.BundlePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.bundleRepo.ListByPackageID(ctx, tnt, packageID, pageSize, cursor)
}

func (s *service) GetForPackage(ctx context.Context, id string, packageID string) (*model.Bundle, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.bundleRepo.GetForPackage(ctx, tnt, id, packageID)
}
