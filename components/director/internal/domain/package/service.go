package mp_package

import (
	"context"
	"fmt"

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

func (s *service) Create(ctx context.Context, applicationID string, in model.PackageInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	if len(in.ID) == 0 {
		in.ID = s.uidService.Generate()
	}
	pkg := in.Package(applicationID, tnt)

	err = s.pkgRepo.Create(ctx, pkg)
	if err != nil {
		return "", err
	}

	err = s.createBundles(ctx, applicationID, in.ID, in.Bundles)
	if err != nil {
		return "", errors.Wrap(err, "while creating related Application resources")
	}

	return in.ID, nil
}

func (s *service) CreateMultiple(ctx context.Context, applicationID string, in []*model.PackageInput) error {
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

func (s *service) Update(ctx context.Context, id string, in model.PackageInput) error {
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
	err = s.updateBundles(ctx, id, in.Bundles)
	if err != nil {
		return errors.Wrap(err, "while creating related Application resources")
	}

	return nil
}

func (s *service) CreateOrUpdate(ctx context.Context, appID, id string, in model.PackageInput) error {
	exists, err := s.Exist(ctx, id)
	if err != nil {
		return err
	}

	if !exists {
		_, err := s.Create(ctx, appID, in)
		return err
	}
	return s.Update(ctx, id, in)
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

func (s *service) createBundles(ctx context.Context, appID, pkgID string, bundles []*model.BundleInput) error {
	for _, item := range bundles {
		if len(item.ID) == 0 {
			item.ID = s.uidService.Generate()
		}
		id, err := s.bundleSvc.Create(ctx, appID, *item)
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

func (s *service) updateBundles(ctx context.Context, pkgID string, bundles []*model.BundleInput) error {
	for _, item := range bundles {
		if len(item.ID) == 0 {
			return fmt.Errorf("error while updating bundle for package %s: bundle id must be specified", pkgID)
		}
		err := s.bundleSvc.Update(ctx, item.ID, *item)
		if err != nil {
			return errors.Wrap(err, "while creating Bundle for Package")
		}
	}

	return nil
}
