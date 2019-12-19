package oauth

import "time"

const (
	contentTypeHeader = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	grantTypeFieldName   = "grant_type"
	credentialsGrantType = "client_credentials"

	scopeFieldName = "scope"
	scopes         = "application:read application:write runtime:read runtime:write label_definition:read label_definition:write health_checks:read"

	clientIDKey     = "ClientID"
	clientSecretKey = "ClientSecret"
)

type Token struct {
	AccessToken string `json:"access_token"`
	Expiration  int64  `json:"expires_in"`
}

type credentials struct {
	clientID     string
	clientSecret string
}

func (token Token) EmptyOrExpired() bool {
	if token.AccessToken == "" {
		return true
	}

	expiration := time.Unix(token.Expiration, 0)
	return time.Now().After(expiration)
}
