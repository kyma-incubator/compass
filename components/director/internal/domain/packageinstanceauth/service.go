package packageinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Repository -output=automock -outpkg=automock -case=underscore
type Repository interface {
	Create(ctx context.Context, item *model.PackageInstanceAuth) error
	GetByID(ctx context.Context, tenantID string, id string) (*model.PackageInstanceAuth, error)
	GetForPackage(ctx context.Context, tenant string, id string, packageID string) (*model.PackageInstanceAuth, error)
	ListByPackageID(ctx context.Context, tenantID string, packageID string) ([]*model.PackageInstanceAuth, error)
	Update(ctx context.Context, item *model.PackageInstanceAuth) error
	Delete(ctx context.Context, tenantID string, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo         Repository
	uidService   UIDService
	timestampGen timestamp.Generator
}

func NewService(repo Repository, uidService UIDService) *service {
	return &service{
		repo:         repo,
		uidService:   uidService,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) Get(ctx context.Context, id string) (*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	instanceAuth, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Package Instance Auth")
	}

	return instanceAuth, nil
}

func (s *service) GetForPackage(ctx context.Context, id string, packageID string) (*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkg, err := s.repo.GetForPackage(ctx, tnt, id, packageID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Package Instance Auth with ID: [%s]", id)
	}

	return pkg, nil
}

func (s *service) RequestDeletion(ctx context.Context, instanceAuth *model.PackageInstanceAuth, defaultPackageInstanceAuth *model.Auth) (bool, error) {
	if instanceAuth == nil {
		return false, errors.New("instance auth is required to request its deletion")
	}

	if defaultPackageInstanceAuth == nil {
		err := instanceAuth.SetDefaultStatus(model.PackageInstanceAuthStatusConditionUnused, s.timestampGen())
		if err != nil {
			return false, errors.Wrapf(err, "while setting status of Instance Auth with ID '%s' to '%s'", instanceAuth.ID, model.PackageInstanceAuthStatusConditionUnused)
		}

		err = s.repo.Update(ctx, instanceAuth)
		if err != nil {
			return false, errors.Wrapf(err, "while updating Package Instance Auth with ID %s", instanceAuth.ID)
		}

		return false, nil
	}

	err := s.Delete(ctx, instanceAuth.ID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.repo.Delete(ctx, tnt, id)

	return errors.Wrapf(err, "while deleting Package Instance Auth with ID %s", id)
}

func (s *service) List(ctx context.Context, packageID string) ([]*model.PackageInstanceAuth, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pkgInstanceAuths, err := s.repo.ListByPackageID(ctx, tnt, packageID)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Package Instance Auths")
	}

	return pkgInstanceAuths, nil
}
