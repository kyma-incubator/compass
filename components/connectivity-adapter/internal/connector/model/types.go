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
	RenewCertURLFormat                = "%s/certificates/renewals"
	RevocationCertURLFormat           = "%s/certificates/revocations"
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

type ClientIdentity struct {
	Application string `json:"application"`
	Group       string `json:"group,omitempty"`
	Tenant      string `json:"tenant,omitempty"`
}

//func NewCSRInfoResponse(certInfo CertInfo, application, token, connectivityAdapterBaseURL, eventServiceBaseURL string) CSRInfoResponse {
//	return CSRInfoResponse{
//		CsrURL:          makeCSRURLs(token, connectivityAdapterBaseURL),
//		API:             makeApiURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
//		CertificateInfo: certInfo,
//	}
//}

func MakeCSRURL(newToken, connectivityAdapterBaseURL string) string {
	csrURL := connectivityAdapterBaseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func MakeApiURLs(application, connectivityAdapterBaseURL string, eventServiceBaseURL string) Api {
	return Api{
		CertificatesURL: connectivityAdapterBaseURL + CertsEndpoint,
		InfoURL:         connectivityAdapterBaseURL + ManagementInfoEndpoint,
		RuntimeURLs:     makeRuntimeURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
	}
}

func makeRuntimeURLs(application, connectivityAdapterBaseURL string, eventServiceBaseURL string) *RuntimeURLs {
	return &RuntimeURLs{
		MetadataURL: connectivityAdapterBaseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, application),
		EventsURL:   eventServiceBaseURL + fmt.Sprintf(EventsEndpointFormat, application),
	}
}

func MakeClientIdentity(application, tenant, group string) ClientIdentity {
	return ClientIdentity{
		Application: application,
		Tenant:      tenant,
		Group:       group,
	}
}

func MakeManagementURLs(application, connectivityAdapterBaseURL string, eventServiceBaseURL string) MgmtURLs {
	return MgmtURLs{
		RuntimeURLs:   makeRuntimeURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
		RenewCertURL:  fmt.Sprintf(RenewCertURLFormat, connectivityAdapterBaseURL),
		RevokeCertURL: fmt.Sprintf(RevocationCertURLFormat, connectivityAdapterBaseURL),
	}
}

//func NewManagementInfoResponse(application, connectivityAdapterBaseURL string, eventServiceBaseURL string) MgmtInfoReponse{
//	return MgmtInfoReponse {
//		ClientIdentity: ClientIdentity{
//
//		},
//		URLs : MgmtURLs{
//			RuntimeURLs: makeRuntimeURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
//			RenewCertURL: fmt.Sprintf(RenewCertURLFormat, connectivityAdapterBaseURL),
//			RevokeCertURL: fmt.Sprintf(RevocationCertURLFormat, connectivityAdapterBaseURL),
//		},
//		CertificateInfo: CertInfo{
//
//		},
//	}
//}
