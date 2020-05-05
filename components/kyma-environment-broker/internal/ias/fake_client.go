package ias

import "fmt"

const (
	FakeIdentityProviderName = "IdentityProviderName"
	FakeIdentityProviderID   = "0dbae593-ab1d-4774-97c1-5118ea22ea2d"
	FakeProviderName         = "ProviderName"
	FakeProviderID           = "eebb54dd-e4d5-43a1-929a-e98ea2831342"
	FakeClientID             = "cid"
	FakeClientSecret         = "csc"
)

type FakeClient struct {
	serviceProviders []*ServiceProvider
}

func NewFakeClient() *FakeClient {
	return &FakeClient{
		serviceProviders: []*ServiceProvider{
			{
				ID:          FakeProviderID,
				DisplayName: fmt.Sprintf("KymaRuntime (instanceID: %s)", FakeProviderName),
				AssertionAttributes: []AssertionAttribute{
					{
						AssertionAttribute: "test",
						UserAttribute:      "test",
					},
				},
			},
		},
	}
}

func (f *FakeClient) GetCompany() (*Company, error) {
	var sp []ServiceProvider
	for _, fsp := range f.serviceProviders {
		sp = append(sp, *fsp)
	}

	return &Company{
		ServiceProviders: sp,
		IdentityProviders: []IdentityProvider{
			{
				Name: FakeIdentityProviderName,
				ID:   FakeIdentityProviderID,
			},
		},
	}, nil
}

func (f *FakeClient) CreateServiceProvider(name string, _ string) error {
	f.serviceProviders = append(f.serviceProviders, &ServiceProvider{
		DisplayName: name,
	})

	return nil
}

func (f FakeClient) GenerateServiceProviderSecret(ss SecretConfiguration) (*ServiceProviderSecret, error) {
	serviceProvider, err := f.GetServiceProvider(ss.ID)
	if err != nil {
		return &ServiceProviderSecret{}, err
	}

	serviceProvider.Secret = append(serviceProvider.Secret, SPSecret{
		SecretID:    FakeClientID,
		Description: ss.RestAPIClientSecret.Description,
		Scopes:      ss.RestAPIClientSecret.Scopes,
	})
	serviceProvider.AuthenticatingIdp.ID = FakeIdentityProviderID
	serviceProvider.AuthenticatingIdp.Name = FakeIdentityProviderName

	return &ServiceProviderSecret{
		ClientID:     FakeClientID,
		ClientSecret: FakeClientSecret,
	}, nil
}

func (f FakeClient) AuthenticationURL(id ProviderID) string {
	return fmt.Sprintf("https://authentication.com/%s", id)
}

func (f *FakeClient) SetType(id string, iasType Type) error {
	serviceProvider, err := f.GetServiceProvider(id)
	if err != nil {
		return err
	}

	serviceProvider.SsoType = iasType.SsoType
	serviceProvider.DisplayName = iasType.ServiceProviderName
	serviceProvider.RedirectURIs = iasType.OpenIDConnectConfig.RedirectURIs

	return nil
}

func (f FakeClient) SetAssertionAttribute(id string, paa PostAssertionAttributes) error {
	serviceProvider, err := f.GetServiceProvider(id)
	if err != nil {
		return err
	}

	serviceProvider.AssertionAttributes = paa.AssertionAttributes

	return nil
}

func (f FakeClient) SetSubjectNameIdentifier(id string, sni SubjectNameIdentifier) error {
	serviceProvider, err := f.GetServiceProvider(id)
	if err != nil {
		return err
	}

	serviceProvider.NameIDAttribute = sni.NameIDAttribute

	return nil
}

func (f FakeClient) SetAuthenticationAndAccess(id string, auth AuthenticationAndAccess) error {
	serviceProvider, err := f.GetServiceProvider(id)
	if err != nil {
		return err
	}

	serviceProvider.RBAConfig = auth.ServiceProviderAccess.RBAConfig

	return nil
}

func (f *FakeClient) DeleteServiceProvider(id string) error {
	for index, sp := range f.serviceProviders {
		if sp.ID == id {
			f.serviceProviders[index] = f.serviceProviders[len(f.serviceProviders)-1]
			f.serviceProviders[len(f.serviceProviders)-1] = nil
			f.serviceProviders = f.serviceProviders[:len(f.serviceProviders)-1]
			return nil
		}
	}

	return nil
}

func (f *FakeClient) GetServiceProvider(id string) (*ServiceProvider, error) {
	for _, sp := range f.serviceProviders {
		if sp.ID == id {
			return sp, nil
		}
	}

	return nil, fmt.Errorf("cannot find ServiceProvider with ID: %s", id)
}
