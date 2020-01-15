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
	GetByExternalTenant(ctx context.Context, externalTenant, provider string) (*model.TenantMapping, error)
	GetByInternalTenant(ctx context.Context, internalTenant string) (*model.TenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.TenantMappingPage, error)
	ExistsByExternalTenant(ctx context.Context, externalTenant, provider string) (bool, error)
	Update(ctx context.Context, model model.TenantMapping) error
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

	tenants := s.multipleToTenantMapping(tenantInputs)

	tenantsFromDb, err := s.tenantMappingRepo.List(ctx, 100, "")
	if err != nil {
		return errors.Wrap(err, "while listing tenants from the db")
	}

	for _, tenant := range tenants {
		if tenant.IsIn(*tenantsFromDb) {
			continue
		}

		toDelete, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenant.ExternalTenant, tenant.Provider)
		if err != nil {
			return errors.Wrap(err, "while getting the tenant to delete")
		}

		err = s.markAsInactive(ctx, *toDelete)
		if err != nil {
			return errors.Wrap(err, "while syncing tenants")
		}
	}

	err = s.upsert(ctx, tenants)
	return nil
}

func (s *service) multipleToTenantMapping(tenantInputs []model.TenantMappingInput) []model.TenantMapping {
	var tenants []model.TenantMapping

	for _, tenant := range tenantInputs {
		id := s.uidService.Generate()
		internalTenant := s.uidService.Generate()
		tenants = append(tenants, *tenant.ToTenantMapping(id, internalTenant))
	}
	return tenants
}

func (s *service) AbsolutelyNotUpsert(ctx context.Context, tenantInputs []model.TenantMappingInput) error {
	tenants := s.multipleToTenantMapping(tenantInputs)
	err := s.upsert(ctx, tenants)
	if err != nil {
		return errors.Wrap(err, "while UPSERTING? many")
	}
	return nil
}

func (s *service) upsert(ctx context.Context, tenants []model.TenantMapping) error {
	for _, tenant := range tenants {
		exists, err := s.tenantMappingRepo.ExistsByExternalTenant(ctx, tenant.ExternalTenant, tenant.Provider)
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

//func (s *service) Delete(ctx context.Context)

func (s *service) markAsInactive(ctx context.Context, tenant model.TenantMapping) error {
	tenant.Status = model.Inactive

	err := s.tenantMappingRepo.Update(ctx, tenant)

	if err != nil {
		return errors.Wrap(err, "while marking the repo as inactive")
	}

	return nil
}
