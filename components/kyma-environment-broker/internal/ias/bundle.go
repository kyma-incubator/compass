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
	DeleteSecret(DeleteSecrets) error
	GenerateServiceProviderSecret(SecretConfiguration) (*ServiceProviderSecret, error)
	AuthenticationURL(ProviderID) string
	SetOIDCConfiguration(string, OIDCType) error
	SetSAMLConfiguration(string, SAMLType) error
	SetAssertionAttribute(string, PostAssertionAttributes) error
	SetSubjectNameIdentifier(string, SubjectNameIdentifier) error
	SetAuthenticationAndAccess(string, AuthenticationAndAccess) error
	SetDefaultAuthenticatingIDP(DefaultAuthIDPConfig) error
}

type ServiceProviderBundle struct {
	client                IASCLient
	config                Config
	serviceProvider       ServiceProvider
	serviceProviderExist  bool
	serviceProviderName   string
	serviceProviderParams ServiceProviderParam
	providerID            ProviderID
	organization          string
}

// NewServiceProviderBundle returns pointer to new ServiceProviderBundle
func NewServiceProviderBundle(bundleIdentifier string, spParams ServiceProviderParam, c IASCLient, cfg Config) *ServiceProviderBundle {
	return &ServiceProviderBundle{
		client:                c,
		config:                cfg,
		serviceProviderParams: spParams,
		serviceProviderName:   fmt.Sprintf("SKR %s (instanceID: %s)", strings.Title(spParams.domain), bundleIdentifier),
		organization:          "global",
	}
}

// ServiceProviderName returns SP name which includes instance ID
func (b *ServiceProviderBundle) ServiceProviderName() string {
	return b.serviceProviderName
}

// ServiceProviserType returns SSO type (SAML or OIDC)
func (b *ServiceProviderBundle) ServiceProviderType() string {
	return b.serviceProviderParams.ssoType
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

func (b *ServiceProviderBundle) configureServiceProviderOIDCType(serviceProviderName string, redirectURI string) error {
	iasType := OIDCType{
		ServiceProviderName: serviceProviderName,
		SsoType:             b.serviceProviderParams.ssoType,
		OpenIDConnectConfig: OpenIDConnectConfig{
			RedirectURIs: []string{redirectURI},
		},
	}

	return b.client.SetOIDCConfiguration(b.serviceProvider.ID, iasType)
}

func (b *ServiceProviderBundle) configureServiceProviderSAMLType(serviceProviderName string, redirectURI string) error {
	iasType := SAMLType{
		ServiceProviderName: serviceProviderName,
		ACSEndpoints: []ACSEndpoint{
			{
				Location:  redirectURI,
				Index:     0,
				IsDefault: true,
			},
		},
	}

	return b.client.SetSAMLConfiguration(b.serviceProvider.ID, iasType)
}

// ConfigureServiceProviderType sets SSO type, name and URLs based on provided URL for ServiceProvider
func (b *ServiceProviderBundle) ConfigureServiceProviderType(consolePath string) error {
	u, err := url.ParseRequestURI(consolePath)
	if err != nil {
		return errors.Wrap(err, "while parsing path for IAS Type")
	}
	serviceProviderDNS := strings.Replace(u.Host, "console.", fmt.Sprintf("%s.", b.serviceProviderParams.domain), 1)
	redirectURI := fmt.Sprintf("%s://%s%s", u.Scheme, serviceProviderDNS, b.serviceProviderParams.redirectPath)

	switch b.serviceProviderParams.ssoType {
	case SAML:
		err = b.configureServiceProviderSAMLType(serviceProviderDNS, redirectURI)
	case OIDC:
		err = b.configureServiceProviderOIDCType(serviceProviderDNS, redirectURI)
	default:
		err = errors.Errorf("Unrecognized ssoType: %s", b.serviceProviderParams.ssoType)
	}

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

	// set "DefaultAuthenticatingIDP"
	defaultAuthIDP := DefaultAuthIDPConfig{
		Organization:   b.organization,
		ID:             b.serviceProvider.ID,
		DefaultAuthIDP: b.client.AuthenticationURL(b.providerID),
	}
	err = b.client.SetDefaultAuthenticatingIDP(defaultAuthIDP)
	if err != nil {
		return errors.Wrap(err, "while configuring DefaultAuthenticatingIDP")
	}

	// set "AuthenticationAndAccess"
	if len(b.serviceProviderParams.allowedGroups) > 0 {
		authenticationAndAccess := AuthenticationAndAccess{
			ServiceProviderAccess: ServiceProviderAccess{
				RBAConfig: RBAConfig{
					RBARules:      make([]RBARules, len(b.serviceProviderParams.allowedGroups)),
					DefaultAction: "Deny",
				},
			},
		}
		for i, group := range b.serviceProviderParams.allowedGroups {
			authenticationAndAccess.ServiceProviderAccess.RBAConfig.RBARules[i] = RBARules{
				Action:    "Allow",
				Group:     group,
				GroupType: "Cloud",
			}
		}
		err = b.client.SetAuthenticationAndAccess(b.serviceProvider.ID, authenticationAndAccess)
		if err != nil {
			return errors.Wrap(err, "while configuring AuthenticationAndAccess")
		}
	}

	return nil
}

// GenerateSecret generates new ID and Secret for ServiceProvider, removes already existing secrets
func (b *ServiceProviderBundle) GenerateSecret() (*ServiceProviderSecret, error) {
	err := b.removeSecrets()
	if err != nil {
		return &ServiceProviderSecret{}, errors.Wrap(err, "while removing existing secrets")
	}

	secretCfg := SecretConfiguration{
		Organization: b.organization,
		ID:           b.serviceProvider.ID,
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

func (b *ServiceProviderBundle) removeSecrets() error {
	if len(b.serviceProvider.Secret) == 0 {
		return nil
	}

	var secretsIDs []string
	for _, s := range b.serviceProvider.Secret {
		secretsIDs = append(secretsIDs, s.SecretID)
	}

	deleteSecrets := DeleteSecrets{
		ClientID:         b.serviceProvider.UserForRest,
		ClientSecretsIDs: secretsIDs,
	}
	return b.client.DeleteSecret(deleteSecrets)
}
