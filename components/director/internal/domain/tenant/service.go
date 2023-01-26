package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	// SubdomainLabelKey is the key of the tenant label for subdomain.
	SubdomainLabelKey = "subdomain"
	// RegionLabelKey is the key of the tenant label for region.
	RegionLabelKey = "region"
)

// TenantMappingRepository is responsible for the repo-layer tenant operations.
//go:generate mockery --name=TenantMappingRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingRepository interface {
	UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) error
	Upsert(ctx context.Context, item model.BusinessTenantMapping) error
	Update(ctx context.Context, model *model.BusinessTenantMapping) error
	Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context) ([]*model.BusinessTenantMapping, error)
	ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error)
	ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error)
	DeleteByExternalTenant(ctx context.Context, externalTenant string) error
	GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error)
	ListByExternalTenants(ctx context.Context, externalTenant []string) ([]*model.BusinessTenantMapping, error)
	ListByParentAndType(ctx context.Context, parentID string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error)
	GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error)
}

// LabelUpsertService is responsible for creating, or updating already existing labels, and their label definitions.
//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

// LabelRepository is responsible for the repo-layer label operations.
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// UIDService is responsible for generating GUIDs, which will be used as internal tenant IDs when tenants are created.
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
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

// NewService returns a new object responsible for service-layer tenant operations.
func NewService(tenantMapping TenantMappingRepository, uidService UIDService) *service {
	return &service{
		uidService:        uidService,
		tenantMappingRepo: tenantMapping,
	}
}

// NewServiceWithLabels returns a new entity responsible for service-layer tenant operations, including operations with labels like listing all labels related to the given tenant.
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

// GetExternalTenant returns the external tenant ID of the tenant with the corresponding internal tenant ID.
func (s *service) GetExternalTenant(ctx context.Context, id string) (string, error) {
	mapping, err := s.tenantMappingRepo.Get(ctx, id)
	if err != nil {
		return "", errors.Wrap(err, "while getting the external tenant")
	}

	return mapping.ExternalTenant, nil
}

// GetInternalTenant returns the internal tenant ID of the tenant with the corresponding external tenant ID.
func (s *service) GetInternalTenant(ctx context.Context, externalTenant string) (string, error) {
	mapping, err := s.tenantMappingRepo.GetByExternalTenant(ctx, externalTenant)
	if err != nil {
		return "", errors.Wrap(err, "while getting the internal tenant")
	}

	return mapping.ID, nil
}

// List returns all tenants present in the Compass storage.
func (s *service) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.List(ctx)
}

// ListsByExternalIDs returns all tenants for provided external IDs.
func (s *service) ListsByExternalIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByExternalTenants(ctx, ids)
}

// ListPageBySearchTerm returns all tenants present in the Compass storage.
func (s *service) ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error) {
	return s.tenantMappingRepo.ListPageBySearchTerm(ctx, searchTerm, pageSize, cursor)
}

// GetTenantByExternalID returns the tenant with the provided external ID.
func (s *service) GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.GetByExternalTenant(ctx, id)
}

// GetTenantByID returns the tenant with the provided ID.
func (s *service) GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.Get(ctx, id)
}

// GetLowestOwnerForResource returns the lowest tenant in the hierarchy that is owner of a given resource.
func (s *service) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	return s.tenantMappingRepo.GetLowestOwnerForResource(ctx, resourceType, objectID)
}

// MultipleToTenantMapping assigns a new internal ID to all the provided tenants, and returns the BusinessTenantMappingInputs as BusinessTenantMappings.
func (s *service) MultipleToTenantMapping(tenantInputs []model.BusinessTenantMappingInput) []model.BusinessTenantMapping {
	tenants := make([]model.BusinessTenantMapping, 0, len(tenantInputs))
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

// Update updates tenant
func (s *service) Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error {
	tenant := tenantInput.ToBusinessTenantMapping(id)

	if err := s.tenantMappingRepo.Update(ctx, tenant); err != nil {
		return errors.Wrapf(err, "while updating tenant with id %s", id)
	}

	return nil
}

// GetCustomerIDParentRecursively gets the top parent external ID (customer_id) for a given tenant
func (s *service) GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error) {
	return s.tenantMappingRepo.GetCustomerIDParentRecursively(ctx, tenantID)
}

// CreateTenantAccessForResource creates a tenant access for a single resource.Type
func (s *service) CreateTenantAccessForResource(ctx context.Context, tenantID, resourceID string, isOwner bool, resourceType resource.Type) error {
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %q does not have access table", resourceType)
	}

	ta := &repo.TenantAccess{
		TenantID:   tenantID,
		ResourceID: resourceID,
		Owner:      isOwner,
	}

	if err := repo.CreateSingleTenantAccess(ctx, m2mTable, ta); err != nil {
		return errors.Wrapf(err, "while creating tenant acccess for resource type %q with ID %q for tenant %q", string(resourceType), ta.ResourceID, ta.TenantID)
	}

	return nil
}

