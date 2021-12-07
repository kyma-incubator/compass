package oauth

import "github.com/form3tech-oss/jwt-go"

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Claims struct {
	Tenant string `json:"x-zid,omitempty"`
	Client string `json:"client_id,omitempty"`
	Scopes string `json:"scopes,omitempty"`
	jwt.StandardClaims
}
