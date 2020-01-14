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
	List(ctx context.Context, pageSize int, cursor string) (model.TenantMapping, error)
	ExistsByExternalTenantID(ctx context.Context, id string) (bool, error)
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

func (s *service) Sync(ctx context.Context, tenants model.TenantMappingInput) error {
	return nil
}
