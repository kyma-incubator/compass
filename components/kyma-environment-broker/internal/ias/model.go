package ias

type Company struct {
	ServiceProviders  []ServiceProvider  `json:"service_providers"`
	IdentityProviders []IdentityProvider `json:"identity_providers"`
}

type ServiceProvider struct {
	DisplayName         string               `json:"display_name"`
	ID                  string               `json:"id"`
	AssertionAttributes []AssertionAttribute `json:"assertion_attributes"`
	DefaultAttributes   []DefaultAttribute   `json:"default_attributes"`
	Organization        string               `json:"organization"`
	SsoType             string               `json:"ssoType"`
	RedirectURIs        []string             `json:"redirect_uris"`
	NameIDAttribute     string               `json:"name_id_attribute"`
	RBAConfig           RBAConfig            `json:"rba_config"`
	AuthenticatingIdp   AuthenticatingIdp    `json:"authenticatingIdp"`
	Secret              []SPSecret           `json:"clientSecrets"`
	ACSEndpoints        []ACSEndpoint        `json:"acs_endpoints"`
}

type AuthenticatingIdp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type SPSecret struct {
	SecretID    string   `json:"clientSecretId"`
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
}

type AssertionAttribute struct {
	AssertionAttribute string `json:"assertionAttribute"`
	UserAttribute      string `json:"userAttribute"`
}

type DefaultAttribute struct {
	AssertionAttribute string `json:"assertionAttribute"`
	Value              string `json:"value"`
}

type PostAssertionAttributes struct {
	AssertionAttributes []AssertionAttribute `json:"assertion_attributes"`
}

type IdentityProvider struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type OIDCType struct {
	SsoType             string              `json:"ssoType"`
	ServiceProviderName string              `json:"sp_name"`
	OpenIDConnectConfig OpenIDConnectConfig `json:"openId_connect_configuration"`
}

type SAMLType struct {
	ServiceProviderName string        `json:"sp_name"`
	ACSEndpoints        []ACSEndpoint `json:"acs_endpoints"`
}

type ACSEndpoint struct {
	Location  string `json:"location"`
	IsDefault bool   `json:"isDefault,omitempty"`
	Index     int32  `json:"index"`
}

type OpenIDConnectConfig struct {
	RedirectURIs           []string `json:"redirect_uris"`
	PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris,omitempty"`
}

type SubjectNameIdentifier struct {
	NameIDAttribute string `json:"name_id_attribute"`
}

type SecretConfiguration struct {
	Organization        string              `json:"organization"`
	ID                  string              `json:"id"`
	RestAPIClientSecret RestAPIClientSecret `json:"rest_api_client_secret"`
}

type DefaultAuthIDPConfig struct {
	Organization   string `json:"organization"`
	ID             string `json:"id"`
	DefaultAuthIDP string `json:"default_auth_idp"`
}

type RestAPIClientSecret struct {
	Description string   `json:"description"`
	Scopes      []string `json:"scopes"`
}

type ServiceProviderSecret struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type AuthenticationAndAccess struct {
	ServiceProviderAccess ServiceProviderAccess `json:"service_provider"`
}

type ServiceProviderAccess struct {
	RBAConfig RBAConfig `json:"rba_config"`
}

type RBAConfig struct {
	RBARules      []RBARules `json:"rba_rules"`
	DefaultAction string     `json:"default_action"`
}

type RBARules struct {
	Action    string `json:"action"`
	Group     string `json:"group"`
	GroupType string `json:"group_type"`
}
