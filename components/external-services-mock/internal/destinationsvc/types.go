package destinationsvc

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config holds destination creator service API test configuration
type Config struct {
	CorrelationIDsKey string `envconfig:"APP_DESTINATION_CREATOR_CORRELATION_IDS_KEY"`
	*DestinationAPIConfig
	*CertificateAPIConfig
}

// DestinationAPIConfig holds a test configuration specific for the destination API of the destination creator service
type DestinationAPIConfig struct {
	SubaccountLevelPath  string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_PATH"`
	InstanceLevelPath    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_INSTANCE_LEVEL_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_REGION_PARAMETER"`
	InstanceIDParam      string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_INSTANCE_ID_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_SUBACCOUNT_ID_PARAMETER"`
	DestinationNameParam string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_NAME_PARAMETER"`
}

// CertificateAPIConfig holds a test configuration specific for the certificate API of the destination creator service
type CertificateAPIConfig struct {
	SubaccountLevelPath  string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_PATH"`
	InstanceLevelPath    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_INSTANCE_LEVEL_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_REGION_PARAMETER"`
	InstanceIDParam      string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_INSTANCE_ID_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_SUBACCOUNT_ID_PARAMETER"`
	CertificateNameParam string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_NAME_PARAMETER"`
}

type DestinationRequestBody interface {
	ToDestination() destinationcreator.Destination
	Validate() error
	GetDestinationType() string
	GetDestinationUniqueIdentifier(subaccountID, instanceID string) string
}

// BaseDestinationRequestBody contains the base fields needed in the destination request body
type BaseDestinationRequestBody struct {
	Name               string                          `json:"name"`
	URL                string                          `json:"url"`
	Type               destinationcreatorpkg.Type      `json:"type"`
	ProxyType          destinationcreatorpkg.ProxyType `json:"proxyType"`
	AuthenticationType destinationcreatorpkg.AuthType  `json:"authenticationType"`
	XCorrelationID     string                          `json:"x-correlation-id"` // old format
	CorrelationIds     string                          `json:"correlationIds"`   // new format
	XSystemTenantID    string                          `json:"x-system-id"`      // local tenant id
	XSystemTenantName  string                          `json:"x-system-name"`    // random or application name
	XSystemType        string                          `json:"x-system-type"`    // application type
}

// DesignTimeDestRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeNoAuth
type DesignTimeDestRequestBody struct {
	BaseDestinationRequestBody
}

// BasicDestRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeBasic
type BasicDestRequestBody struct {
	BaseDestinationRequestBody
	User     string `json:"user"`
	Password string `json:"password"`
}

// SAMLAssertionDestRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeSAMLAssertion
type SAMLAssertionDestRequestBody struct {
	BaseDestinationRequestBody
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// ClientCertificateAuthDestRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeClientCertificate
type ClientCertificateAuthDestRequestBody struct {
	BaseDestinationRequestBody
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// OAuth2ClientCredsDestRequestBody contains the necessary fields for the destination request body with authentication type OAuth2ClientCredentials
type OAuth2ClientCredsDestRequestBody struct {
	BaseDestinationRequestBody
	TokenServiceURL     string `json:"tokenServiceURL"`
	TokenServiceURLType string `json:"tokenServiceURLType"`
	ClientID            string `json:"clientId"`
	KeyStoreLocation    string `json:"tokenServiceKeystoreLocation"`
	ClientSecret        string `json:"clientSecret"`
}

// CertificateRequestBody contains the necessary fields for the destination creator certificate request body
type CertificateRequestBody struct {
	FileName string `json:"fileName"`
}

// CertificateResponseBody contains the response data from the destination creator certificate request
type CertificateResponseBody struct {
	FileName         string `json:"fileName"`
	CommonName       string `json:"commonName"`
	CertificateChain string `json:"certificateChain"`
}

// reqBodyNameRegex is a regex defined by the destination creator API specifying what destination names are allowed
var reqBodyNameRegex = "[a-zA-Z0-9_-]{1,64}"

func (b *BaseDestinationRequestBody) GetDestinationUniqueIdentifier(subaccountID, instanceID string) string {
	return fmt.Sprintf("name_%s_subacc_%s_instance_%s", b.Name, subaccountID, instanceID)
}

// Validate validates that the AuthTypeNoAuth request body contains the required fields, and they are valid
func (n *DesignTimeDestRequestBody) Validate() error {
	return validation.ValidateStruct(n,
		validation.Field(&n.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&n.URL, validation.Required),
		validation.Field(&n.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&n.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&n.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeNoAuth)),
	)
}

func (n *DesignTimeDestRequestBody) ToDestination() destinationcreator.Destination {
	correlationID := ""
	if n.XCorrelationID != "" {
		correlationID = n.XCorrelationID
	} else {
		correlationID = n.CorrelationIds
	}

	return &destinationcreator.NoAuthenticationDestination{
		Name:              n.Name,
		URL:               n.URL,
		Type:              n.Type,
		ProxyType:         n.ProxyType,
		Authentication:    n.AuthenticationType,
		XCorrelationID:    correlationID,
		XSystemTenantID:   n.XSystemTenantID,
		XSystemType:       n.XSystemType,
		XSystemTenantName: n.XSystemTenantName,
	}
}

func (n *DesignTimeDestRequestBody) GetDestinationType() string {
	return destinationcreator.DesignTimeDestinationType
}

// Validate validates that the AuthTypeBasic request body contains the required fields, and they are valid
func (b *BasicDestRequestBody) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeBasic)),
		validation.Field(&b.User, validation.Required, validation.Length(1, 256)),
	)
}

