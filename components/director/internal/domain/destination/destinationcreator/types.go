package destinationcreator

const (
	TypeHTTP Type = "HTTP"
	TypeRFC  Type = "RFC"
	TypeLDAP Type = "LDAP"
	TypeMAIL Type = "MAIL"

	AuthTypeNoAuth     AuthType = "NoAuthentication"
	AuthTypeBasic      AuthType = "BasicAuthentication"
	AuthTypeSAMLBearer AuthType = "OAuth2SAMLBearerAssertion"

	ProxyTypeInternet    ProxyType = "Internet"
	ProxyTypeOnPremise   ProxyType = "OnPremise"
	ProxyTypePrivateLink ProxyType = "PrivateLink"
)

type Type string
type AuthType string
type ProxyType string

// Destination Creator API types

// RequestBody // todo::: add godoc
type RequestBody struct {
	Name               string    `json:"name"`
	Url                string    `json:"url"`
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
