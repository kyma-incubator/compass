package tenant

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"k8s.io/utils/strings/slices"

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
	// LicenseTypeLabelKey is the key of the tenant label for licensetype.
	LicenseTypeLabelKey = "licensetype"
	// CustomerIDLabelKey is the key of the SAP-managed customer subaccounts
	CustomerIDLabelKey = "customerId"
	// CostObjectIDLabelKey is the key for cost object tenant ID
	CostObjectIDLabelKey = "costObjectId"
	// CostObjectTypeLabelKey is the key for cost object tenant type
	CostObjectTypeLabelKey = "costObjectType"
)

// TenantMappingRepository is responsible for the repo-layer tenant operations.
//
//go:generate mockery --name=TenantMappingRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantMappingRepository interface {
	UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) (string, error)
	Upsert(ctx context.Context, item model.BusinessTenantMapping) (string, error)
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
	ListByType(ctx context.Context, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error)
	ListByIds(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error)
	ListByIdsAndType(ctx context.Context, ids []string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error)
	GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error)
}

// LabelUpsertService is responsible for creating, or updating already existing labels, and their label definitions.
//
//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelUpsertService interface {
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

// LabelRepository is responsible for the repo-layer label operations.
//
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// UIDService is responsible for generating GUIDs, which will be used as internal tenant IDs when tenants are created.
//
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
	converter         BusinessTenantMappingConverter
}

// NewService returns a new object responsible for service-layer tenant operations.
func NewService(tenantMapping TenantMappingRepository, uidService UIDService, converter BusinessTenantMappingConverter) *service {
	return &service{
		uidService:        uidService,
		tenantMappingRepo: tenantMapping,
		converter:         converter,
	}
}

// NewServiceWithLabels returns a new entity responsible for service-layer tenant operations, including operations with labels like listing all labels related to the given tenant.
func NewServiceWithLabels(tenantMapping TenantMappingRepository, uidService UIDService, labelRepo LabelRepository, labelUpsertSvc LabelUpsertService, converter BusinessTenantMappingConverter) *labeledService {
	return &labeledService{
		service: service{
			uidService:        uidService,
			tenantMappingRepo: tenantMapping,
			converter:         converter,
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

// ListByType returns all tenants for a provided Type.
func (s *service) ListByType(ctx context.Context, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByType(ctx, tenantType)
}

// ListByIDs returns all tenants with id in ids.
func (s *service) ListByIDs(ctx context.Context, ids []string) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByIds(ctx, ids)
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
func (s *service) MultipleToTenantMapping(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput) ([]model.BusinessTenantMapping, error) {
	tenants := make([]model.BusinessTenantMapping, 0, len(tenantInputs))
	tenantIDs := make(map[string]string, len(tenantInputs))
	for _, tenant := range tenantInputs {
		id := s.uidService.Generate()
		tenants = append(tenants, *tenant.ToBusinessTenantMapping(id))
		tenantIDs[tenant.ExternalTenant] = id
	}
	for i := 0; i < len(tenants); i++ { // Convert parent ID from external to internal id reference
		parentInternalIDs := make([]string, 0, len(tenants[i].Parents))

		for _, parentID := range tenants[i].Parents {
			if parentID == "" {
				continue
			}
			if parentInternalID, ok := tenantIDs[parentID]; ok { // If the parent is inserted in this request
				parentInternalIDs = append(parentInternalIDs, parentInternalID)
			} else { // If the parent is already present in the DB - swap the external ID for the parent that is provided with the internal ID from the DB
				internalPrentID, err := s.GetInternalTenant(ctx, parentID)
				if err != nil {
					return nil, errors.Wrapf(err, "while getting internal tenant: %s", parentID)
				}
				parentInternalIDs = append(parentInternalIDs, internalPrentID)
			}
		}
		tenants[i].Parents = parentInternalIDs
	}

	for i := 0; i < len(tenants); i++ { // Convert parent ID from external to internal id reference
		tenantID := tenants[i].ID
		for _, parentID := range tenants[i].Parents {
			var moved bool
			tenants, moved = MoveBeforeIfShould(tenants, parentID, tenantID) // Move my parent before me (to be inserted first) if it is not already
			if moved && i >= 0 {                                             // In case the added tenant is first, and it has more than one parent inserted with this request `i` may end up being negative number on the next iteration of the loop so decrease `i` only if it is non-negative
				i-- // Process the moved parent as well
			}
		}
	}
	return tenants, nil
}

// Update updates tenant
func (s *service) Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error {
	tenant := tenantInput.ToBusinessTenantMapping(id)

	if err := s.tenantMappingRepo.Update(ctx, tenant); err != nil {
		return errors.Wrapf(err, "while updating tenant with id %s", id)
	}

	return nil
}

// GetParentsRecursivelyByExternalTenant gets the top parents for a given external tenant
func (s *service) GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, externalTenant)
}

// CreateTenantAccessForResource creates a tenant access for a single resource.Type
func (s *service) CreateTenantAccessForResource(ctx context.Context, tenantAccess *model.TenantAccess) error {
	resourceType := tenantAccess.ResourceType
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %q does not have access table", resourceType)
	}

	ta := s.converter.TenantAccessToEntity(tenantAccess)

	if err := repo.CreateSingleTenantAccess(ctx, m2mTable, ta); err != nil {
		return errors.Wrapf(err, "while creating tenant acccess for resource type %q with ID %q for tenant %q", string(resourceType), ta.ResourceID, ta.TenantID)
	}

	return nil
}

