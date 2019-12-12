package oauth

const (
	contentTypeHeader = "Content-Type"

	grantTypeFieldName   = "grant_type"
	credentialsGrantType = "client_credentials"

	scopeFieldName = "scope"
	scopes         = "application:read application:write runtime:read runtime:write label_definition:read label_definition:write health_checks:read"

	clientIDKey     = "ClientID"
	clientSecretKey = "ClientSecret"
)

type Token struct {
	AccessToken string `json:"access_token"`
	Expiration  int    `json:"expires_in"`
}

type credentials struct {
	clientID     string
	clientSecret string
}
