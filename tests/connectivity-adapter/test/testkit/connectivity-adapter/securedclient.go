package connectivity_adapter

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"

	"github.com/stretchr/testify/require"
)

type SecuredClient interface {
	GetMgmInfo(t *testing.T, url string) (*ManagementInfoResponse, *Error)
	RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error)
	RevokeCertificate(t *testing.T, url string) *Error

	ListServices(t *testing.T, url string) ([]model.Service, *Error)
	CreateService(t *testing.T, url string, service model.ServiceDetails) (*CreateServiceResponse, *Error)
	GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *Error)
	UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *Error)
	DeleteService(t *testing.T, url string, id string) *Error
}

type securedConnectorClient struct {
	httpClient *http.Client
	tenant     string
}

func NewSecuredClient(skipVerify bool, key *rsa.PrivateKey, certs []byte, tenant string) SecuredClient {
	client := newTLSClientWithCert(skipVerify, key, certs)

	return &securedConnectorClient{
		httpClient: client,
		tenant:     tenant,
	}
}

func newTLSClientWithCert(skipVerify bool, key *rsa.PrivateKey, certificate ...[]byte) *http.Client {
	tlsCert := tls.Certificate{
		Certificate: certificate,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: skipVerify,
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	return &http.Client{
		Transport: transport,
	}
}

func (cc securedConnectorClient) GetMgmInfo(t *testing.T, url string) (*ManagementInfoResponse, *Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var mgmInfoResponse ManagementInfoResponse
	errorResp := cc.secureConnectorRequest(t, request, &mgmInfoResponse, http.StatusOK)

	return &mgmInfoResponse, errorResp
}

func (cc securedConnectorClient) RenewCertificate(t *testing.T, url string, csr string) (*CrtResponse, *Error) {
	body, err := json.Marshal(CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	var certificateResponse CrtResponse
	errorResp := cc.secureConnectorRequest(t, request, &certificateResponse, http.StatusCreated)

	return &certificateResponse, errorResp
}

func (cc securedConnectorClient) RevokeCertificate(t *testing.T, url string) *Error {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	request.Close = true

	return cc.secureConnectorRequest(t, request, nil, http.StatusCreated)
}

func (cc securedConnectorClient) ListServices(t *testing.T, url string) ([]model.Service, *Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var services []model.Service
	errorResp := cc.secureConnectorRequest(t, request, &services, http.StatusOK)

	return services, errorResp
}

func (cc securedConnectorClient) CreateService(t *testing.T, url string, service model.ServiceDetails) (*CreateServiceResponse, *Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, url, http.MethodPost, service)

	var createServiceResponse CreateServiceResponse
	errorResp := cc.secureConnectorRequest(t, request, &createServiceResponse, http.StatusOK)

	return &createServiceResponse, errorResp
}

func (cc securedConnectorClient) GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *Error) {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodGet)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc securedConnectorClient) UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodPut, service)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc securedConnectorClient) DeleteService(t *testing.T, url string, id string) *Error {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodDelete)

	return cc.secureConnectorRequest(t, request, nil, http.StatusOK)
}

func (cc securedConnectorClient) secureConnectorRequest(t *testing.T, request *http.Request, data interface{}, expectedStatus int) *Error {
	response, err := cc.httpClient.Do(request)
	require.NoError(t, err)
	defer func() {
		err := response.Body.Close()
		require.NoError(t, err)
	}()

	if response.StatusCode != expectedStatus {
		return parseErrorResponse(t, response)
	}

	if data != nil {
		err = json.NewDecoder(response.Body).Decode(&data)
		require.NoError(t, err)
	}

	return nil
}
