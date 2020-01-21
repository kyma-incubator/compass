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

func NewCSRInfoResponse(certInfo CertInfo, clientIdFromToken, token, connectivityAdapterBaseURL, eventServiceBaseURL string) CSRInfoResponse {
	return CSRInfoResponse{
		CsrURL:          makeCSRURLs(token, connectivityAdapterBaseURL),
		API:             makeApiURLs(clientIdFromToken, connectivityAdapterBaseURL, eventServiceBaseURL),
		CertificateInfo: certInfo,
	}
}

func makeCSRURLs(newToken, connectivityAdapterBaseURL string) string {
	csrURL := connectivityAdapterBaseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func makeApiURLs(clientIdFromToken, connectivityAdapterBaseURL string, eventServiceBaseURL string) Api {
	return Api{
		CertificatesURL: connectivityAdapterBaseURL + CertsEndpoint,
		InfoURL:         connectivityAdapterBaseURL + ManagementInfoEndpoint,
		RuntimeURLs: &RuntimeURLs{
			MetadataURL: connectivityAdapterBaseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, clientIdFromToken),
			EventsURL:   eventServiceBaseURL + fmt.Sprintf(EventsEndpointFormat, clientIdFromToken),
		},
	}
}
