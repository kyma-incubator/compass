package destinationcreator

import (
	"encoding/json"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/destinationcreator"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// Config holds destination creator service API test configuration
type Config struct {
	CorrelationIDsKey string `envconfig:"APP_DESTINATION_CREATOR_CORRELATION_IDS_KEY"`
	*DestinationAPIConfig
	*CertificateAPIConfig
}

// DestinationAPIConfig holds a test configuration specific for the destination API of the destination creator service
type DestinationAPIConfig struct {
	Path                 string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_REGION_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_SUBACCOUNT_ID_PARAMETER"`
	DestinationNameParam string `envconfig:"APP_DESTINATION_CREATOR_DESTINATION_NAME_PARAMETER"`
}

// CertificateAPIConfig holds a test configuration specific for the certificate API of the destination creator service
type CertificateAPIConfig struct {
	Path                 string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_PATH"`
	RegionParam          string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_REGION_PARAMETER"`
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_SUBACCOUNT_ID_PARAMETER"`
	CertificateNameParam string `envconfig:"APP_DESTINATION_CREATOR_CERTIFICATE_NAME_PARAMETER"`
}

// BaseDestinationRequestBody contains the base fields needed in the destination request body
type BaseDestinationRequestBody struct {
	Name                 string                       `json:"name"`
	URL                  string                       `json:"url"`
	Type                 destinationcreator.Type      `json:"type"`
	ProxyType            destinationcreator.ProxyType `json:"proxyType"`
	AuthenticationType   destinationcreator.AuthType  `json:"authenticationType"`
	AdditionalProperties json.RawMessage              `json:"additionalProperties,omitempty"`
}

// DesignTimeRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeNoAuth
type DesignTimeRequestBody struct {
	BaseDestinationRequestBody
}

// BasicRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeBasic
type BasicRequestBody struct {
	BaseDestinationRequestBody
	User     string `json:"user"`
	Password string `json:"password"`
}

// SAMLAssertionRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeSAMLAssertion
type SAMLAssertionRequestBody struct {
	BaseDestinationRequestBody
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// CertificateRequestBody contains the necessary fields for the destination creator certificate request body
type CertificateRequestBody struct {
	Name string `json:"name"`
}

// CertificateResponseBody contains the response data from the destination creator certificate request
type CertificateResponseBody struct {
	FileName         string `json:"fileName"`
	CommonName       string `json:"commonName"`
	CertificateChain string `json:"certificateChain"`
}

// reqBodyNameRegex is a regex defined by the destination creator API specifying what destination names are allowed
var reqBodyNameRegex = "[a-zA-Z0-9_-]{1,64}"

// Validate validates that the AuthTypeNoAuth request body contains the required fields and they are valid
func (n *DesignTimeRequestBody) Validate() error {
	return validation.ValidateStruct(n,
		validation.Field(&n.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&n.URL, validation.Required),
		validation.Field(&n.Type, validation.In(destinationcreator.TypeHTTP, destinationcreator.TypeRFC, destinationcreator.TypeLDAP, destinationcreator.TypeMAIL)),
		validation.Field(&n.ProxyType, validation.In(destinationcreator.ProxyTypeInternet, destinationcreator.ProxyTypeOnPremise, destinationcreator.ProxyTypePrivateLink)),
		validation.Field(&n.AuthenticationType, validation.In(destinationcreator.AuthTypeNoAuth)),
	)
}

// Validate validates that the AuthTypeBasic request body contains the required fields and they are valid
func (b *BasicRequestBody) Validate(destinationCreatorCfg *Config) error {
	areAdditionalPropertiesValid := newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg)

	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(destinationcreator.TypeHTTP, destinationcreator.TypeRFC, destinationcreator.TypeLDAP, destinationcreator.TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(destinationcreator.ProxyTypeInternet, destinationcreator.ProxyTypeOnPremise, destinationcreator.ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(destinationcreator.AuthTypeBasic)),
		validation.Field(&b.User, validation.Required, validation.Length(1, 256)),
		validation.Field(&b.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the AuthTypeSAMLAssertion request body contains the required fields and they are valid
func (s *SAMLAssertionRequestBody) Validate(destinationCreatorCfg *Config) error {
	areAdditionalPropertiesValid := newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg)

	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(destinationcreator.TypeHTTP, destinationcreator.TypeRFC, destinationcreator.TypeLDAP, destinationcreator.TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(destinationcreator.ProxyTypeInternet, destinationcreator.ProxyTypeOnPremise, destinationcreator.ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(destinationcreator.AuthTypeSAMLAssertion)),
		validation.Field(&s.Audience, validation.Required),
		validation.Field(&s.KeyStoreLocation, validation.Required),
		validation.Field(&s.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the SAML assertion certificate request body contains the required fields and they are valid
func (c *CertificateRequestBody) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
	)
}

type destinationDetailsAdditionalPropertiesValidator struct {
	destinationCreatorCfg *Config
}

func newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg *Config) *destinationDetailsAdditionalPropertiesValidator {
	return &destinationDetailsAdditionalPropertiesValidator{
		destinationCreatorCfg: destinationCreatorCfg,
	}
}

// Validate is a custom method that validates the correlation IDs, as part of the arbitrary destination additional attributes, are in the expected format - string divided by comma
func (d *destinationDetailsAdditionalPropertiesValidator) Validate(value interface{}) error {
	j, ok := value.(json.RawMessage)
	if !ok {
		return errors.Errorf("Invalid type: %T, expected: %T", value, json.RawMessage{})
	}

	if valid := json.Valid(j); !valid {
		return errors.New("The additional properties json is not valid")
	}

	if d.destinationCreatorCfg == nil {
		return errors.New("The destination creator config could not be empty")
	}

	if d.destinationCreatorCfg.CorrelationIDsKey == "" {
		return errors.New("The correlation IDs key in the destination creator config could not be empty")
	}

	correlationIDsResult := gjson.Get(string(j), d.destinationCreatorCfg.CorrelationIDsKey)
	if !correlationIDsResult.Exists() {
		return errors.New("The correlationIds property part of the additional properties json is required but it does not exist.")
	}

	if correlationIDs := correlationIDsResult.String(); correlationIDs == "" {
		return errors.New("The correlationIds property part of the additional properties could not be empty")
	}

	return nil
}
