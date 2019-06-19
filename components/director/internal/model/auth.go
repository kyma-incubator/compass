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
	Credential            CredentialData
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
}

type CredentialData struct {
	Basic *BasicCredentialData
	Oauth *OAuthCredentialData
}

type BasicCredentialData struct {
	Username string
	Password string
}
type OAuthCredentialData struct {
	ClientID     string
	ClientSecret string
	URL          string
}

type AuthInput struct {
	Credential            *CredentialDataInput
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
	RequestAuth           *CredentialRequestAuthInput
}

type CredentialDataInput struct {
	Basic *BasicCredentialDataInput
	Oauth *OAuthCredentialDataInput
}

type BasicCredentialDataInput struct {
	Username string
	Password string
}

type OAuthCredentialDataInput struct {
	ClientID     string
	ClientSecret string
	URL          string
}

type CredentialRequestAuthInput struct {
	Csrf *CSRFTokenCredentialRequestAuthInput
}

type CSRFTokenCredentialRequestAuthInput struct {
	TokenEndpointURL string
	Credential            *CredentialDataInput
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
}