// CreateTenantAccessForResourceRecursively creates a tenant access for a single resource.Type recursively
func (s *service) CreateTenantAccessForResourceRecursively(ctx context.Context, tenantAccess *model.TenantAccess) error {
	resourceType := tenantAccess.ResourceType
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %q does not have access table", resourceType)
	}

	ta := s.converter.TenantAccessToEntity(tenantAccess)

	if err := repo.CreateTenantAccessRecursively(ctx, m2mTable, ta); err != nil {
		return errors.Wrapf(err, "while creating tenant acccess for resource type %q with ID %q for tenant %q", string(resourceType), ta.ResourceID, ta.TenantID)
	}

	return nil
}

// DeleteTenantAccessForResourceRecursively deletes a tenant access for a single resource.Type recursively
func (s *service) DeleteTenantAccessForResourceRecursively(ctx context.Context, tenantAccess *model.TenantAccess) error {
	resourceType := tenantAccess.ResourceType
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %q does not have access table", resourceType)
	}

	ta := s.converter.TenantAccessToEntity(tenantAccess)

	if err := repo.DeleteTenantAccessRecursively(ctx, m2mTable, tenantAccess.InternalTenantID, []string{tenantAccess.ResourceID}, tenantAccess.InternalTenantID); err != nil {
		return errors.Wrapf(err, "while deleting tenant acccess for resource type %q with ID %q for tenant %q", string(resourceType), ta.ResourceID, ta.TenantID)
	}

	btm, err := s.tenantMappingRepo.GetByExternalTenant(ctx, tenantAccess.ExternalTenantID)
	if err != nil {
		return err
	}

	if IsAtomTenant(btm.Type) {
		rootTenants, err := s.tenantMappingRepo.GetParentsRecursivelyByExternalTenant(ctx, tenantAccess.ExternalTenantID)
		if err != nil {
			return err
		}

		rootTenantIDs := make([]string, 0, len(rootTenants))
		for _, rootTenant := range rootTenants {
			rootTenantIDs = append(rootTenantIDs, rootTenant.ID)
		}

		return repo.DeleteTenantAccessFromDirective(ctx, m2mTable, []string{tenantAccess.ResourceID}, rootTenantIDs)
	}

	return nil
}

// GetTenantAccessForResource gets a tenant access record for the specified resource
func (s *service) GetTenantAccessForResource(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) (*model.TenantAccess, error) {
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return nil, errors.Errorf("entity %q does not have access table", resourceType)
	}

	ta, err := repo.GetSingleTenantAccess(ctx, m2mTable, tenantID, resourceID)
	if err != nil {
		return nil, err
	}

	tenantAccessModel := s.converter.TenantAccessFromEntity(ta)
	tenantAccessModel.ResourceType = resourceType

	return tenantAccessModel, nil
}

// ListByParentAndType list tenants by parent ID and tenant.Type
func (s *service) ListByParentAndType(ctx context.Context, parentID string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByParentAndType(ctx, parentID, tenantType)
}

