package tenant

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=IntegrationSystemRepository -output=automock -outpkg=automock -case=underscore
type TenantMappingRepository interface {
	Create(ctx context.Context, item model.TenantMapping) error
	Get(ctx context.Context, id string) (*model.TenantMapping, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.TenantMapping, error)
	GetByInternalTenant(ctx context.Context, internalTenant string) (*model.TenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.TenantMappingPage, error)
	ExistsByExternalTenant(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, model model.TenantMapping) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	tenantMappingRepo TenantMappingRepository

	uidService UIDService
}

func NewService(tenantMapping TenantMappingRepository, uidService UIDService) *service {
	return &service{
		tenantMappingRepo: tenantMapping,
		uidService:        uidService,
	}
}

func (s *service) GetExternalTenant(ctx context.Context, internalTenant string) (string, error) {
	mapping, err := s.tenantMappingRepo.GetByInternalTenant(ctx, internalTenant)
	if err != nil {
		return "", errors.Wrap(err, "while getting the external tenant")
	}
	if mapping == nil {
		return "", errors.New(fmt.Sprintf("tenant with the id %s is inactive", internalTenant))
	}

	return mapping.ExternalTenant, nil
}

func (s *service) GetInternalTenant(ctx context.Context, externalTenant string) (string, error) {
	mapping, err := s.tenantMappingRepo.GetByExternalTenant(ctx, externalTenant)
	if err != nil {
		return "", errors.Wrap(err, "while getting the internal tenant")
	}

	return mapping.InternalTenant, nil
}

func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.TenantMappingPage, error) {
	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("page size must be between 1 and 100")
	}

	return s.tenantMappingRepo.List(ctx, pageSize, cursor)
}

//Just a prototype
func (s *service) Sync(ctx context.Context, tenantInputs []model.TenantMappingInput) error {

	var tenants []model.TenantMapping

	for _, tenant := range tenantInputs {
		id := s.uidService.Generate()
		internalTenant := s.uidService.Generate()
		tenants = append(tenants, *tenant.ToTenantMapping(id, internalTenant))
	}

	tenantsFromDb, err := s.tenantMappingRepo.List(ctx, 100, "")
	if err != nil {
		return errors.Wrap(err, "while listing tenants from the db")
	}

	for _, tenant := range tenants {
		if tenant.IsIn(*tenantsFromDb) {
			continue
		}
		err := s.markAsInactive(ctx, tenant)
		if err != nil {
			return errors.Wrap(err, "while syncing tenants")
		}
	}
	err = s.AbsolutelyNotUpsert(ctx, tenants)
	return nil
}

func (s *service) AbsolutelyNotUpsert(ctx context.Context, tenants []model.TenantMapping) error {
	for _, tenant := range tenants {
		exists, err := s.tenantMappingRepo.ExistsByExternalTenant(ctx, tenant.ExternalTenant)
		if err != nil {
			return errors.Wrap(err, "while checking the existance of tenant")
		}
		if exists {
			continue
		}
		err = s.tenantMappingRepo.Create(ctx, tenant)
		if err != nil {
			return errors.Wrap(err, "while creating the tenant")
		}
	}
	return nil
}

func (s *service) markAsInactive(ctx context.Context, tenant model.TenantMapping) error {
	tenant.Status = model.Inactive

	err := s.tenantMappingRepo.Update(ctx, tenant)

	if err != nil {
		return errors.Wrap(err, "while marking the repo as inactive")
	}

	return nil
}
