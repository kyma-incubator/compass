package subscription

import (
	"fmt"
	"strings"
)

const (
	TenantPathParamValue = "tenant"
	RegionPathParamValue = "eu-1"
	DefaultSubdomain     = "default-subdomain"
	AppName              = "app-name-default" // todo:: chech where is used
)

type Config struct {
	TenantFetcherURL                   string
	RootAPI                            string
	RegionalHandlerEndpoint            string
	TenantPathParam                    string
	RegionPathParam                    string
	SubscriptionProviderID             string
	TenantFetcherFullRegionalURL       string `envconfig:"-"`
	TestConsumerAccountID              string
	TestConsumerSubaccountID           string
	TestConsumerTenantID               string
	PropagatedProviderSubaccountHeader string
	SubscriptionProviderAppNameValue   string
	TestTenantOnDemandID               string
}

// ProviderConfig includes the configuration for tenant providers - the tenant ID json property names - account, subaccount, customer. The subdomain property name and subscription provider ID property.
type ProviderConfig struct {
	TenantIDProperty               string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIDProperty     string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIDProperty             string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty              string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
	SubscriptionProviderIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY"`
	ProviderSubaccountIDProperty   string `envconfig:"APP_TENANT_PROVIDER_PROVIDER_SUBACCOUNT_ID_PROPERTY"`
	SubscribedSubaccountIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIBED_SUBACCOUNT_ID_PROPERTY"` // todo:: check where is used and for what?
	ConsumerTenantIDProperty            string `envconfig:"APP_TENANT_PROVIDER_CONSUMER_TENANT_ID_PROPERTY"`
	SubscriptionProviderAppNameProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_APP_NAME_PROPERTY"`
}

func BuildTenantFetcherRegionalURL(tenantConfig *Config) {
	regionalEndpoint := strings.Replace(tenantConfig.RegionalHandlerEndpoint, fmt.Sprintf("{%s}", tenantConfig.TenantPathParam), TenantPathParamValue, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", tenantConfig.RegionPathParam), RegionPathParamValue, 1)
	tenantConfig.TenantFetcherFullRegionalURL = tenantConfig.TenantFetcherURL + tenantConfig.RootAPI + regionalEndpoint
}
