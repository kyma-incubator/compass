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
	TokenEndpointURL      string
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

func (i *AuthInput) ToAuth() *Auth {
	if i == nil {
		return nil
	}

	var credential CredentialData
	if i.Credential != nil {
		credential = *i.Credential.ToCredentialData()
	}

	var requestAuth *CredentialRequestAuth
	if i.RequestAuth != nil {
		requestAuth = i.RequestAuth.ToCredentialRequestAuth()
	}

	return &Auth{
		Credential:            credential,
		AdditionalQueryParams: i.AdditionalQueryParams,
		AdditionalHeaders:     i.AdditionalHeaders,
		RequestAuth:           requestAuth,
	}
}

type CredentialDataInput struct {
	Basic *BasicCredentialDataInput
	Oauth *OAuthCredentialDataInput
}

func (i *CredentialDataInput) ToCredentialData() *CredentialData {
	if i == nil {
		return nil
	}

	var basic *BasicCredentialData
	var oauth *OAuthCredentialData

	if i.Basic != nil {
		basic = i.Basic.ToBasicCredentialData()
	}

	if i.Oauth != nil {
		oauth = i.Oauth.ToOAuthCredentialData()
	}

	return &CredentialData{
		Basic: basic,
		Oauth: oauth,
	}
}

type BasicCredentialDataInput struct {
	Username string
	Password string
}

func (i *BasicCredentialDataInput) ToBasicCredentialData() *BasicCredentialData {
	if i == nil {
		return nil
	}

	return &BasicCredentialData{
		Username: i.Username,
		Password: i.Password,
	}
}

type OAuthCredentialDataInput struct {
	ClientID     string
	ClientSecret string
	URL          string
}

func (i *OAuthCredentialDataInput) ToOAuthCredentialData() *OAuthCredentialData {
	if i == nil {
		return nil
	}

	return &OAuthCredentialData{
		ClientID:     i.ClientID,
		ClientSecret: i.ClientSecret,
		URL:          i.URL,
	}
}

type CredentialRequestAuthInput struct {
	Csrf *CSRFTokenCredentialRequestAuthInput
}

func (i *CredentialRequestAuthInput) ToCredentialRequestAuth() *CredentialRequestAuth {
	if i == nil {
		return nil
	}

	var csrf *CSRFTokenCredentialRequestAuth
	if i.Csrf != nil {
		csrf = i.Csrf.ToCSRFTokenCredentialRequestAuth()
	}

	return &CredentialRequestAuth{
		Csrf: csrf,
	}
}

type CSRFTokenCredentialRequestAuthInput struct {
	TokenEndpointURL      string
	Credential            *CredentialDataInput
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
}

func (i *CSRFTokenCredentialRequestAuthInput) ToCSRFTokenCredentialRequestAuth() *CSRFTokenCredentialRequestAuth {
	if i == nil {
		return nil
	}

	var credential CredentialData
	if i.Credential != nil {
		credential = *i.Credential.ToCredentialData()
	}

	return &CSRFTokenCredentialRequestAuth{
		Credential:            credential,
		AdditionalHeaders:     i.AdditionalHeaders,
		AdditionalQueryParams: i.AdditionalQueryParams,
		TokenEndpointURL:      i.TokenEndpointURL,
	}
}
