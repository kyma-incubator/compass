package destinationcreator

import (
	"encoding/json"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	// TypeHTTP represents the HTTP destination type
	TypeHTTP Type = "HTTP"
	// TypeRFC represents the RFC destination type
	TypeRFC Type = "RFC"
	// TypeLDAP represents the LDAP destination type
	TypeLDAP Type = "LDAP"
	// TypeMAIL represents the MAIL destination type
	TypeMAIL Type = "MAIL"

	// AuthTypeNoAuth represents the NoAuth destination authentication
	AuthTypeNoAuth AuthType = "NoAuthentication"
	// AuthTypeBasic represents the Basic destination authentication
	AuthTypeBasic AuthType = "BasicAuthentication"
	// AuthTypeSAMLAssertion represents the SAMLAssertion destination authentication
	AuthTypeSAMLAssertion AuthType = "SAMLAssertion"
	// AuthTypeSAMLBearer represents the SAMLBearer destination authentication
	AuthTypeSAMLBearer AuthType = "OAuth2SAMLBearerAssertion"

	// ProxyTypeInternet represents the Internet proxy type
	ProxyTypeInternet ProxyType = "Internet"
	// ProxyTypeOnPremise represents the OnPremise proxy type
	ProxyTypeOnPremise ProxyType = "OnPremise"
	// ProxyTypePrivateLink represents the PrivateLink proxy type
	ProxyTypePrivateLink ProxyType = "PrivateLink"
)

// Validator validates destination creator request body
type Validator interface {
	Validate() error
}

// Type represents the destination type
type Type string

// AuthType represents the destination authentication type
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string

// Destination Creator API types

// CertificateResponse contains the response data from the destination creator certificate request
type CertificateResponse struct {
	FileName         string `json:"fileName"`
	CommonName       string `json:"commonName"`
	CertificateChain string `json:"certificateChain"`
}

// BaseDestinationRequestBody contains the base fields needed in the destination request body
type BaseDestinationRequestBody struct {
	Name                 string          `json:"name"`
	URL                  string          `json:"url"`
	Type                 Type            `json:"type"`
	ProxyType            ProxyType       `json:"proxyType"`
	AuthenticationType   AuthType        `json:"authenticationType"`
	AdditionalProperties json.RawMessage `json:"additionalProperties,omitempty"`
}

// NoAuthRequestBody contains the necessary fields for the destination request body with authentication type AuthTypeNoAuth
type NoAuthRequestBody struct {
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

// reqBodyNameRegex is a regex defined by the destination creator API specifying what destination names are allowed
var reqBodyNameRegex = "[a-zA-Z0-9_-]{1,64}"

// Validate validates that the AuthTypeNoAuth request body contains the required fields and they are valid
func (n *NoAuthRequestBody) Validate() error {
	return validation.ValidateStruct(n,
		validation.Field(&n.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&n.URL, validation.Required),
		validation.Field(&n.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
		validation.Field(&n.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
		validation.Field(&n.AuthenticationType, validation.In(AuthTypeNoAuth)),
	)
}

// Validate validates that the AuthTypeBasic request body contains the required fields and they are valid
func (b *BasicRequestBody) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(AuthTypeBasic)),
		validation.Field(&b.User, validation.Required, validation.Length(1, 256)),
		validation.Field(&b.AdditionalProperties, areAdditionalPropertiesValid),
	)
}

// Validate validates that the AuthTypeSAMLAssertion request body contains the required fields and they are valid
func (s *SAMLAssertionRequestBody) Validate() error {
	return validation.ValidateStruct(s,
		validation.Field(&s.Name, validation.Required, validation.Length(1, 64), validation.Match(regexp.MustCompile(reqBodyNameRegex))),
		validation.Field(&s.URL, validation.Required),
		validation.Field(&s.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
		validation.Field(&s.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
		validation.Field(&s.AuthenticationType, validation.In(AuthTypeSAMLAssertion)),
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

// Validate validates that the SAML assertion certificate response body contains the required fields and they are valid
func (cr *CertificateResponse) Validate() error {
	return validation.ValidateStruct(cr,
		validation.Field(&cr.FileName, validation.Required),
		validation.Field(&cr.CommonName, validation.Required),
		validation.Field(&cr.CertificateChain, validation.Required),
	)
}

type destinationDetailsAdditionalPropertiesValidator struct{}

var areAdditionalPropertiesValid = &destinationDetailsAdditionalPropertiesValidator{}

// Validate ads
func (d *destinationDetailsAdditionalPropertiesValidator) Validate(value interface{}) error {
	j, ok := value.(json.RawMessage)
	if !ok {
		return errors.Errorf("Invalid type: %T, expected: %T", value, json.RawMessage{})
	}

	if valid := json.Valid(j); !valid {
		return errors.New("The additional properties json is not valid")
	}

	correlationIDsResult := gjson.Get(string(j), "correlationIds") // todo::: extract as environment variable
	if !correlationIDsResult.Exists() {
		return errors.New("The correlationIds property part of the additional properties json is required but it does not exist.")
	}

	if correlationIDs := correlationIDsResult.String(); correlationIDs == "" {
		return errors.New("The correlationIds property part of the additional properties could not be empty")
	}

	return nil
}