func (b *BasicDestRequestBody) ToDestination() destinationcreator.Destination {
	correlationID := ""
	if b.XCorrelationID != "" {
		correlationID = b.XCorrelationID
	} else {
		correlationID = b.CorrelationIds
	}

	return &destinationcreator.BasicDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:              b.Name,
			Type:              b.Type,
			URL:               b.URL,
			Authentication:    b.AuthenticationType,
			ProxyType:         b.ProxyType,
			XCorrelationID:    correlationID,
			XSystemTenantID:   b.XSystemTenantID,
			XSystemType:       b.XSystemType,
			XSystemTenantName: b.XSystemTenantName,
		},
		User:     b.User,
		Password: b.Password,
	}
}

func (b *BasicDestRequestBody) GetDestinationType() string {
	return destinationcreator.BasicAuthDestinationType
}

// Validate validates that the AuthTypeSAMLAssertion request body contains the required fields, and they are valid
func (s *SAMLAssertionDestRequestBody) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeSAMLAssertion)),
		validation.Field(&s.Audience, validation.Required),
		validation.Field(&s.KeyStoreLocation, validation.Required),
	)
}

func (s *SAMLAssertionDestRequestBody) ToDestination() destinationcreator.Destination {
	correlationID := ""
	if s.XCorrelationID != "" {
		correlationID = s.XCorrelationID
	} else {
		correlationID = s.CorrelationIds
	}

	return &destinationcreator.SAMLAssertionDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:              s.Name,
			Type:              s.Type,
			URL:               s.URL,
			Authentication:    s.AuthenticationType,
			ProxyType:         s.ProxyType,
			XCorrelationID:    correlationID,
			XSystemTenantID:   s.XSystemTenantID,
			XSystemType:       s.XSystemType,
			XSystemTenantName: s.XSystemTenantName,
		},
		Audience:         s.Audience,
		KeyStoreLocation: s.KeyStoreLocation,
	}
}

func (s *SAMLAssertionDestRequestBody) GetDestinationType() string {
	return destinationcreator.SAMLAssertionDestinationType
}

// Validate validates that the AuthTypeClientCertificate request body contains the required fields, and they are valid
func (s *ClientCertificateAuthDestRequestBody) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeClientCertificate)),
		validation.Field(&s.KeyStoreLocation, validation.Required),
	)
}

func (s *ClientCertificateAuthDestRequestBody) ToDestination() destinationcreator.Destination {
	correlationID := ""
	if s.XCorrelationID != "" {
		correlationID = s.XCorrelationID
	} else {
		correlationID = s.CorrelationIds
	}

	return &destinationcreator.ClientCertificateAuthenticationDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:              s.Name,
			Type:              s.Type,
			URL:               s.URL,
			Authentication:    s.AuthenticationType,
			ProxyType:         s.ProxyType,
			XCorrelationID:    correlationID,
			XSystemTenantID:   s.XSystemTenantID,
			XSystemType:       s.XSystemType,
			XSystemTenantName: s.XSystemTenantName,
		},
		KeyStoreLocation: s.KeyStoreLocation,
	}
}

