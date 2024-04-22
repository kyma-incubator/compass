package tenantfetchersvc

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

const autogeneratedTenantProvider = "autogenerated"

// DirectorGraphQLClient expects graphql implementation
//
//go:generate mockery --name=DirectorGraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorGraphQLClient interface {
	WriteTenants(context.Context, []graphql.BusinessTenantMappingInput) error
	DeleteTenants(ctx context.Context, tenants []graphql.BusinessTenantMappingInput) error
	UpdateTenant(ctx context.Context, id string, tenant graphql.BusinessTenantMappingInput) error
	SubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionProviderAppName, subscriptionPayload string) error
	UnsubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionPayload string) error
	ExistsTenantByExternalID(ctx context.Context, tenantID string) (bool, error)
}

// TenantConverter expects tenant converter implementation
//
//go:generate mockery --name=TenantConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantConverter interface {
	MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput
	ToGraphQLInput(model.BusinessTenantMappingInput) graphql.BusinessTenantMappingInput
}

// TenantSubscriptionRequest represents the information provided during tenant provisioning request in Compass, which includes tenant IDs, subdomain, and region of the tenant.
// The tenant which triggered the provisioning request is only one, and one of the tenant IDs in the request is its external ID, where the other tenant IDs are external IDs from its parents hierarchy.
type TenantSubscriptionRequest struct {
	AccountTenantID             string
	SubaccountTenantID          string
	CustomerTenantID            string
	CostObjectTenantID          string
	Subdomain                   string
	Region                      string
	SubscriptionProviderID      string
	SubscriptionLcenseType      string
	ProviderSubaccountID        string
	ConsumerTenantID            string
	SubscriptionProviderAppName string
	SubscriptionPayload         string
}

// MainTenantID is used to determine the external tenant ID of the tenant for which the provisioning request was triggered.
func (r *TenantSubscriptionRequest) MainTenantID() string {
	if len(r.SubaccountTenantID) > 0 {
		return r.SubaccountTenantID
	}

	return r.AccountTenantID
}

type provisioner struct {
	gqlClient      DirectorGraphQLClient
	converter      TenantConverter
	tenantProvider string
}

// NewTenantProvisioner returns a TenantProvisioner initialized with the provided TenantService, and tenant provider.
// All tenants, created by the provisioner, besides the Customer ones, will have the value of tenantProvider as a provider.
func NewTenantProvisioner(directorClient DirectorGraphQLClient, tenantConverter TenantConverter, tenantProvider string) *provisioner {
	return &provisioner{
		gqlClient:      directorClient,
		converter:      tenantConverter,
		tenantProvider: tenantProvider,
	}
}

// ProvisionMissingTenants provisions tenants according to their type
func (p *provisioner) ProvisionMissingTenants(ctx context.Context, request *TenantSubscriptionRequest) error {
	var newBusinessTenantMappings = []model.BusinessTenantMappingInput{}
	requestedBusinessTenantMappingInputs := p.tenantsFromRequest(ctx, *request)
	for _, currentBusinessTenantMappingInput := range requestedBusinessTenantMappingInputs {
		log.C(ctx).Infof("ALEX ProvisionMissingTenants tenant %s", currentBusinessTenantMappingInput.ExternalTenant)
		exists, err := p.gqlClient.ExistsTenantByExternalID(ctx, currentBusinessTenantMappingInput.ExternalTenant)
		log.C(ctx).Infof("ALEX ProvisionMissingTenants %+v, %+v", exists, err)

		if err != nil {
			return err
		}
		if !exists {
			newBusinessTenantMappings = append(newBusinessTenantMappings, currentBusinessTenantMappingInput)
		}
	}

	if len(newBusinessTenantMappings) > 0 {
		log.C(ctx).Infof("ALEX ProvisionMissingTenants missing %d tenants %+v", len(newBusinessTenantMappings), newBusinessTenantMappings)
		tenantsToCreateGQL := p.converter.MultipleInputToGraphQLInput(newBusinessTenantMappings)
		return p.gqlClient.WriteTenants(ctx, tenantsToCreateGQL)
	}
	return nil
}

func (p *provisioner) tenantsFromRequest(ctx context.Context, request TenantSubscriptionRequest) []model.BusinessTenantMappingInput {
	tenants := make([]model.BusinessTenantMappingInput, 0, 3)
	customerID := request.CustomerTenantID
	accountID := request.AccountTenantID
	costObjectID := request.CostObjectTenantID

	log.C(ctx).Infof("ALEX Tenants retrieved from the subscription request. Customer id %q,account id %q, cost-object id %q", customerID, accountID, costObjectID)
	var licenseType *string

	if len(request.SubscriptionLcenseType) > 0 {
		licenseType = &request.SubscriptionLcenseType
	}

	if len(customerID) > 0 {
		tenants = append(tenants, p.newCustomerTenant(customerID, licenseType))
	}

	accountParent := []string{customerID}
	if len(costObjectID) > 0 {
		tenants = append(tenants, p.newCostObjectTenant(costObjectID, licenseType))
		accountParent = []string{costObjectID}
	}

	accountTenant := p.newAccountTenant(request.AccountTenantID, accountParent, request.Subdomain, request.Region, licenseType)
	if len(request.SubaccountTenantID) > 0 { // This means that the request is for Subaccount provisioning, therefore the subdomain and the region are for the subaccount and not for the GA
		accountTenant.Subdomain = ""
		accountTenant.Region = ""
	}
	tenants = append(tenants, accountTenant)

	if len(request.SubaccountTenantID) > 0 {
		tenants = append(tenants, p.newSubaccountTenant(request.SubaccountTenantID, []string{accountID}, request.Subdomain, request.Region, licenseType))
	}
	return tenants
}

func (p *provisioner) newCustomerTenant(tenantID string, licenseType *string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, []string{}, "", "", autogeneratedTenantProvider, licenseType, tenantEntity.Customer)
}

func (p *provisioner) newAccountTenant(tenantID string, parents []string, subdomain, region string, licenseType *string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, parents, subdomain, region, p.tenantProvider, licenseType, tenantEntity.Account)
}

func (p *provisioner) newSubaccountTenant(tenantID string, parents []string, subdomain, region string, licenseType *string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, parents, subdomain, region, p.tenantProvider, licenseType, tenantEntity.Subaccount)
}

func (p *provisioner) newCostObjectTenant(tenantID string, licenseType *string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, []string{}, "", "", autogeneratedTenantProvider, licenseType, tenantEntity.CostObject)
}

func (p *provisioner) newTenant(tenantID string, parents []string, subdomain, region, provider string, licenseType *string, tenantType tenantEntity.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           tenantID,
		ExternalTenant: tenantID,
		Parents:        parents,
		Subdomain:      subdomain,
		Region:         region,
		Type:           tenantEntity.TypeToStr(tenantType),
		Provider:       provider,
		LicenseType:    licenseType,
	}
}
