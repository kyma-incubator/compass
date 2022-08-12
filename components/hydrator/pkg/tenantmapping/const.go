package tenantmapping

const (
	// UserObjectContextProvider missing godoc
	UserObjectContextProvider = "UserObjectContextProvider"
	// SystemAuthObjectContextProvider missing godoc
	SystemAuthObjectContextProvider = "SystemAuthObjectContextProvider"
	// AuthenticatorObjectContextProvider missing godoc
	AuthenticatorObjectContextProvider = "AuthenticatorObjectContextProvider"
	// CertServiceObjectContextProvider missing godoc
	CertServiceObjectContextProvider = "CertServiceObjectContextProvider"
	// TenantHeaderObjectContextProvider missing godoc
	TenantHeaderObjectContextProvider = "TenantHeaderObjectContextProvider"
	// ConsumerProviderObjectContextProvider is an object context provider for the consumer-provider flow
	ConsumerProviderObjectContextProvider = "ConsumerProviderObjectContextProvider"
	// ConsumerTenantKey key for consumer tenant id in Claims.Tenant
	ConsumerTenantKey = "consumerTenant"
	// ExternalTenantKey key for external tenant id in Claims.Tenant
	ExternalTenantKey = "externalTenant"
	// ProviderTenantKey key for provider tenant id in Claims.Tenant
	ProviderTenantKey = "providerTenant"
	// ProviderExternalTenantKey key for external provider tenant id in Claims.Tenant
	ProviderExternalTenantKey = "providerExternalTenant"
)
