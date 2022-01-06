package oauth

import "github.com/form3tech-oss/jwt-go"

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Claims struct {
	ClientId string `json:"client_id,omitempty"`
	Scopes   string `json:"scopes,omitempty"`
	Tenant   string `json:"x-zid,omitempty"`
	jwt.StandardClaims
}