// ListByIDsAndType list tenants by IDs and tenant.Type
func (s *service) ListByIDsAndType(ctx context.Context, ids []string, tenantType tenantpkg.Type) ([]*model.BusinessTenantMapping, error) {
	return s.tenantMappingRepo.ListByIdsAndType(ctx, ids, tenantType)
}

// ExtractTenantIDForTenantScopedFormationTemplates returns the tenant ID based on its type:
//  1. If it's a SA -> return its parent GA id
//  2. If it's any other tenant type -> return its ID
func (s *service) ExtractTenantIDForTenantScopedFormationTemplates(ctx context.Context) (string, error) {
	internalTenantID, err := s.getTenantFromContext(ctx)
	if err != nil {
		return "", err
	}

	if internalTenantID == "" {
		return "", nil
	}

	tenantObject, err := s.GetTenantByID(ctx, internalTenantID)
	if err != nil {
		return "", err
	}

	if tenantObject.Type == tenantpkg.Subaccount {
		for _, parent := range tenantObject.Parents {
			tnt, err := s.GetTenantByID(ctx, parent)
			if err != nil {
				return "", err
			}
			if tnt.Type == tenantpkg.Account {
				return parent, nil
			}
		}
		return "", errors.Errorf("unexpected error. Tenant with id %s must have parent of type %s", internalTenantID, tenantpkg.Account)
	}

	return tenantObject.ID, nil
}

// getTenantFromContext validates and returns the tenant present in the context:
//   - if both internalID and externalID are present -> proceed with tenant scoped formation templates (return the internalID from ctx)
//   - if both internalID and externalID are NOT present -> -> proceed with global formation templates (return empty id)
//   - otherwise return TenantNotFoundError
func (s *service) getTenantFromContext(ctx context.Context) (string, error) {
	tntCtx, err := LoadTenantPairFromContextNoChecks(ctx)
	if err != nil {
		return "", err
	}

	if tntCtx.InternalID != "" && tntCtx.ExternalID != "" {
		return tntCtx.InternalID, nil
	}

	if tntCtx.InternalID == "" && tntCtx.ExternalID == "" {
		return "", nil
	}

	return "", apperrors.NewTenantNotFoundError(tntCtx.ExternalID)
}

// CreateManyIfNotExists creates all provided tenants if they do not exist.
// It creates or updates the subdomain, region, and customerId labels of the provided tenants, no matter if they are pre-existing or not.
func (s *labeledService) CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) (map[string]tenantpkg.Type, error) {
	return s.upsertTenants(ctx, tenantInputs, s.tenantMappingRepo.UnsafeCreate)
}

// UpsertMany creates all provided tenants if they do not exist. If they do exist, they are internally updated.
// It creates or updates the subdomain, region, and customerId labels of the provided tenants, no matter if they are pre-existing or not.
func (s *labeledService) UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) (map[string]tenantpkg.Type, error) {
	return s.upsertTenants(ctx, tenantInputs, s.tenantMappingRepo.Upsert)
}

// UpsertSingle creates a provided tenant if it does not exist. If it does exist, it is internally updated.
// It creates or updates the subdomain, region, and customerId labels of the provided tenant, no matter if it is pre-existing or not.
func (s *labeledService) UpsertSingle(ctx context.Context, tenantInput model.BusinessTenantMappingInput) (string, error) {
	return s.upsertTenant(ctx, tenantInput, s.tenantMappingRepo.Upsert)
}

func (s *labeledService) upsertTenant(ctx context.Context, tenantInput model.BusinessTenantMappingInput, upsertFunc func(context.Context, model.BusinessTenantMapping) (string, error)) (string, error) {
	parents, err := s.ListsByExternalIDs(ctx, tenantInput.Parents)
	if err != nil {
		return "", errors.Wrap(err, "while listing tenants by external ids")
	}
	parentInternalIDs := make([]string, 0, len(parents))
	for _, parent := range parents {
		parentInternalIDs = append(parentInternalIDs, parent.ID)
	}
	tenantInput.Parents = parentInternalIDs

	id := s.uidService.Generate()
	tenant := *tenantInput.ToBusinessTenantMapping(id)
	tenantList := []model.BusinessTenantMappingInput{tenantInput}
	subdomains, regions := tenantLocality(tenantList)
	customerIDs := tenantCustomerIDs(tenantList)
	costObjectIDs := tenantCostObjectIDs(tenantList)
	costObjectTypes := tenantCostObjectTypes(tenantList)

	subdomain := ""
	region := ""
	customerID := ""
	costObjectID := ""
	costObjectType := ""

	if s, ok := subdomains[tenant.ExternalTenant]; ok {
		subdomain = s
	}
	if r, ok := regions[tenant.ExternalTenant]; ok {
		region = r
	}
	if id, ok := customerIDs[tenant.ExternalTenant]; ok {
		customerID = id
	}
	if id, ok := costObjectIDs[tenant.ExternalTenant]; ok {
		costObjectID = id
	}
	if t, ok := costObjectTypes[tenant.ExternalTenant]; ok {
		costObjectType = t
	}

	tenantID, err := s.createIfNotExists(ctx, tenant, subdomain, region, customerID, costObjectID, costObjectType, upsertFunc)
	if err != nil {
		return "", errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
	}

	return tenantID, nil
}

