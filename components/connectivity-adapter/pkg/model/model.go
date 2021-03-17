package model

import (
	"encoding/json"
	"strconv"
	"strings"
)

const (
	TokenFormat                       = "?token=%s"
	CertsEndpoint                     = "/v1/applications/certificates"
	ManagementInfoEndpoint            = "/v1/applications/management/info"
	ApplicationRegistryEndpointFormat = "/%s/v1/metadata/services"
	EventsInfoEndpoint                = "/subscribed"
	RenewCertURLFormat                = "%s/v1/applications/certificates/renewals"
	RevocationCertURLFormat           = "%s/v1/applications/certificates/revocations"
)

type SpecResponse string

func (sp *SpecResponse) UnmarshalJSON(bytes []byte) error {
	if sp == nil {
		sp = new(SpecResponse)
	}

	if IsXML(string(bytes)) {
		*sp = SpecResponse(bytes)
		return nil
	}

	jsonRawMessage := json.RawMessage{}
	if err := json.Unmarshal(bytes, &jsonRawMessage); err != nil {
		return err
	}

	*sp = SpecResponse(jsonRawMessage)

	return nil
}

func (sp *SpecResponse) MarshalJSON() ([]byte, error) {
	if sp == nil {
		return nil, nil
	}
	if IsXML(string(*sp)) {
		unquote, err := strconv.Unquote(string(*sp))
		if err == nil {
			*sp = SpecResponse(unquote)
		}

		return []byte(strconv.Quote(string(*sp))), nil
	}
	return json.Marshal(json.RawMessage(*sp))
}

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
	Spec                           *SpecResponse        `json:"spec,omitempty"`
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
	Spec *SpecResponse `json:"spec" valid:"required~spec cannot be empty"`
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

type CertRequest struct {
	CSR string `json:"csr"`
}

type CertResponse struct {
	CRTChain  string `json:"crt"`
	ClientCRT string `json:"clientCrt"`
	CaCRT     string `json:"caCrt"`
}

type CSRInfoResponse struct {
	CsrURL          string   `json:"csrUrl"`
	API             Api      `json:"api"`
	CertificateInfo CertInfo `json:"certificate"`
}

type MgmtInfoReponse struct {
	ClientIdentity  ClientIdentity `json:"clientIdentity"`
	URLs            MgmtURLs       `json:"urls"`
	CertificateInfo CertInfo       `json:"certificate"`
}

type RuntimeURLs struct {
	EventsInfoURL string `json:"eventsInfoUrl"`
	EventsURL     string `json:"eventsUrl"`
	MetadataURL   string `json:"metadataUrl"`
}

type MgmtURLs struct {
	*RuntimeURLs
	RenewCertURL  string `json:"renewCertUrl"`
	RevokeCertURL string `json:"revokeCertUrl"`
}

type Api struct {
	*RuntimeURLs
	InfoURL         string `json:"infoUrl"`
	CertificatesURL string `json:"certificatesUrl"`
}

type CertInfo struct {
	Subject      string `json:"subject"`
	Extensions   string `json:"extensions"`
	KeyAlgorithm string `json:"key-algorithm"`
}

type ClientIdentity struct {
	Application string `json:"application"`
	Group       string `json:"group,omitempty"`
	Tenant      string `json:"tenant,omitempty"`
}

type TokenResponse struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

func IsXML(content string) bool {
	const snippetLength = 512

	if unquoted, err := strconv.Unquote(content); err == nil {
		content = unquoted
	}

	var snippet string
	length := len(content)
	if length < snippetLength {
		snippet = content
	} else {
		snippet = content[:snippetLength]
	}

	openingIndex := strings.Index(snippet, "<")
	closingIndex := strings.Index(snippet, ">")

	return openingIndex == 0 && openingIndex < closingIndex
}
