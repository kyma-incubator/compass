package ias

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceProviderBundle_ServiceProviderType(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	// when
	ssoType := bundle.ServiceProviderType()

	// then
	assert.Equal(t, OIDC, ssoType)
}

func TestServiceProviderBundle_FetchServiceProviderData(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	// when
	err := bundle.FetchServiceProviderData()

	// then
	assert.NoError(t, err)
	assert.True(t, bundle.ServiceProviderExist())
	assert.Equal(t, ProviderID(FakeIdentityProviderID), bundle.providerID)
}

func TestServiceProviderBundle_CreateServiceProvider(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle("sp", ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	// when
	err := bundle.CreateServiceProvider()

	// then
	assert.NoError(t, err)

	err = bundle.FetchServiceProviderData()
	assert.NoError(t, err)
	assert.True(t, bundle.ServiceProviderExist())
}

func TestServiceProviderBundle_ConfigureServiceProviderType_OIDC(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	err := bundle.FetchServiceProviderData()
	assert.NoError(t, err)

	// when
	err = bundle.ConfigureServiceProviderType("https://console.example.com")

	// then
	assert.NoError(t, err)
	provider, err := client.GetServiceProvider(FakeGrafanaID)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("SKR Grafana (instanceID: %s)", FakeGrafanaName), provider.DisplayName)
	assert.Equal(t, "openIdConnect", provider.SsoType)
	assert.Equal(t, "https://grafana.example.com/login/generic_oauth", provider.RedirectURIs[0])
}

func TestServiceProviderBundle_ConfigureServiceProviderType_SAML(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeDexName, ServiceProviderInputs[SPDexID], client, Config{IdentityProvider: FakeIdentityProviderName})

	err := bundle.FetchServiceProviderData()
	assert.NoError(t, err)

	// when
	err = bundle.ConfigureServiceProviderType("https://console.example.com")

	// then
	assert.NoError(t, err)
	provider, err := client.GetServiceProvider(FakeDexID)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("SKR Dex (instanceID: %s)", FakeDexName), provider.DisplayName)
	assert.Equal(t, "https://dex.example.com/callback", provider.ACSEndpoints[0].Location)
	assert.Equal(t, int32(0), provider.ACSEndpoints[0].Index)
	assert.Equal(t, true, provider.ACSEndpoints[0].IsDefault)
}

func TestServiceProviderBundle_ConfigureServiceProvider(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	err := bundle.FetchServiceProviderData()
	assert.NoError(t, err)

	// when
	err = bundle.ConfigureServiceProvider()

	// then
	assert.NoError(t, err)
	provider, err := client.GetServiceProvider(FakeGrafanaID)
	assert.NoError(t, err)

	assert.Len(t, provider.AssertionAttributes, 4)
	assert.ElementsMatch(t, []AssertionAttribute{
		{AssertionAttribute: "first_name", UserAttribute: "firstName"},
		{AssertionAttribute: "last_name", UserAttribute: "lastName"},
		{AssertionAttribute: "email", UserAttribute: "mail"},
		{AssertionAttribute: "groups", UserAttribute: "companyGroups"},
	}, provider.AssertionAttributes)

	assert.Equal(t, "mail", provider.NameIDAttribute)

	assert.Equal(t, FakeIdentityProviderID, provider.AuthenticatingIdp.ID)
	assert.Equal(t, FakeIdentityProviderName, provider.AuthenticatingIdp.Name)

	assert.Len(t, provider.RBAConfig.RBARules, 2)
	assert.ElementsMatch(t, []RBARules{
		{Action: "Allow", Group: "skr-monitoring-admin", GroupType: "Cloud"},
		{Action: "Allow", Group: "skr-monitoring-viewer", GroupType: "Cloud"},
	}, provider.RBAConfig.RBARules)
	assert.Equal(t, "Deny", provider.RBAConfig.DefaultAction)
}

func TestServiceProviderBundle_GenerateSecret(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	err := bundle.FetchServiceProviderData()
	assert.NoError(t, err)

	// when
	secret, err := bundle.GenerateSecret()

	// then
	assert.NoError(t, err)
	assert.Equal(t, FakeClientID, secret.ClientID)
	assert.Equal(t, FakeClientSecret, secret.ClientSecret)

	provider, err := client.GetServiceProvider(FakeGrafanaID)
	assert.NoError(t, err)
	assert.Len(t, provider.Secret, 1)
	assert.Equal(t, FakeClientID, provider.Secret[0].SecretID)
	assert.Equal(t, "SAP Kyma Runtime Secret", provider.Secret[0].Description)
	assert.ElementsMatch(t, []string{"ManageApp", "ManageUsers", "OAuth"}, provider.Secret[0].Scopes)

	// when
	err = bundle.FetchServiceProviderData()
	assert.NoError(t, err)
	secret, err = bundle.GenerateSecret()

	// then
	provider, err = client.GetServiceProvider(FakeGrafanaID)
	assert.NoError(t, err)
	assert.Len(t, provider.Secret, 1)
}

func TestServiceProviderBundle_DeleteServiceProvider(t *testing.T) {
	// given
	client := NewFakeClient()
	bundle := NewServiceProviderBundle(FakeGrafanaName, ServiceProviderInputs[SPGrafanaID], client, Config{IdentityProvider: FakeIdentityProviderName})

	// when
	err := bundle.DeleteServiceProvider()

	// then
	assert.NoError(t, err)
	provider, err := client.GetServiceProvider(FakeGrafanaID)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("cannot find ServiceProvider with ID: %s", FakeGrafanaID))
	assert.Nil(t, provider)
}
