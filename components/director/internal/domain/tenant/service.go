package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	SubdomainLabelKey = "subdomain"
	RegionLabelKey    = "region"
)

//go:generate mockery --name=TenantMappingRepository --output=automock --outpkg=automock --case=underscore
type TenantMappingRepository interface {
	Create(ctx context.Context, item model.BusinessTenantMapping) error
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error)
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
	subdomains, regions := tenantLocality(tenantInputs)
	for tenantIdx, tenant := range tenants {
		subdomain := ""
		region := ""
		if s, ok := subdomains[tenant.ExternalTenant]; ok {
			subdomain = s
		}
		if r, ok := regions[tenant.ExternalTenant]; ok {
			region = r
		}
		tenantID, err := s.createIfNotExists(ctx, tenant, subdomain, region)
		if err != nil {
			return errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
		}
		// the tenant already exists in our DB with a different ID, and we should update all child resources to use the correct internal ID
		if tenantID != tenant.ID {
			for i := tenantIdx; i < len(tenants); i++ {
				if tenants[i].Parent == tenant.ID {
					tenants[i].Parent = tenantID
				}
			}
		}
	}

	return nil
}

func (s *labeledService) createIfNotExists(ctx context.Context, tenant model.BusinessTenantMapping, subdomain, region string) (string, error) {
	tenantID := tenant.ID
	tenantFromDB, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenant.ExternalTenant)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return "", errors.Wrapf(err, "while checking the existence of tenant with external ID %s", tenant.ExternalTenant)
	}
	if tenantFromDB != nil {
		return tenantFromDB.ID, s.upsertLabels(ctx, tenantFromDB.ID, subdomain, region)
	}

	if err = s.tenantMappingRepo.Create(ctx, tenant); err != nil && !apperrors.IsNotUniqueError(err) {
		return "", errors.Wrapf(err, "while creating tenant with ID %s and external ID %s", tenant.ID, tenant.ExternalTenant)
	} else if apperrors.IsNotUniqueError(err) {
		tenantFromDB, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenant.ExternalTenant)
		if err != nil {
			return "", errors.Wrapf(err, "while getting internal ID of tenant %s", tenant.ExternalTenant)
		}
		tenantID = tenantFromDB.ID
	}

	return tenantID, s.upsertLabels(ctx, tenant.ID, subdomain, region)
}

func (s *labeledService) upsertLabels(ctx context.Context, tenantID, subdomain, region string) error {
	if len(subdomain) > 0 {
		if err := s.upsertSubdomainLabel(ctx, tenantID, subdomain); err != nil {
			return errors.Wrapf(err, "while setting subdomain label for tenant with ID %s", tenantID)
		}
	}
	if len(region) > 0 {
		if err := s.upsertRegionLabel(ctx, tenantID, region); err != nil {
			return errors.Wrapf(err, "while setting subdomain label for tenant with ID %s", tenantID)
		}
	}
	return nil
}

func tenantLocality(tenants []model.BusinessTenantMappingInput) (map[string]string, map[string]string) {
	subdomains := make(map[string]string)
	regions := make(map[string]string)
	for _, t := range tenants {
		if len(t.Subdomain) > 0 {
			subdomains[t.ExternalTenant] = t.Subdomain
		}
		if len(t.Region) > 0 {
			regions[t.ExternalTenant] = t.Region
		}
	}

	return subdomains, regions
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

func (s *labeledService) upsertSubdomainLabel(ctx context.Context, tenantID, subdomain string) error {
	label := &model.LabelInput{
		Key:        SubdomainLabelKey,
		Value:      subdomain,
		ObjectID:   tenantID,
		ObjectType: model.TenantLabelableObject,
	}
	return s.labelUpsertSvc.UpsertLabel(ctx, tenantID, label)
}

func (s *labeledService) upsertRegionLabel(ctx context.Context, tenantID, region string) error {
	label := &model.LabelInput{
		Key:        RegionLabelKey,
		Value:      region,
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