// ListByParentAndType list tenants by parent ID and tenant.Type
func (s *service) ListByParentAndType(ctx context.Context, parentID string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByParentAndType(ctx, parentID, tenantType)
}

// CreateManyIfNotExists creates all provided tenants if they do not exist.
// It creates or updates the subdomain and region labels of the provided tenants, no matter if they are pre-existing or not.
func (s *labeledService) CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) ([]string, error) {
	return s.upsertTenants(ctx, tenantInputs, s.tenantMappingRepo.UnsafeCreate)
}

// UpsertMany creates all provided tenants if they do not exist. If they do exist, they are internally updated.
// It creates or updates the subdomain and region labels of the provided tenants, no matter if they are pre-existing or not.
func (s *labeledService) UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) ([]string, error) {
	return s.upsertTenants(ctx, tenantInputs, s.tenantMappingRepo.Upsert)
}

// UpsertSingle creates a provided tenant if it does not exist. If it does exist, it is internally updated.
// It creates or updates the subdomain and region labels of the provided tenant, no matter if it is pre-existing or not.
func (s *labeledService) UpsertSingle(ctx context.Context, tenantInput model.BusinessTenantMappingInput) (string, error) {
	return s.upsertTenant(ctx, tenantInput, s.tenantMappingRepo.Upsert)
}

func (s *labeledService) upsertTenant(ctx context.Context, tenantInput model.BusinessTenantMappingInput, upsertFunc func(context.Context, model.BusinessTenantMapping) error) (string, error) {
	id := s.uidService.Generate()
	tenant := *tenantInput.ToBusinessTenantMapping(id)
	subdomains, regions := tenantLocality([]model.BusinessTenantMappingInput{tenantInput})

	subdomain := ""
	region := ""
	if s, ok := subdomains[tenant.ExternalTenant]; ok {
		subdomain = s
	}
	if r, ok := regions[tenant.ExternalTenant]; ok {
		region = r
	}

	tenantID, err := s.createIfNotExists(ctx, tenant, subdomain, region, upsertFunc)
	if err != nil {
		return "", errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
	}

	return tenantID, nil
}

func (s *labeledService) upsertTenants(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput, upsertFunc func(context.Context, model.BusinessTenantMapping) error) ([]string, error) {
	tenants := s.MultipleToTenantMapping(tenantInputs)
	subdomains, regions := tenantLocality(tenantInputs)
	tenantIDs := make([]string, 0, len(tenants))

	for tenantIdx, tenant := range tenants {
		subdomain := ""
		region := ""
		if s, ok := subdomains[tenant.ExternalTenant]; ok {
			subdomain = s
		}
		if r, ok := regions[tenant.ExternalTenant]; ok {
			region = r
		}
		tenantID, err := s.createIfNotExists(ctx, tenant, subdomain, region, upsertFunc)
		if err != nil {
			return nil, errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
		}
		// the tenant already exists in our DB with a different ID, and we should update all child resources to use the correct internal ID
		tenantIDs = append(tenantIDs, tenantID)
		if tenantID != tenant.ID {
			for i := tenantIdx; i < len(tenants); i++ {
				if tenants[i].Parent == tenant.ID {
					tenants[i].Parent = tenantID
				}
			}
		}
	}

	return tenantIDs, nil
}

func (s *labeledService) createIfNotExists(ctx context.Context, tenant model.BusinessTenantMapping, subdomain, region string, action func(context.Context, model.BusinessTenantMapping) error) (string, error) {
	if err := action(ctx, tenant); err != nil {
		return "", err
	}

	tenantFromDB, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenant.ExternalTenant)
	if err != nil {
		return "", errors.Wrapf(err, "while retrieving the internal tenant ID of tenant with external ID %s", tenant.ExternalTenant)
	}

	return tenantFromDB.ID, s.upsertLabels(ctx, tenantFromDB.ID, subdomain, region)
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

// DeleteMany removes all tenants with the provided external tenant ids from the Compass storage.
func (s *service) DeleteMany(ctx context.Context, externalTenantIDs []string) error {
	for _, externalTenantID := range externalTenantIDs {
		err := s.tenantMappingRepo.DeleteByExternalTenant(ctx, externalTenantID)
		if err != nil {
			return errors.Wrap(err, "while deleting tenant")
		}
	}

	return nil
}

// ListLabels returns all labels directly linked to the given tenant, like subdomain or region.
// That excludes labels of other resource types in the context of the given tenant, for example labels of an application in the given tenant - those labels are not returned.
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
