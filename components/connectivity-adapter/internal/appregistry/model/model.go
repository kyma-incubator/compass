package model

import "encoding/json"

type Service struct {
	ID          string             `json:"id"`
	Provider    string             `json:"provider"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Identifier  string             `json:"identifier,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
}

type ServiceDetails struct {
	Provider         string             `json:"provider" valid:"required~Provider field cannot be empty."`
	Name             string             `json:"name" valid:"required~Name field cannot be empty."`
	Description      string             `json:"description" valid:"required~Description field cannot be empty."`
	ShortDescription string             `json:"shortDescription,omitempty"`
	Identifier       string             `json:"identifier,omitempty"`
	Labels           *map[string]string `json:"labels,omitempty"`
	Api              *API               `json:"api,omitempty"`
	Events           *Events            `json:"events,omitempty"`
	Documentation    *Documentation     `json:"documentation,omitempty"`
}

type CreateServiceResponse struct {
	ID string `json:"id"`
}

type API struct {
	TargetUrl                      string               `json:"targetUrl" valid:"url,required~targetUrl field cannot be empty."`
	Credentials                    *CredentialsWithCSRF `json:"credentials,omitempty"`
	Spec                           json.RawMessage      `json:"spec,omitempty"`
	SpecificationUrl               string               `json:"specificationUrl,omitempty"`
	ApiType                        string               `json:"apiType,omitempty"`
	RequestParameters              *RequestParameters   `json:"requestParameters,omitempty"`
	SpecificationCredentials       *Credentials         `json:"specificationCredentials,omitempty"`
	SpecificationRequestParameters *RequestParameters   `json:"specificationRequestParameters,omitempty"`
	Headers                        *map[string][]string `json:"headers,omitempty"`
	QueryParameters                *map[string][]string `json:"queryParameters,omitempty"`
}

type RequestParameters struct {
	Headers         *map[string][]string `json:"headers,omitempty"`
	QueryParameters *map[string][]string `json:"queryParameters,omitempty"`
}

type Credentials struct {
	Oauth *Oauth     `json:"oauth,omitempty"`
	Basic *BasicAuth `json:"basic,omitempty"`
}

type CredentialsWithCSRF struct {
	OauthWithCSRF          *OauthWithCSRF          `json:"oauth,omitempty"`
	BasicWithCSRF          *BasicAuthWithCSRF      `json:"basic,omitempty"`
	CertificateGenWithCSRF *CertificateGenWithCSRF `json:"certificateGen,omitempty"`
}

type CSRFInfo struct {
	TokenEndpointURL string `json:"tokenEndpointURL" valid:"url,required~tokenEndpointURL field cannot be empty"`
}

type Oauth struct {
	URL               string             `json:"url" valid:"url,required~oauth url field cannot be empty"`
	ClientID          string             `json:"clientId" valid:"required~oauth clientId field cannot be empty"`
	ClientSecret      string             `json:"clientSecret" valid:"required~oauth clientSecret cannot be empty"`
	RequestParameters *RequestParameters `json:"requestParameters,omitempty"`
}

type OauthWithCSRF struct {
	Oauth
	CSRFInfo *CSRFInfo `json:"csrfInfo,omitempty"`
}

type BasicAuth struct {
	Username string `json:"username" valid:"required~basic auth username field cannot be empty"`
	Password string `json:"password" valid:"required~basic auth password field cannot be empty"`
}

type BasicAuthWithCSRF struct {
	BasicAuth
	CSRFInfo *CSRFInfo `json:"csrfInfo,omitempty"`
}

type CertificateGen struct {
	CommonName  string `json:"commonName"`
	Certificate string `json:"certificate"`
}

type CertificateGenWithCSRF struct {
	CertificateGen
	CSRFInfo *CSRFInfo `json:"csrfInfo,omitempty"`
}

type Events struct {
	Spec json.RawMessage `json:"spec" valid:"required~spec cannot be empty"`
}

type Documentation struct {
	DisplayName string       `json:"displayName" valid:"required~displayName field cannot be empty in documentation"`
	Description string       `json:"description" valid:"required~description field cannot be empty in documentation"`
	Type        string       `json:"type" valid:"required~type field cannot be empty in documentation"`
	Tags        []string     `json:"tags,omitempty"`
	Docs        []DocsObject `json:"docs,omitempty"`
}

type DocsObject struct {
	Title  string `json:"title"`
	Type   string `json:"type"`
	Source string `json:"source"`
}
