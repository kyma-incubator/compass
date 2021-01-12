package mp_package

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/pkg/errors"
)

//go:generate mockery -name=PackageRepository -output=automock -outpkg=automock -case=underscore
type PackageRepository interface {
	Create(ctx context.Context, item *model.Package) error
	Update(ctx context.Context, item *model.Package) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Package, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Package, error)
	GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Package, error)
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.PackagePage, error)
}

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	Create(ctx context.Context, item *model.APIDefinition) error
	Update(ctx context.Context, item *model.APIDefinition) error
}

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	Create(ctx context.Context, items *model.EventDefinition) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	Create(ctx context.Context, item *model.Document) error
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
	pkgRepo          PackageRepository
	apiRepo          APIRepository
	eventAPIRepo     EventAPIRepository
	documentRepo     DocumentRepository
	fetchRequestRepo FetchRequestRepository

	uidService          UIDService
	fetchRequestService FetchRequestService
	timestampGen        timestamp.Generator
}

func NewService(pkgRepo PackageRepository, apiRepo APIRepository, eventAPIRepo EventAPIRepository, documentRepo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService, fetchRequestService FetchRequestService) *service {
	return &service{
		pkgRepo:             pkgRepo,
		apiRepo:             apiRepo,
		eventAPIRepo:        eventAPIRepo,
		documentRepo:        documentRepo,
		fetchRequestRepo:    fetchRequestRepo,
		uidService:          uidService,
		fetchRequestService: fetchRequestService,
		timestampGen:        timestamp.DefaultGenerator(),
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.PackageCreateInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	pkg := in.ToPackage(id, applicationID, tnt)

	err = s.pkgRepo.Create(ctx, pkg)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Package with id %s and name %s for Application with id %s", id, pkg.Name, applicationID)
	}
	log.C(ctx).Infof("Successfully created a Package with id %s and name %s for Application with id %s", id, pkg.Name, applicationID)

	log.C(ctx).Infof("Creating related resources in Package with id %s and name %s for Application with id %s", id, pkg.Name, applicationID)
	err = s.createRelatedResources(ctx, in, tnt, id)
	if err != nil {
		return "", errors.Wrapf(err, "while creating related resources for Application with id %s", applicationID)
	}

	return id, nil
}

func (s *service) CreateMultiple(ctx context.Context, applicationID string, in []*model.PackageCreateInput) error {
	if in == nil {
		return nil
	}

	for _, pkg := range in {
		if pkg == nil {
			continue
		}

		_, err := s.Create(ctx, applicationID, *pkg)
		if err != nil {
			return errors.Wrapf(err, "while creating Package for Application with id %s", applicationID)
		}
	}

	return nil
}

func (s *service) Update(ctx context.Context, id string, in model.PackageUpdateInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	pkg, err := s.pkgRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Package with id %s", id)
	}

	pkg.SetFromUpdateInput(in)

	err = s.pkgRepo.Update(ctx, pkg)
	if err != nil {
		return errors.Wrapf(err, "while updating Package with id %s", id)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.pkgRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Package with id %s", id)
	}

	return nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.pkgRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Package with ID: [%s]", id)
	}

	return exist, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.Package, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	pkg, err := s.pkgRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package with ID: [%s]", id)
	}

	return pkg, nil
}

func (s *service) GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Package, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := s.pkgRepo.GetByInstanceAuthID(ctx, tnt, instanceAuthID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package by PackageInstanceAuth with id %s", instanceAuthID)
	}

	return pkg, nil
}

func (s *service) GetForApplication(ctx context.Context, id string, applicationID string) (*model.Package, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := s.pkgRepo.GetForApplication(ctx, tnt, id, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package with ID: [%s]", id)
	}

	return pkg, nil
}

func (s *service) ListByApplicationID(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.PackagePage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.pkgRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) createRelatedResources(ctx context.Context, in model.PackageCreateInput, tenant string, packageID string) error {
	err := s.createAPIs(ctx, packageID, tenant, in.APIDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	err = s.createEvents(ctx, packageID, tenant, in.EventDefinitions)
	if err != nil {
		return errors.Wrapf(err, "while creating Events for application")
	}

	err = s.createDocuments(ctx, packageID, tenant, in.Documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) createAPIs(ctx context.Context, packageID, tenant string, apis []*model.APIDefinitionInput) error {
	var err error
	for _, item := range apis {
		apiDefID := s.uidService.Generate()

		api := item.ToAPIDefinitionWithinPackage(apiDefID, packageID, tenant)

		err = s.apiRepo.Create(ctx, api)
		if err != nil {
			return errors.Wrapf(err, "while creating APIDefinition with id %s within Package with id %s", apiDefID, packageID)
		}
		log.C(ctx).Infof("Successfully created APIDefinition with id %s within Package with id %s", apiDefID, packageID)

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			fr, err := s.createFetchRequest(ctx, tenant, item.Spec.FetchRequest, model.APIFetchRequestReference, apiDefID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}

			api.Spec.Data = s.fetchRequestService.HandleAPISpec(ctx, fr)
			err = s.apiRepo.Update(ctx, api)
			if err != nil {
				return errors.Wrap(err, "while updating api with api spec")
			}
		}

	}

	return nil
}

func (s *service) createEvents(ctx context.Context, packageID, tenant string, events []*model.EventDefinitionInput) error {
	var err error
	for _, item := range events {
		eventID := s.uidService.Generate()
		err = s.eventAPIRepo.Create(ctx, item.ToEventDefinitionWithinPackage(eventID, packageID, tenant))
		if err != nil {
			return errors.Wrapf(err, "while creating EventDefinition with id %s in Package with id %s", eventID, packageID)
		}
		log.C(ctx).Infof("Successfully created EventDefinition with id %s in Package with id %s", eventID, packageID)

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.Spec.FetchRequest, model.EventAPIFetchRequestReference, eventID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}
		}
	}
	return nil
}

func (s *service) createDocuments(ctx context.Context, packageID, tenant string, events []*model.DocumentInput) error {
	var err error
	for _, item := range events {
		documentID := s.uidService.Generate()

		err = s.documentRepo.Create(ctx, item.ToDocumentWithinPackage(documentID, tenant, packageID))
		if err != nil {
			return errors.Wrapf(err, "while creating Document with id %s in Package with id %s", documentID, packageID)
		}
		log.C(ctx).Infof("Successfully created Document with id %s in Package with id %s", documentID, packageID)

		if item.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.FetchRequest, model.DocumentFetchRequestReference, documentID)
			if err != nil {
				return err
			}
		}
	}
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
		return nil, errors.Wrapf(err, "while creating FetchRequest with id %s for %s with id %s", id, objectType, objectID)
	}
	log.C(ctx).Infof("Successfully created FetchRequest with id %s for type %s with id %s", id, objectType, objectID)
	return fr, nil
}
