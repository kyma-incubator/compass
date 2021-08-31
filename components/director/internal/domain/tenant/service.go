package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const subdomainLabelKey = "subdomain"

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

//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore
type LabelRepository interface {
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type labeledService struct {
	service
	labelRepo      LabelRepository
	labelUpsertSvc LabelUpsertService
}

type service struct {
	uidService        UIDService
	tenantMappingRepo TenantMappingRepository
}

func NewService(tenantMapping TenantMappingRepository, uidService UIDService) *service {
	return &service{
		uidService:        uidService,
		tenantMappingRepo: tenantMapping,
	}
}

func NewServiceWithLabels(tenantMapping TenantMappingRepository, uidService UIDService, labelRepo LabelRepository, labelUpsertSvc LabelUpsertService) *labeledService {
	return &labeledService{
		service: service{
			uidService:        uidService,
			tenantMappingRepo: tenantMapping,
		},
		labelRepo:      labelRepo,
		labelUpsertSvc: labelUpsertSvc,
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

func (s *service) GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.GetByExternalTenant(ctx, id)
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

func (s *labeledService) CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error {
	tenants := s.MultipleToTenantMapping(tenantInputs)
	subdomains := tenantSubdomains(tenantInputs)
	for _, tenant := range tenants {
		subdomain := ""
		if s, ok := subdomains[tenant.ExternalTenant]; ok {
			subdomain = s
		}
		if err := s.createIfNotExists(ctx, tenant, subdomain); err != nil {
			return errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
		}
	}

	return nil
}

func (s *labeledService) createIfNotExists(ctx context.Context, tenant model.BusinessTenantMapping, subdomain string) error {
	exists, err := s.tenantMappingRepo.ExistsByExternalTenant(ctx, tenant.ExternalTenant)
	if err != nil {
		return errors.Wrapf(err, "while checking the existence of tenant with external ID %s", tenant.ExternalTenant)
	}
	if exists {
		return nil
	}

	if err = s.tenantMappingRepo.Create(ctx, tenant); err != nil {
		return errors.Wrapf(err, "while creating tenant with ID %s and external ID %s", tenant.ID, tenant.ExternalTenant)
	}

	if len(subdomain) > 0 {
		if err := s.addSubdomainLabel(ctx, tenant.ID, subdomain); err != nil {
			return errors.Wrapf(err, "while setting subdomain label for tenant with external ID %s", tenant.ExternalTenant)
		}
	}

	return nil
}

func tenantSubdomains(tenants []model.BusinessTenantMappingInput) map[string]string {
	subdomains := make(map[string]string)
	for _, t := range tenants {
		if len(t.Subdomain) > 0 {
			subdomains[t.ExternalTenant] = t.Subdomain
		}
	}

	return subdomains
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

func (s *labeledService) ListLabels(ctx context.Context, tenantID string) (map[string]*model.Label, error) {
	log.C(ctx).Infof("getting labels for tenant with ID %s", tenantID)
	if err := s.ensureTenantExists(ctx, tenantID); err != nil {
		return nil, err
	}

	labels, err := s.labelRepo.ListForObject(ctx, tenantID, model.TenantLabelableObject, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "whilie listing labels for tenant with ID %s", tenantID)
	}

	return labels, nil
}

func (s *labeledService) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	tenantID := labelInput.ObjectID
	labelInput.ObjectType = model.TenantLabelableObject

	if err := s.ensureTenantExists(ctx, tenantID); err != nil {
		return errors.Wrapf(err, "while ensuring tenant with %s exists", tenantID)
	}
	if err := s.labelUpsertSvc.UpsertLabel(ctx, tenantID, labelInput); err != nil {
		return errors.Wrapf(err, "while creating label for tenant with ID %s", tenantID)
	}

	return nil
}

func (s *labeledService) addSubdomainLabel(ctx context.Context, tenantID, subdomain string) error {
	label := &model.LabelInput{
		Key:        subdomainLabelKey,
		Value:      subdomain,
		ObjectID:   tenantID,
		ObjectType: model.TenantLabelableObject,
	}
	return s.labelUpsertSvc.UpsertLabel(ctx, tenantID, label)
}

func (s *service) ensureTenantExists(ctx context.Context, id string) error {
	exists, err := s.tenantMappingRepo.Exists(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while checking if tenant with ID %s exists", id)
	}

	if !exists {
		return apperrors.NewNotFoundError(resource.Tenant, id)
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