func (s *labeledService) upsertTenants(ctx context.Context, tenantInputs []model.BusinessTenantMappingInput, upsertFunc func(context.Context, model.BusinessTenantMapping) (string, error)) (map[string]tenantpkg.Type, error) {
	tenants, err := s.MultipleToTenantMapping(ctx, tenantInputs)
	if err != nil {
		return nil, err
	}

	subdomains, regions := tenantLocality(tenantInputs)
	customerIDs := tenantCustomerIDs(tenantInputs)
	costObjectIDs := tenantCostObjectIDs(tenantInputs)
	costObjectTypes := tenantCostObjectTypes(tenantInputs)

	tenantsMap := make(map[string]tenantpkg.Type)

	for tenantIdx, tenant := range tenants {
		subdomain := ""
		region := ""
		customerID := ""
		costObjectID := ""
		costObjectType := ""
		if s, ok := subdomains[tenant.ExternalTenant]; ok {
			subdomain = s
		}
		if r, ok := regions[tenant.ExternalTenant]; ok {
			region = r
		}
		if id, ok := customerIDs[tenant.ExternalTenant]; ok {
			customerID = id
		}
		if id, ok := costObjectIDs[tenant.ExternalTenant]; ok {
			costObjectID = id
		}
		if t, ok := costObjectTypes[tenant.ExternalTenant]; ok {
			costObjectType = t
		}

		tenantID, err := s.createIfNotExists(ctx, tenant, subdomain, region, customerID, costObjectID, costObjectType, upsertFunc)
		if err != nil {
			return nil, errors.Wrapf(err, "while creating tenant with external ID %s", tenant.ExternalTenant)
		}

		// the tenant already exists in our DB with a different ID, and we should update all child resources to use the correct internal ID
		tenantsMap[tenantID] = tenant.Type
		if tenantID != tenant.ID {
			for i := tenantIdx; i < len(tenants); i++ {
				if slices.Contains(tenants[i].Parents, tenant.ID) {
					// remove tenant.ID from the parents array and replace it with the id returned from the DB - tenantID
					tenants[i].Parents = slices.Filter(nil, tenants[i].Parents, func(s string) bool {
						return s != tenant.ID
					})
					tenants[i].Parents = append(tenants[i].Parents, tenantID)
				}
			}
		}
	}

	return tenantsMap, nil
}

func (s *labeledService) createIfNotExists(ctx context.Context, tenant model.BusinessTenantMapping, subdomain, region, customerID, costObjectID, costObjectType string, action func(context.Context, model.BusinessTenantMapping) (string, error)) (string, error) {
	internalID, err := action(ctx, tenant)
	if err != nil {
		return "", err
	}

	return internalID, s.upsertLabels(ctx, internalID, subdomain, region, str.PtrStrToStr(tenant.LicenseType), customerID, costObjectID, costObjectType)
}

