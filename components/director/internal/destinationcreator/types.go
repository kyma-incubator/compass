package destinationcreator

import (
	"encoding/json"
	"fmt"
	"regexp"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

var (
	certificateSAMLAssertionDestinationPrefix       = fmt.Sprintf("%s-", destinationcreatorpkg.AuthTypeSAMLAssertion)
	certificateSAMLBearerAssertionDestinationPrefix = fmt.Sprintf("%s-", destinationcreatorpkg.AuthTypeSAMLBearerAssertion)
	certificateClientCertificateDestinationPrefix   = fmt.Sprintf("%s-", destinationcreatorpkg.AuthTypeClientCertificate)
)

// Validator validates destination creator request body
type Validator interface {
	Validate() error
}

// Destination Creator API types

// CertificateResponse contains the response data from the destination creator certificate request
type CertificateResponse struct {
	FileName         string `json:"fileName"`
	CommonName       string `json:"commonName"`
	CertificateChain string `json:"certificateChain"`
}

// BaseDestinationRequestBody contains the base fields needed in the destination request body
type BaseDestinationRequestBody struct {
	Name                 string                          `json:"name"`
	URL                  string                          `json:"url"`
	Type                 destinationcreatorpkg.Type      `json:"type"`
	ProxyType            destinationcreatorpkg.ProxyType `json:"proxyType"`
	AuthenticationType   destinationcreatorpkg.AuthType  `json:"authenticationType"`
	AdditionalProperties json.RawMessage                 `json:"additionalProperties,omitempty"`
}

// NoAuthDestinationRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeNoAuth
type NoAuthDestinationRequestBody struct {
	BaseDestinationRequestBody
}

// BasicAuthDestinationRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeBasic
type BasicAuthDestinationRequestBody struct {
	BaseDestinationRequestBody
	User     string `json:"user"`
	Password string `json:"password"`
}

// SAMLAssertionDestinationRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeSAMLAssertion
type SAMLAssertionDestinationRequestBody struct {
	BaseDestinationRequestBody
	Audience         string `json:"audience"`
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// ClientCertAuthDestinationRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeClientCertificate
type ClientCertAuthDestinationRequestBody struct {
	BaseDestinationRequestBody
	KeyStoreLocation string `json:"keyStoreLocation"`
}

// CertificateRequestBody contains the necessary fields for the destination creator certificate request body
type CertificateRequestBody struct {
	Name       string `json:"name"`
	SelfSigned bool   `json:"selfSigned"`
}

// reqBodyNameRegex is a regex defined by the destination creator API specifying what destination names are allowed
var reqBodyNameRegex = "[a-zA-Z0-9_-]{1,64}"

// Validate validates that the AuthTypeNoAuth request body contains the required fields and they are valid
func (n *NoAuthDestinationRequestBody) Validate() error {
	return validation.ValidateStruct(n,
		validation.Field(&n.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&n.URL, validation.Required),
		validation.Field(&n.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&n.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&n.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeNoAuth)),
	)
}

// Validate validates that the AuthTypeBasic request body contains the required fields and they are valid
func (b *BasicAuthDestinationRequestBody) Validate(destinationCreatorCfg *Config) error {
	areAdditionalPropertiesValid := newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg)

	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeBasic)),
		validation.Field(&b.User, validation.Required, validation.Length(1, 256)),
		validation.Field(&b.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the AuthTypeSAMLAssertion request body contains the required fields and they are valid
func (s *SAMLAssertionDestinationRequestBody) Validate(destinationCreatorCfg *Config) error {
	areAdditionalPropertiesValid := newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg)

	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeSAMLAssertion)),
		validation.Field(&s.Audience, validation.Required),
		validation.Field(&s.KeyStoreLocation, validation.Required),
		validation.Field(&s.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the AuthTypeClientCertificate request body contains the required fields and they are valid
func (s *ClientCertAuthDestinationRequestBody) Validate(destinationCreatorCfg *Config) error {
	areAdditionalPropertiesValid := newDestinationDetailsAdditionalPropertiesValidator(destinationCreatorCfg)

	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(destinationcreatorpkg.TypeHTTP, destinationcreatorpkg.TypeRFC, destinationcreatorpkg.TypeLDAP, destinationcreatorpkg.TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(destinationcreatorpkg.ProxyTypeInternet, destinationcreatorpkg.ProxyTypeOnPremise, destinationcreatorpkg.ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(destinationcreatorpkg.AuthTypeClientCertificate)),
		validation.Field(&s.KeyStoreLocation, validation.Required),
		validation.Field(&s.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the SAML assertion certificate request body contains the required fields and they are valid
func (c *CertificateRequestBody) Validate() error {
	return validation.ValidateStruct(c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, destinationcreatorpkg.MaxDestinationNameLength), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
	)
}

// Validate validates that the SAML assertion certificate response body contains the required fields and they are valid
func (cr *CertificateResponse) Validate() error {
	return validation.ValidateStruct(cr,
		validation.Field(&cr.FileName, validation.Required),
		validation.Field(&cr.CommonName, validation.Required),
		validation.Field(&cr.CertificateChain, validation.Required),
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
