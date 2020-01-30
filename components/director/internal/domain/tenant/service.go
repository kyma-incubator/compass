package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"
)

//go:generate mockery -name=TenantMappingRepository -output=automock -outpkg=automock -case=underscore
type TenantMappingRepository interface {
	Create(ctx context.Context, item model.BusinessTenantMapping) error
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error)
	ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error)
	Update(ctx context.Context, model *model.BusinessTenantMapping) error
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

func (s *service) GetExternalTenant(ctx context.Context, id string) (string, error) {
	mapping, err := s.tenantMappingRepo.Get(ctx, id)
	if err != nil {
		return "", errors.Wrap(err, "while getting the external tenant")
	}

	return mapping.ExternalTenant, nil
}

func (s *service) GetInternalTenant(ctx context.Context, externalTenant string) (string, error) {
	mapping, err := s.tenantMappingRepo.GetByExternalTenant(ctx, externalTenant)
	if err != nil {
		return "", errors.Wrap(err, "while getting the internal tenant")
	}

	return mapping.ID, nil
}

func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error) {
	if pageSize < 1 || pageSize > 100 {
		return nil, errors.New("page size must be between 1 and 100")
	}

	return s.tenantMappingRepo.List(ctx, pageSize, cursor)
}

func (s *service) Sync(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error {

	tenants := s.multipleToTenantMapping(tenantInputs)

	var tenantsFromDb []*model.BusinessTenantMapping
	tenantPage, err := s.tenantMappingRepo.List(ctx, 100, "")
	if err != nil {
		return errors.Wrap(err, "while listing tenants")
	}
	tenantsFromDb = append(tenantsFromDb, tenantPage.Data...)
	for {
		if !tenantPage.PageInfo.HasNextPage {
			break
		}
		cursor := tenantPage.PageInfo.EndCursor
		tenantPage, err = s.tenantMappingRepo.List(ctx, 100, cursor)
		if err != nil {
			return errors.Wrap(err, "while listing tenants")
		}
		tenantsFromDb = append(tenantsFromDb, tenantPage.Data...)
	}

	for _, tenantFromDb := range tenantsFromDb {
		if tenantFromDb.IsIn(tenants) {
			continue
		}
		err := s.markAsInactive(ctx, *tenantFromDb)
		if err != nil {
			return errors.Wrap(err, "while marking the tenant as inactive")
		}
	}

	err = s.createIfNotExists(ctx, tenants)
	if err != nil {
		return errors.Wrap(err, "while creating tenants")
	}

	return nil
}

func (s *service) multipleToTenantMapping(tenantInputs []model.BusinessTenantMappingInput) []model.BusinessTenantMapping {
	var tenants []model.BusinessTenantMapping

	for _, tenant := range tenantInputs {
		id := s.uidService.Generate()
		internalTenant := s.uidService.Generate()
		tenants = append(tenants, *tenant.ToBusinessTenantMapping(id, internalTenant))
	}
	return tenants
}

func (s *service) Create(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error {
	tenants := s.multipleToTenantMapping(tenantInputs)
	err := s.createIfNotExists(ctx, tenants)
	if err != nil {
		return errors.Wrap(err, "while creating many")
	}
	return nil
}

func (s *service) createIfNotExists(ctx context.Context, tenants []model.BusinessTenantMapping) error {
	for _, tenant := range tenants {
		exists, err := s.tenantMappingRepo.ExistsByExternalTenant(ctx, tenant.ExternalTenant)
		if err != nil {
			return errors.Wrap(err, "while checking the existence of tenant")
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

func (s *service) DeleteMany(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error {
	for _, tenantInput := range tenantInputs {
		tenant, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenantInput.ExternalTenant)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				continue
			}
			return errors.Wrap(err, "while getting the tenant mapping")
		}

		err = s.markAsInactive(ctx, *tenant)
		if err != nil {
			return errors.Wrap(err, "while marking the tenant as inactive")
		}
	}

	return nil
}

func (s *service) markAsInactive(ctx context.Context, tenant model.BusinessTenantMapping) error {
	tenant.Status = model.Inactive

	err := s.tenantMappingRepo.Update(ctx, &tenant)

	if err != nil {
		return errors.Wrap(err, "while updating the tenant")
	}

	return nil
}
