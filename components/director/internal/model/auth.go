package model

// Auth missing godoc
type Auth struct {
	Credential            CredentialData
	AccessStrategy        *string
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
	RequestAuth           *CredentialRequestAuth
	OneTimeToken          *OneTimeToken
	CertCommonName        string
}

// CredentialRequestAuth missing godoc
type CredentialRequestAuth struct {
	Csrf *CSRFTokenCredentialRequestAuth
}

// CSRFTokenCredentialRequestAuth missing godoc
type CSRFTokenCredentialRequestAuth struct {
	TokenEndpointURL      string
	Credential            CredentialData
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
}

// CredentialData missing godoc
type CredentialData struct {
	Basic            *BasicCredentialData
	Oauth            *OAuthCredentialData
	CertificateOAuth *CertificateOAuthCredentialData
}

// BasicCredentialData missing godoc
type BasicCredentialData struct {
	Username string
	Password string
}

// OAuthCredentialData missing godoc
type OAuthCredentialData struct {
	ClientID     string
	ClientSecret string
	URL          string
}

// CertificateOAuthCredentialData represents a structure for mTLS OAuth credentials
type CertificateOAuthCredentialData struct {
	ClientID    string
	Certificate string
	URL         string
}

// AuthInput missing godoc
type AuthInput struct {
	Credential            *CredentialDataInput
	AccessStrategy        *string
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
	RequestAuth           *CredentialRequestAuthInput
	OneTimeToken          *OneTimeToken
}

// ToAuth missing godoc
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
		AccessStrategy:        i.AccessStrategy,
		AdditionalQueryParams: i.AdditionalQueryParams,
		AdditionalHeaders:     i.AdditionalHeaders,
		RequestAuth:           requestAuth,
		OneTimeToken:          i.OneTimeToken,
	}
}

// CredentialDataInput missing godoc
type CredentialDataInput struct {
	Basic            *BasicCredentialDataInput
	Oauth            *OAuthCredentialDataInput
	CertificateOAuth *CertificateOAuthCredentialDataInput
}

// ToCredentialData missing godoc
func (i *CredentialDataInput) ToCredentialData() *CredentialData {
	if i == nil {
		return nil
	}

	var basic *BasicCredentialData
	var oauth *OAuthCredentialData
	var certOAuth *CertificateOAuthCredentialData

	if i.Basic != nil {
		basic = i.Basic.ToBasicCredentialData()
	}

	if i.Oauth != nil {
		oauth = i.Oauth.ToOAuthCredentialData()
	}

	if i.CertificateOAuth != nil {
		certOAuth = i.CertificateOAuth.ToCertificateOAuthCredentialData()
	}

	return &CredentialData{
		Basic:            basic,
		Oauth:            oauth,
		CertificateOAuth: certOAuth,
	}
}

// BasicCredentialDataInput missing godoc
type BasicCredentialDataInput struct {
	Username string
	Password string
}

// ToBasicCredentialData missing godoc
func (i *BasicCredentialDataInput) ToBasicCredentialData() *BasicCredentialData {
	if i == nil {
		return nil
	}

	return &BasicCredentialData{
		Username: i.Username,
		Password: i.Password,
	}
}

// OAuthCredentialDataInput missing godoc
type OAuthCredentialDataInput struct {
	ClientID     string
	ClientSecret string
	URL          string
}

// ToOAuthCredentialData missing godoc
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

// CertificateOAuthCredentialDataInput represents an input structure for mTLS OAuth credentials
type CertificateOAuthCredentialDataInput struct {
	ClientID    string
	Certificate string
	URL         string
}

// ToCertificateOAuthCredentialData converts a CertificateOAuthCredentialDataInput into CertificateOAuthCredentialData
func (i *CertificateOAuthCredentialDataInput) ToCertificateOAuthCredentialData() *CertificateOAuthCredentialData {
	if i == nil {
		return nil
	}

	return &CertificateOAuthCredentialData{
		ClientID:    i.ClientID,
		Certificate: i.Certificate,
		URL:         i.URL,
	}
}

// CredentialRequestAuthInput missing godoc
type CredentialRequestAuthInput struct {
	Csrf *CSRFTokenCredentialRequestAuthInput
}

// ToCredentialRequestAuth missing godoc
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

// CSRFTokenCredentialRequestAuthInput missing godoc
type CSRFTokenCredentialRequestAuthInput struct {
	TokenEndpointURL      string
	Credential            *CredentialDataInput
	AdditionalHeaders     map[string][]string
	AdditionalQueryParams map[string][]string
}

// ToCSRFTokenCredentialRequestAuth missing godoc
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
