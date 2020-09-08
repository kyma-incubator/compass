package mp_package

import (
	"context"

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
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.PackagePage, error)
	AssociateBundle(ctx context.Context, id, bundleID string) error
}

//go:generate mockery -name=BundleRepository -output=automock -outpkg=automock -case=underscore
type BundleRepository interface {
	Create(ctx context.Context, item *model.Bundle) error
	Update(ctx context.Context, item *model.Bundle) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Bundle, error)
	GetForApplication(ctx context.Context, tenant string, id string, applicationID string) (*model.Bundle, error)
	GetByInstanceAuthID(ctx context.Context, tenant string, instanceAuthID string) (*model.Bundle, error)
	ListByApplicationID(ctx context.Context, tenantID, applicationID string, pageSize int, cursor string) (*model.BundlePage, error)
	GetForPackage(ctx context.Context, tenantID, id string, packageID string) (*model.Bundle, error)
	ListByPackageID(ctx context.Context, tenantID, packageID string, pageSize int, cursor string) (*model.BundlePage, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	pkgRepo    PackageRepository
	bundleRepo BundleRepository
	bundleSvc  BundleService

	uidService   UIDService
	timestampGen timestamp.Generator
}

func (s *service) AssociateBundle(ctx context.Context, id, bundleID string) error {
	return s.pkgRepo.AssociateBundle(ctx, id, bundleID)
}

func NewService(pkgRepo PackageRepository, bundleRepo BundleRepository, uidService UIDService, bundleSvc BundleService) *service {
	return &service{
		pkgRepo:      pkgRepo,
		bundleRepo:   bundleRepo,
		uidService:   uidService,
		bundleSvc:    bundleSvc,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.PackageCreateInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	pkg := in.Package(id, applicationID, tnt)

	err = s.pkgRepo.Create(ctx, pkg)
	if err != nil {
		return "", err
	}

	err = s.createBundles(ctx, applicationID, id, in.Bundles)
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
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
	}

	return s.pkgRepo.ListByApplicationID(ctx, tnt, applicationID, pageSize, cursor)
}

func (s *service) createBundles(ctx context.Context, appID, pkgID string, bundles []*model.BundleCreateInput) error {
	for _, item := range bundles {
		id, err := s.bundleSvc.Create(ctx,appID, *item)
		if err != nil {
			return errors.Wrap(err, "while creating Bundle for Package")
		}

		err = s.AssociateBundle(ctx, pkgID, id)
		if err != nil {
			return errors.Wrap(err, "while associating Bundle with Package")
		}
	}

	return nil
}
