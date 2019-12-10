package oauth

const (
	ContentTypeHeader = "Content-Type"

	GrantTypeFieldName   = "grant_type"
	CredentialsGrantType = "client_credentials"

	ScopeFieldName = "scope"
	Scopes         = "application:read application:write runtime:read runtime:write label_definition:read label_definition:write health_checks:read"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	Expiration  int    `json:"expires_in"`
}

type Credentials struct {
	ClientID     string
	ClientSecret string
}
