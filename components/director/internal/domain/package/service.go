package mp_package

import (
	"context"

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
	ListForApplication(ctx context.Context, tenant, applicationID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	Create(ctx context.Context, item *model.APIDefinition) error
	DeleteAllByApplicationID(ctx context.Context, tenant, id string) error
}

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	ListForApplication(ctx context.Context, tenantID string, applicationID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	Create(ctx context.Context, items *model.EventDefinition) error
	DeleteAllByApplicationID(ctx context.Context, tenantID string, appID string) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	Create(ctx context.Context, item *model.Document) error
	DeleteAllByApplicationID(ctx context.Context, tenant string, applicationID string) error
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Create(ctx context.Context, item *model.FetchRequest) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	pkgRepo          PackageRepository
	apiRepo          APIRepository
	eventAPIRepo     EventAPIRepository
	documentRepo     DocumentRepository
	fetchRequestRepo FetchRequestRepository

	uidService   UIDService
	timestampGen timestamp.Generator
}

func NewService(pkgRepo PackageRepository, apiRepo APIRepository, eventAPIRepo EventAPIRepository, documentRepo DocumentRepository, fetchRequestRepo FetchRequestRepository, uidService UIDService) *service {
	return &service{
		pkgRepo:          pkgRepo,
		apiRepo:          apiRepo,
		eventAPIRepo:     eventAPIRepo,
		documentRepo:     documentRepo,
		fetchRequestRepo: fetchRequestRepo,
		uidService:       uidService,
		timestampGen:     timestamp.DefaultGenerator(),
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
		return "", err
	}

	err = s.createRelatedResources(ctx, in, tnt, id)
	if err != nil {
		return "", errors.Wrap(err, "while creating related Application resources")
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
			return errors.Wrap(err, "while creating Package for Application")
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
		return errors.Wrapf(err, "while getting Package with ID: [%s]", id)
	}

	pkg.SetFromUpdateInput(in)

	err = s.pkgRepo.Update(ctx, pkg)
	if err != nil {
		return errors.Wrapf(err, "while updating Package with ID: [%s]", id)
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
		return errors.Wrapf(err, "while deleting Package with ID: [%s]", id)
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
		return nil, errors.Wrapf(err, "while getting Package by Instance Auth ID: [%s]", instanceAuthID)
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

	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("page size must be between 1 and 100")
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
		err = s.apiRepo.Create(ctx, item.ToAPIDefinitionWithinPackage(apiDefID, &packageID, tenant))
		if err != nil {
			return errors.Wrap(err, "while creating API for application")
		}

		if item.Spec != nil && item.Spec.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.Spec.FetchRequest, model.APIFetchRequestReference, apiDefID)
			if err != nil {
				return errors.Wrap(err, "while creating FetchRequest for application")
			}
		}
	}
	return nil
}

func (s *service) createEvents(ctx context.Context, packageID, tenant string, events []*model.EventDefinitionInput) error {
	var err error
	for _, item := range events {
		eventID := s.uidService.Generate()
		err = s.eventAPIRepo.Create(ctx, item.ToEventDefinitionWithinPackage(eventID, &packageID, tenant))
		if err != nil {
			return errors.Wrap(err, "while creating EventDefinitions for application")
		}

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
		err = s.documentRepo.Create(ctx, item.ToDocumentWithinPackage(documentID, tenant, &packageID))
		if err != nil {
			return errors.Wrapf(err, "while creating Document for application")
		}

		if item.FetchRequest != nil {
			_, err = s.createFetchRequest(ctx, tenant, item.FetchRequest, model.DocumentFetchRequestReference, documentID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *service) createFetchRequest(ctx context.Context, tenant string, in *model.FetchRequestInput, objectType model.FetchRequestReferenceObjectType, objectID string) (*string, error) {
	if in == nil {
		return nil, nil
	}

	id := s.uidService.Generate()
	fr := in.ToFetchRequest(s.timestampGen(), id, tenant, objectType, objectID)
	err := s.fetchRequestRepo.Create(ctx, fr)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating FetchRequest for %s with ID %s", objectType, objectID)
	}

	return &id, nil
}
