package tenantbusinesstype

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// TenantBusinessTypeRepository represents the Tenant Business Type repository layer
//
//go:generate mockery --name=TenantBusinessTypeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantBusinessTypeRepository interface {
	Create(ctx context.Context, item *model.TenantBusinessType) error
	GetByID(ctx context.Context, id string) (*model.TenantBusinessType, error)
	ListAll(ctx context.Context) ([]*model.TenantBusinessType, error)
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo   TenantBusinessTypeRepository
	uidSvc UIDService
}

// NewService creates a Tenant Business Type service
func NewService(repo TenantBusinessTypeRepository, uidSvc UIDService) *service {
	return &service{
		repo:   repo,
		uidSvc: uidSvc,
	}
}

// Create creates a Tenant Business Type using `in`
func (s *service) Create(ctx context.Context, in *model.TenantBusinessTypeInput) (string, error) {
	tenantBusinessTypeID := s.uidSvc.Generate()

	log.C(ctx).Infof("Creating tenant business type with code: %q and name: %q", in.Code, in.Name)
	if err := s.repo.Create(ctx, in.ToModel(tenantBusinessTypeID)); err != nil {
		return "", errors.Wrapf(err, "while creating tenant business type with code: %q and name: %q", in.Code, in.Name)
	}

	return tenantBusinessTypeID, nil
}

// GetByID get tenant business type with given id
func (s *service) GetByID(ctx context.Context, id string) (*model.TenantBusinessType, error) {
	log.C(ctx).Infof("Getting tenant business type with ID: %q", id)

	tbt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting tenant business type with ID: %q", id)
	}

	return tbt, nil
}

// ListAll lists all tenant business types
func (s *service) ListAll(ctx context.Context) ([]*model.TenantBusinessType, error) {
	log.C(ctx).Infof("Listing all tenant business types")

	tbts, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing all tenant business types")
	}

	return tbts, nil
}
