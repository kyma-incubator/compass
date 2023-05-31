package destinationcreator

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

// Type represents the HTTP destination type
type Type string

// AuthType represents the destination authentication
type AuthType string

// ProxyType represents the destination proxy type
type ProxyType string

// Destination Creator API types

// RequestBody // todo::: add godoc
type RequestBody struct {
	Name               string    `json:"name"`
	URL                string    `json:"url"`
	Type               Type      `json:"type"`
	ProxyType          ProxyType `json:"proxyType"`
	AuthenticationType AuthType  `json:"authenticationType"`
	User               string    `json:"user"`
	Password           string    `json:"password"`
}

// ErrorResponse // todo::: add godoc
type ErrorResponse struct {
	Error Error `json:"error"`
}

// Error // todo::: add godoc
type Error struct {
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}
