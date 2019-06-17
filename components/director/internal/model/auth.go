package model

type Auth struct {
	Credential            CredentialData
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
	RequestAuth           *CredentialRequestAuth
}

type CredentialRequestAuth struct {
	Csrf *CSRFTokenCredentialRequestAuth
}

type CSRFTokenCredentialRequestAuth struct {
	TokenEndpointURL string
	Auth             *Auth
}

type CredentialData interface {
	IsCredentialData()
}

type BasicCredentialData struct {
	Username string
	Password string
}

func (BasicCredentialData) IsCredentialData() {}

type OAuthCredentialData struct {
	ClientID     string
	ClientSecret string
	URL          string
}

func (OAuthCredentialData) IsCredentialData() {}
