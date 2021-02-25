package mp_package

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery -name=PackageRepository -output=automock -outpkg=automock -case=underscore
type PackageRepository interface {
	Create(ctx context.Context, item *model.Package) error
	Update(ctx context.Context, item *model.Package) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Package, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	pkgRepo PackageRepository

	uidService UIDService
}

func NewService(pkgRepo PackageRepository, uidService UIDService) *service {
	return &service{
		pkgRepo:    pkgRepo,
		uidService: uidService,
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.PackageInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	pkg := in.ToPackage(id, tnt, applicationID)

	err = s.pkgRepo.Create(ctx, pkg)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Package with id %s and title %s for Application with id %s", id, pkg.Title, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Package with id %s and title %s for Application with id %s", id, pkg.Title, applicationID)

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.PackageInput) error {
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
		return false, errors.Wrapf(err, "while getting Package with ID: %q", id)
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
		return nil, errors.Wrapf(err, "while getting Package with ID: %q", id)
	}

	return pkg, nil
}
