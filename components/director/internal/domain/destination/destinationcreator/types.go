package destinationcreator

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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

// Type represents the HTTP destination type
type Type string

// AuthType represents the destination authentication
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string

// Destination Creator API types

// BaseDestinationRequestBody todo::: go doc
type BaseDestinationRequestBody struct {
	Name                 string          `json:"name"`
	URL                  string          `json:"url"`
	Type                 Type            `json:"type"`
	ProxyType            ProxyType       `json:"proxyType"`
	AuthenticationType   AuthType        `json:"authenticationType"`
	AdditionalAttributes json.RawMessage `json:"additionalAttributes,omitempty"`
}

// NoAuthRequestBody // todo::: add godoc
type NoAuthRequestBody struct {
	BaseDestinationRequestBody
}

// BasicRequestBody // todo::: add godoc
type BasicRequestBody struct {
	BaseDestinationRequestBody
	User     string `json:"user"`
	Password string `json:"password"`
}

func (n *NoAuthRequestBody) Validate() error {
	return validation.ValidateStruct(n,
		validation.Field(&n.Name, validation.Required, validation.Length(1, 200)),
		validation.Field(&n.URL, validation.Required),
		validation.Field(&n.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
		validation.Field(&n.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
		validation.Field(&n.AuthenticationType, validation.In(AuthTypeNoAuth)),
	)
}

func (b *BasicRequestBody) Validate() error {
	return validation.ValidateStruct(b,
		validation.Field(&b.Name, validation.Required, validation.Length(1, 200)),
		validation.Field(&b.URL, validation.Required),
		validation.Field(&b.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
		validation.Field(&b.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
		validation.Field(&b.AuthenticationType, validation.In(AuthTypeBasic)),
	)
}
