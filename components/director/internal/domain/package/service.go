package ordpackage

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// PackageRepository missing godoc
//
//go:generate mockery --name=PackageRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type PackageRepository interface {
	Create(ctx context.Context, tenant string, item *model.Package) error
	CreateGlobal(ctx context.Context, model *model.Package) error
	Update(ctx context.Context, tenant string, item *model.Package) error
	UpdateGlobal(ctx context.Context, model *model.Package) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Package, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Package, error)
	ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.Package, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	pkgRepo PackageRepository

	uidService UIDService
}

// NewService missing godoc
func NewService(pkgRepo PackageRepository, uidService UIDService) *service {
	return &service{
		pkgRepo:    pkgRepo,
		uidService: uidService,
	}
}

// Create creates a package for a given resource.Type
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.PackageInput, pkgHash uint64) (string, error) {
	id := s.uidService.Generate()
	pkg := in.ToPackage(id, resourceType, resourceID, pkgHash)

	var (
		err error
		tnt string
	)
	if resourceType == resource.ApplicationTemplateVersion {
		err = s.pkgRepo.CreateGlobal(ctx, pkg)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return "", err
		}

		err = s.pkgRepo.Create(ctx, tnt, pkg)
	}
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Package with id %s and title %s for %s with id %s", id, pkg.Title, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created a Package with id %s and title %s for %s with id %s", id, pkg.Title, resourceType, resourceID)

	return id, nil
}

// Update updates a package by ID for a given resource.Type
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.PackageInput, pkgHash uint64) error {
	var (
		pkg *model.Package
		tnt string
		err error
	)

	if resourceType == resource.ApplicationTemplateVersion {
		pkg, err = s.pkgRepo.GetByIDGlobal(ctx, id)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return err
		}

		pkg, err = s.pkgRepo.GetByID(ctx, tnt, id)
	}
	if err != nil {
		return errors.Wrapf(err, "while getting Package with id %s", id)
	}

	pkg.SetFromUpdateInput(in, pkgHash)

	if resourceType == resource.ApplicationTemplateVersion {
		err = s.pkgRepo.UpdateGlobal(ctx, pkg)
	} else {
		err = s.pkgRepo.Update(ctx, tnt, pkg)
	}
	if err != nil {
		return errors.Wrapf(err, "while updating Package with id %s", id)
	}

	return nil
}

// Delete missing godoc
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

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.pkgRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Package with ID: %q", id)
	}

	return exist, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.Package, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	pkg, err := s.pkgRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package with ID: %q", id)
	}

	return pkg, nil
}

// ListByApplicationID missing godoc
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.pkgRepo.ListByResourceID(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionID lists packages by Application Template Version ID without tenant isolation
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Package, error) {
	return s.pkgRepo.ListByResourceID(ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion)
}