func (s *ClientCertificateAuthDestRequestBody) GetDestinationType() string {
	return destinationcreator.ClientCertDestinationType
}

// Validate validates that the AuthTypeBasic request body contains the required fields, and they are valid
func (b *OAuth2ClientCredsDestRequestBody) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeOAuth2ClientCredentials)),
		validation.Field(&b.TokenServiceURL, validation.Required),
		validation.Field(&b.ClientID, validation.Required),
		validation.Field(&b.KeyStoreLocation, validation.When(b.ClientSecret != "", validation.Empty).Else(validation.Required)),
		validation.Field(&b.ClientSecret, validation.When(b.KeyStoreLocation != "", validation.Empty).Else(validation.Required)),
	)
}

func (b *OAuth2ClientCredsDestRequestBody) ToDestination() destinationcreator.Destination {
	correlationID := ""
	if b.XCorrelationID != "" {
		correlationID = b.XCorrelationID
	} else {
		correlationID = b.CorrelationIds
	}

	if b.KeyStoreLocation != "" {
		return &destinationcreator.OAuth2mTLSDestination{
			NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
				Name:              b.Name,
				Type:              b.Type,
				URL:               b.URL,
				Authentication:    b.AuthenticationType,
				ProxyType:         b.ProxyType,
				XCorrelationID:    correlationID,
				XSystemTenantID:   b.XSystemTenantID,
				XSystemType:       b.XSystemType,
				XSystemTenantName: b.XSystemTenantName,
			},
			TokenServiceURL:     b.TokenServiceURL,
			TokenServiceURLType: b.TokenServiceURLType,
			ClientID:            b.ClientID,
			KeyStoreLocation:    b.KeyStoreLocation,
		}
	}

	return &destinationcreator.OAuth2ClientCredentialsDestination{
		NoAuthenticationDestination: destinationcreator.NoAuthenticationDestination{
			Name:              b.Name,
			Type:              b.Type,
			URL:               b.URL,
			Authentication:    b.AuthenticationType,
			ProxyType:         b.ProxyType,
			XCorrelationID:    correlationID,
			XSystemTenantID:   b.XSystemTenantID,
			XSystemType:       b.XSystemType,
			XSystemTenantName: b.XSystemTenantName,
		},
		TokenServiceURL: b.TokenServiceURL,
		ClientID:        b.ClientID,
		ClientSecret:    b.ClientSecret,
	}
}

func (b *OAuth2ClientCredsDestRequestBody) GetDestinationType() string {
	if b.KeyStoreLocation != "" {
		return destinationcreator.OAuth2mTLSType
	} else {
		return destinationcreator.OAuth2ClientCredentialsType
	}
}

// Validate validates that the SAML assertion certificate request body contains the required fields, and they are valid
func (c *CertificateRequestBody) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.FileName, validation.Required, validation.By(func(value interface{}) error {
			split := strings.Split(c.FileName, ".")
			if len(split) != 2 {
				return fmt.Errorf("file name %q must contain a single '.'", c.FileName)
			}

			if len(split[0]) > destinationcreatorpkg.MaxDestinationNameLength {
				return fmt.Errorf("file name %q must have a length of at most %d without the file extension", c.FileName, destinationcreatorpkg.MaxDestinationNameLength)
			}

			matched := regexp.MustCompile(reqBodyNameRegex).MatchString(split[0])
			if !matched {
				return fmt.Errorf("file name %s must match the regex %s without the file extension", c.FileName, reqBodyNameRegex)
			}
			return nil
		})),
	)
}

type PostResponse struct {
	Name   string `json:"name"`
	Status int    `json:"status"`
	Cause  string `json:"cause,omitempty"`
}

func GetDestinationPrefixNameIdentifier(name string) string {
	return fmt.Sprintf("name_%s", name)
}

type DeleteStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type DeleteResponse struct {
	Count   int
	Summary []DeleteStatus
}
