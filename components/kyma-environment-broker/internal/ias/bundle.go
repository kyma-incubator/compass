package ias

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type (
	ProviderID string

	Config struct {
		URL              string
		UserSecret       string
		UserID           string
		IdentityProvider string
		Disabled         bool
	}
)

//go:generate mockery -name=IASCLient -output=automock -outpkg=automock -case=underscore
type IASCLient interface {
	GetCompany() (*Company, error)
	CreateServiceProvider(string, string) error
	DeleteServiceProvider(string) error
	GenerateServiceProviderSecret(SecretConfiguration) (*ServiceProviderSecret, error)
	AuthenticationURL(ProviderID) string
	SetType(string, Type) error
	SetAssertionAttribute(string, PostAssertionAttributes) error
	SetSubjectNameIdentifier(string, SubjectNameIdentifier) error
	SetAuthenticationAndAccess(string, AuthenticationAndAccess) error
}

type ServiceProviderBundle struct {
	client               IASCLient
	config               Config
	serviceProvider      ServiceProvider
	serviceProviderExist bool
	serviceProviderName  string
	providerID           ProviderID
	organization         string
	ssoType              string
}

// NewServiceProviderBundle returns pointer to new ServiceProviderBundle
func NewServiceProviderBundle(bundleIdentifier string, c IASCLient, cfg Config) *ServiceProviderBundle {
	return &ServiceProviderBundle{
		client:              c,
		config:              cfg,
		serviceProviderName: fmt.Sprintf("KymaRuntime (instanceID: %s)", bundleIdentifier),
		organization:        "global",
		ssoType:             "openIdConnect",
	}
}

// ServiceProviderName returns SP name which includes instance ID
func (b *ServiceProviderBundle) ServiceProviderName() string {
	return b.serviceProviderName
}

// FetchServiceProviderData fetches all ServiceProviders and IdentityProviders for company
// saves specific elements based on the name
func (b *ServiceProviderBundle) FetchServiceProviderData() error {
	company, err := b.client.GetCompany()
	if err != nil {
		return errors.Wrap(err, "while getting company")
	}

	for _, identifiers := range company.IdentityProviders {
		if identifiers.Name == b.config.IdentityProvider {
			b.providerID = ProviderID(identifiers.ID)
			break
		}
	}
	if b.providerID == "" {
		return errors.Errorf("provider ID for %s name does not exist", b.config.IdentityProvider)
	}

	for _, provider := range company.ServiceProviders {
		if provider.DisplayName == b.serviceProviderName {
			b.serviceProvider = provider
			b.serviceProviderExist = true
			break
		}
	}

	return nil
}

// ServiceProviderExist deteminates whether a particular item has been found
func (b *ServiceProviderBundle) ServiceProviderExist() bool {
	return b.serviceProviderExist
}

// CreateServiceProvider creates new ServiceProvider on IAS based on name
// it will be create in specific company/organization
func (b *ServiceProviderBundle) CreateServiceProvider() error {
	err := b.client.CreateServiceProvider(b.serviceProviderName, b.organization)
	if err != nil {
		return errors.Wrap(err, "while creating ServiceProvider")
	}
	err = b.FetchServiceProviderData()
	if err != nil {
		return errors.Wrap(err, "while fetching ServiceProvider")
	}

	return nil
}

// DeleteServiceProvider removes ServiceProvider from IAS
func (b *ServiceProviderBundle) DeleteServiceProvider() error {
	err := b.FetchServiceProviderData()
	if err != nil {
		return errors.Wrap(err, "while fetching ServiceProvider before deleting")
	}
	if !b.serviceProviderExist {
		return nil
	}

	err = b.client.DeleteServiceProvider(b.serviceProvider.ID)
	if err != nil {
		return errors.Wrap(err, "while deleting ServiceProvider")
	}

	return nil
}

// ConfigureServiceProviderType sets SSO type, name and URLs based on provided URL for ServiceProvider
func (b *ServiceProviderBundle) ConfigureServiceProviderType(consolePath string) error {
	u, err := url.ParseRequestURI(consolePath)
	if err != nil {
		return errors.Wrap(err, "while parsing path for IAS Type")
	}
	path := strings.Replace(u.Host, "console.", "grafana.", 1)

	iasType := Type{
		ServiceProviderName: path,
		SsoType:             b.ssoType,
		OpenIDConnectConfig: OpenIDConnectConfig{
			RedirectURIs: []string{fmt.Sprintf("%s://%s/login/generic_oauth", u.Scheme, path)},
		},
	}
	err = b.client.SetType(b.serviceProvider.ID, iasType)
	if err != nil {
		return errors.Wrap(err, "while configuring IAS Type")
	}

	return nil
}

// ConfigureServiceProvider sets configuration such as assertion attributes, name identifier and
// gropus allows to connect with specific ServiceProvider
func (b *ServiceProviderBundle) ConfigureServiceProvider() error {
	// set "AssertionAttributes"
	attributeDeliver := NewAssertionAttributeDeliver()
	sciAttributes := PostAssertionAttributes{
		AssertionAttributes: attributeDeliver.GenerateAssertionAttribute(b.serviceProvider),
	}
	err := b.client.SetAssertionAttribute(b.serviceProvider.ID, sciAttributes)
	if err != nil {
		return errors.Wrap(err, "while configuring AssertionAttributes")
	}

	// set "SubjectNameIdentifier"
	subjectNameIdentifier := SubjectNameIdentifier{
		NameIDAttribute: "mail",
	}
	err = b.client.SetSubjectNameIdentifier(b.serviceProvider.ID, subjectNameIdentifier)
	if err != nil {
		return errors.Wrap(err, "while configuring SubjectNameIdentifier")
	}

	// set "AuthenticationAndAccess"
	authenticationAndAccess := AuthenticationAndAccess{
		ServiceProviderAccess: ServiceProviderAccess{
			RBAConfig: RBAConfig{
				RBARules: []RBARules{
					{
						Action:    "Allow",
						Group:     "skr-monitoring-admin",
						GroupType: "Cloud",
					},
					{
						Action:    "Allow",
						Group:     "skr-monitoring-viewer",
						GroupType: "Cloud",
					},
				},
				DefaultAction: "Allow",
			},
		},
	}
	err = b.client.SetAuthenticationAndAccess(b.serviceProvider.ID, authenticationAndAccess)
	if err != nil {
		return errors.Wrap(err, "while configuring AuthenticationAndAccess")
	}

	return nil
}

// GenerateSecret generates new ID and Secret for ServiceProvider
func (b *ServiceProviderBundle) GenerateSecret() (*ServiceProviderSecret, error) {
	secretCfg := SecretConfiguration{
		Organization:   b.organization,
		ID:             b.serviceProvider.ID,
		DefaultAuthIDp: b.client.AuthenticationURL(b.providerID),
		RestAPIClientSecret: RestAPIClientSecret{
			Description: "SAP Kyma Runtime Secret",
			Scopes:      []string{"ManageApp", "ManageUsers", "OAuth"},
		},
	}

	sps, err := b.client.GenerateServiceProviderSecret(secretCfg)
	if err != nil {
		return &ServiceProviderSecret{}, errors.Wrap(err, "while creating ServiceProviderSecret")
	}

	return sps, nil
}
