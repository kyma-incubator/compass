package model

import (
	"fmt"
)

const (
	TokenFormat                       = "?token=%s"
	CertsEndpoint                     = "/v1/applications/certificates"
	ManagementInfoEndpoint            = "/v1/applications/management/info"
	ApplicationRegistryEndpointFormat = "/%s/v1/metadata/services"
	EventsEndpointFormat              = "/%s/v1/events"
	EventsInfoEndpointFormat          = "/%s/v1/events/subscribed"
	RenewCertURLFormat                = "%s/v1/applications/certificates/renewals"
	RevocationCertURLFormat           = "%s/v1/applications/certificates/revocations"
	SigningRequestInfoEndpoint        = "%s/v1/applications/signingRequests/info"
	TokenURLFormat                    = "%s?token=%s"
)

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

func MakeCSRURL(newToken, connectivityAdapterBaseURL string) string {
	csrURL := connectivityAdapterBaseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func MakeApiURLs(application, connectivityAdapterBaseURL string, connectivityAdapterMTLSBaseURL string, eventServiceBaseURL string) Api {
	return Api{
		CertificatesURL: connectivityAdapterBaseURL + CertsEndpoint,
		InfoURL:         connectivityAdapterMTLSBaseURL + ManagementInfoEndpoint,
		RuntimeURLs:     makeRuntimeURLs(application, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
	}
}

func makeRuntimeURLs(application, connectivityAdapterBaseURL string, eventServiceBaseURL string) *RuntimeURLs {
	return &RuntimeURLs{
		MetadataURL:   connectivityAdapterBaseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, application),
		EventsURL:     eventServiceBaseURL + fmt.Sprintf(EventsEndpointFormat, application),
		EventsInfoURL: eventServiceBaseURL + fmt.Sprintf(EventsInfoEndpointFormat, application),
	}
}

func MakeClientIdentity(application, tenant, group string) ClientIdentity {
	return ClientIdentity{
		Application: application,
		Tenant:      tenant,
		Group:       group,
	}
}

func MakeManagementURLs(application, connectivityAdapterMTLSBaseURL string, eventServiceBaseURL string) MgmtURLs {
	return MgmtURLs{
		RuntimeURLs:   makeRuntimeURLs(application, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
		RenewCertURL:  fmt.Sprintf(RenewCertURLFormat, connectivityAdapterMTLSBaseURL),
		RevokeCertURL: fmt.Sprintf(RevocationCertURLFormat, connectivityAdapterMTLSBaseURL),
	}
}

func MakeTokenResponse(connectivityAdapterBaseURL string, token string) TokenResponse {
	csrInfoUrl := fmt.Sprintf(SigningRequestInfoEndpoint, connectivityAdapterBaseURL)

	return TokenResponse{
		Token: token,
		URL:   fmt.Sprintf(TokenURLFormat, csrInfoUrl, token),
	}
}
