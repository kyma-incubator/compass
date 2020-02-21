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
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.PackagePage, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	pkgRepo PackageRepository

	uidService   UIDService
	timestampGen timestamp.Generator
}

func NewService(pkgRepo PackageRepository, uidService UIDService) *service {
	return &service{
		pkgRepo:      pkgRepo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator(),
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

	//TODO implement creating related resources
	return id, nil
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
