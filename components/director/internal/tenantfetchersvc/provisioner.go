package tenantfetchersvc

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

const autogeneratedTenantProvider = "autogenerated"

// DirectorGraphQLClient expects graphql implementation
//go:generate mockery --name=DirectorGraphQLClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type DirectorGraphQLClient interface {
	WriteTenants(context.Context, []graphql.BusinessTenantMappingInput) error
	SubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, region, appName string) error
	UnsubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, region string) error
}

// TenantConverter expects tenant converter implementation
//go:generate mockery --name=TenantConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantConverter interface {
	MultipleInputToGraphQLInput([]model.BusinessTenantMappingInput) []graphql.BusinessTenantMappingInput
}

// TenantSubscriptionRequest represents the information provided during tenant provisioning request in Compass, which includes tenant IDs, subdomain, and region of the tenant.
// The tenant which triggered the provisioning request is only one, and one of the tenant IDs in the request is its external ID, where the other tenant IDs are external IDs from its parents hierarchy.
type TenantSubscriptionRequest struct {
	AccountTenantID        string
	SubaccountTenantID     string
	CustomerTenantID       string
	Subdomain              string
	Region                 string
	SubscriptionProviderID string
	ProviderSubaccountID   string
	SubscriptionAppName    string
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

// ProvisionTenants provisions tenants according to their type
func (p *provisioner) ProvisionTenants(ctx context.Context, request *TenantSubscriptionRequest) error {
	tenantsToCreateGQL := p.converter.MultipleInputToGraphQLInput(p.tenantsFromRequest(*request))
	return p.gqlClient.WriteTenants(ctx, tenantsToCreateGQL)
}

func (p *provisioner) tenantsFromRequest(request TenantSubscriptionRequest) []model.BusinessTenantMappingInput {
	tenants := make([]model.BusinessTenantMappingInput, 0, 3)
	customerID := request.CustomerTenantID
	accountID := request.AccountTenantID

	if len(request.CustomerTenantID) > 0 {
		tenants = append(tenants, p.newCustomerTenant(request.CustomerTenantID))
	}

	accountTenant := p.newAccountTenant(request.AccountTenantID, customerID, request.Subdomain, request.Region)
	if len(request.SubaccountTenantID) > 0 { // This means that the request is for Subaccount provisioning, therefore the subdomain and the region are for the subaccount and not for the GA
		accountTenant.Subdomain = ""
		accountTenant.Region = ""
	}
	tenants = append(tenants, accountTenant)

	if len(request.SubaccountTenantID) > 0 {
		tenants = append(tenants, p.newSubaccountTenant(request.SubaccountTenantID, accountID, request.Subdomain, request.Region))
	}
	return tenants
}

func (p *provisioner) newCustomerTenant(tenantID string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, "", "", "", autogeneratedTenantProvider, tenantEntity.Customer)
}

func (p *provisioner) newAccountTenant(tenantID, parent, subdomain, region string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, parent, subdomain, region, p.tenantProvider, tenantEntity.Account)
}

func (p *provisioner) newSubaccountTenant(tenantID, parent, subdomain, region string) model.BusinessTenantMappingInput {
	return p.newTenant(tenantID, parent, subdomain, region, p.tenantProvider, tenantEntity.Subaccount)
}

func (p *provisioner) newTenant(tenantID, parent, subdomain, region, provider string, tenantType tenantEntity.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           tenantID,
		ExternalTenant: tenantID,
		Parent:         parent,
		Subdomain:      subdomain,
		Region:         region,
		Type:           tenantEntity.TypeToStr(tenantType),
		Provider:       provider,
	}
}
