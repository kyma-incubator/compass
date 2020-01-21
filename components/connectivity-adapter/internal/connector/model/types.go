package model

import (
	"fmt"
)

const (
	TokenFormat                       = "?token=%s"
	CertsEndpoint                     = "/v1/applications/certificates"
	ManagementInfoEndpoint            = "/v1/applications/management/info"
	ApplicationRegistryEndpointFormat = "/%s/v1/metadata"
	EventsEndpointFormat              = "/%s/v1/events"
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
	ClientIdentity  interface{} `json:"clientIdentity"`
	URLs            MgmtURLs    `json:"urls"`
	CertificateInfo CertInfo    `json:"certificate"`
}

type RuntimeURLs struct {
	//EventsInfoURL string `json:"eventsInfoUrl"` // TODO: Where is it used?
	EventsURL   string `json:"eventsUrl"`
	MetadataURL string `json:"metadataUrl"`
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

func NewCSRInfoResponse(certInfo CertInfo, clientIdFromToken, token, baseURL string) CSRInfoResponse {
	return CSRInfoResponse{
		CsrURL:          makeCSRURLs(token, baseURL),
		API:             makeApiURLs(clientIdFromToken, baseURL),
		CertificateInfo: certInfo,
	}
}

func makeCSRURLs(newToken, baseURL string) string {
	csrURL := baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func makeApiURLs(clientIdFromToken, baseURL string) Api {
	return Api{
		CertificatesURL: baseURL + CertsEndpoint,
		InfoURL:         baseURL + ManagementInfoEndpoint,
		RuntimeURLs: &RuntimeURLs{
			MetadataURL: baseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, clientIdFromToken),
			EventsURL:   baseURL + fmt.Sprintf(EventsEndpointFormat, clientIdFromToken),
		},
	}
}