func (s *labeledService) upsertLabels(ctx context.Context, tenantID, subdomain, region, licenseType, customerID, costObjectID, costObjectType string) error {
	labelKeyValueMappings := map[string]string{
		SubdomainLabelKey:      subdomain,
		RegionLabelKey:         region,
		LicenseTypeLabelKey:    licenseType,
		CustomerIDLabelKey:     customerID,
		CostObjectIDLabelKey:   costObjectID,
		CostObjectTypeLabelKey: costObjectType,
	}

	for labelKey, labelValue := range labelKeyValueMappings {
		if len(labelValue) > 0 {
			if err := s.UpsertLabel(ctx, tenantID, labelKey, labelValue); err != nil {
				return errors.Wrapf(err, "while setting %q label for tenant with ID %s", labelKey, tenantID)
			}
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

func tenantCustomerIDs(tenants []model.BusinessTenantMappingInput) map[string]string {
	customerIDs := make(map[string]string)
	for _, t := range tenants {
		if t.CustomerID != nil {
			customerIDs[t.ExternalTenant] = str.PtrStrToStr(t.CustomerID)
		}
	}

	return customerIDs
}

func tenantCostObjectIDs(tenants []model.BusinessTenantMappingInput) map[string]string {
	costObjectIDs := make(map[string]string)
	for _, t := range tenants {
		if t.CostObjectID != nil {
			costObjectIDs[t.ExternalTenant] = str.PtrStrToStr(t.CostObjectID)
		}
	}

	return costObjectIDs
}

func tenantCostObjectTypes(tenants []model.BusinessTenantMappingInput) map[string]string {
	costObjectTypes := make(map[string]string)
	for _, t := range tenants {
		if t.CostObjectType != nil {
			costObjectTypes[t.ExternalTenant] = str.PtrStrToStr(t.CostObjectType)
		}
	}

	return costObjectTypes
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
	if err := s.Exists(ctx, tenantID); err != nil {
		return nil, err
	}

	labels, err := s.labelRepo.ListForObject(ctx, tenantID, model.TenantLabelableObject, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "whilie listing labels for tenant with ID %s", tenantID)
	}

	return labels, nil
}

// UpsertLabel upserts label that is directly linked to the provided tenant
func (s *labeledService) UpsertLabel(ctx context.Context, tenantID, key string, value interface{}) error {
	label := &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectID:   tenantID,
		ObjectType: model.TenantLabelableObject,
	}
	return s.labelUpsertSvc.UpsertLabel(ctx, tenantID, label)
}

// Exists checks if tenant with the provided internal ID exists in the Compass storage.
func (s *service) Exists(ctx context.Context, id string) error {
	exists, err := s.tenantMappingRepo.Exists(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while checking if tenant with ID %s exists", id)
	}

	if !exists {
		return apperrors.NewNotFoundError(resource.Tenant, id)
	}

	return nil
}

// ExistsByExternalTenant checks if tenant with the provided external ID exists in the Compass storage.
func (s *service) ExistsByExternalTenant(ctx context.Context, externalTenant string) error {
	exists, err := s.tenantMappingRepo.ExistsByExternalTenant(ctx, externalTenant)
	if err != nil {
		return errors.Wrapf(err, "while checking if tenant with External Tenant %s exists", externalTenant)
	}

	if !exists {
		return apperrors.NewNotFoundError(resource.Tenant, externalTenant)
	}

	return nil
}

// MoveBeforeIfShould moves the tenant with id right before index only if it is not already before it
func MoveBeforeIfShould(tenants []model.BusinessTenantMapping, parentTenantID, childTenantID string) ([]model.BusinessTenantMapping, bool) {
	var childTenantIndex int
	for i, tenant := range tenants {
		if tenant.ID == childTenantID {
			childTenantIndex = i
		}
	}

	var parentTenantIndex int
	for i, tenant := range tenants {
		if tenant.ID == parentTenantID {
			parentTenantIndex = i
		}
	}

	if parentTenantIndex <= childTenantIndex { // the parent tenant is already before the child tenant
		return tenants, false
	}

	newTenants := make([]model.BusinessTenantMapping, 0, len(tenants))
	for i := range tenants {
		if i == parentTenantIndex {
			continue // skip adding the parent tenant to the new tenants here at it was already placed right before its child tenant
		}
		if i == childTenantIndex {
			newTenants = append(newTenants, tenants[parentTenantIndex], tenants[i])
			continue
		}
		newTenants = append(newTenants, tenants[i])
	}
	return newTenants, true
}

// IsAtomTenant checks whether the tenant comes from atom
func IsAtomTenant(tenantType tenantpkg.Type) bool {
	if tenantType == tenantpkg.ResourceGroup || tenantType == tenantpkg.Folder || tenantType == tenantpkg.Organization {
		return true
	}

	return false
}
