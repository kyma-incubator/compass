package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=TenantMappingRepository --output=automock --outpkg=automock --case=underscore
type TenantMappingRepository interface {
	Create(ctx context.Context, item model.BusinessTenantMapping) error
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error)
	Update(ctx context.Context, model *model.BusinessTenantMapping) error
	DeleteByExternalTenant(ctx context.Context, externalTenant string) error
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
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

func (s *service) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.List(ctx)
}

func (s *service) MultipleToTenantMapping(tenantInputs []model.BusinessTenantMappingInput) []model.BusinessTenantMapping {
	var tenants []model.BusinessTenantMapping
	tenantIDs := make(map[string]string, len(tenantInputs))
	for _, tenant := range tenantInputs {
		id := s.uidService.Generate()
		tenants = append(tenants, *tenant.ToBusinessTenantMapping(id))
		tenantIDs[tenant.ExternalTenant] = id
	}
	for i := 0; i < len(tenants); i++ { // Convert parent ID from external to internal id reference
		if len(tenants[i].Parent) > 0 {
			if _, ok := tenantIDs[tenants[i].Parent]; ok { // If the parent is inserted in this request (otherwise we assume that it is already in the db)
				tenants[i].Parent = tenantIDs[tenants[i].Parent]

				var moved bool
				tenants, moved = MoveBeforeIfShould(tenants, tenants[i].Parent, i) // Move my parent before me (to be inserted first) if it is not already
				if moved {
					i-- // Process the moved parent as well
				}
			}
		}
	}
	return tenants
}

func (s *service) CreateManyIfNotExists(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) error {
	tenants := s.MultipleToTenantMapping(tenantInputs)
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
		err := s.tenantMappingRepo.DeleteByExternalTenant(ctx, tenantInput.ExternalTenant)
		if err != nil {
			return errors.Wrap(err, "while deleting tenant")
		}
	}

	return nil
}

// MoveBeforeIfShould moves the tenant with id right before index only if it is not already before it
func MoveBeforeIfShould(tenants []model.BusinessTenantMapping, id string, indx int) ([]model.BusinessTenantMapping, bool) {
	var itemIndex int
	for i, tenant := range tenants {
		if tenant.ID == id {
			itemIndex = i
		}
	}

	if itemIndex <= indx { // already before indx
		return tenants, false
	}

	newTenants := make([]model.BusinessTenantMapping, 0, len(tenants))
	for i := range tenants {
		if i == itemIndex {
			continue
		}
		if i == indx {
			newTenants = append(newTenants, tenants[itemIndex], tenants[i])
			continue
		}
		newTenants = append(newTenants, tenants[i])
	}
	return newTenants, true
}
