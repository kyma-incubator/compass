package clients

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"

	"github.com/stretchr/testify/require"
)

type SecuredClient interface {
	GetMgmInfo(t *testing.T, url string) (*pkg.ManagementInfoResponse, *pkg.Error)
	RenewCertificate(t *testing.T, url string, csr string) (*pkg.CrtResponse, *pkg.Error)
	RevokeCertificate(t *testing.T, url string) *pkg.Error

	ListServices(t *testing.T, url string) ([]model.Service, *pkg.Error)
	CreateService(t *testing.T, url string, service model.ServiceDetails) (*pkg.CreateServiceResponse, *pkg.Error)
	GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *pkg.Error)
	UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *pkg.Error)
	DeleteService(t *testing.T, url string, id string) *pkg.Error
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

func (cc securedConnectorClient) GetMgmInfo(t *testing.T, url string) (*pkg.ManagementInfoResponse, *pkg.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var mgmInfoResponse pkg.ManagementInfoResponse
	errorResp := cc.secureConnectorRequest(t, request, &mgmInfoResponse, http.StatusOK)

	return &mgmInfoResponse, errorResp
}

func (cc securedConnectorClient) RenewCertificate(t *testing.T, url string, csr string) (*pkg.CrtResponse, *pkg.Error) {
	body, err := json.Marshal(pkg.CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	var certificateResponse pkg.CrtResponse
	errorResp := cc.secureConnectorRequest(t, request, &certificateResponse, http.StatusCreated)

	return &certificateResponse, errorResp
}

func (cc securedConnectorClient) RevokeCertificate(t *testing.T, url string) *pkg.Error {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	request.Close = true

	return cc.secureConnectorRequest(t, request, nil, http.StatusCreated)
}

func (cc securedConnectorClient) ListServices(t *testing.T, url string) ([]model.Service, *pkg.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var services []model.Service
	errorResp := cc.secureConnectorRequest(t, request, &services, http.StatusOK)

	return services, errorResp
}

func (cc securedConnectorClient) CreateService(t *testing.T, url string, service model.ServiceDetails) (*pkg.CreateServiceResponse, *pkg.Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, url, http.MethodPost, service)

	var createServiceResponse pkg.CreateServiceResponse
	errorResp := cc.secureConnectorRequest(t, request, &createServiceResponse, http.StatusOK)

	return &createServiceResponse, errorResp
}

func (cc securedConnectorClient) GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *pkg.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodGet)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc securedConnectorClient) UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *pkg.Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodPut, service)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc securedConnectorClient) DeleteService(t *testing.T, url string, id string) *pkg.Error {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodDelete)

	return cc.secureConnectorRequest(t, request, nil, http.StatusNoContent)
}

func (cc securedConnectorClient) secureConnectorRequest(t *testing.T, request *http.Request, data interface{}, expectedStatus int) *pkg.Error {
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
