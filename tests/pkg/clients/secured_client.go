package clients

import (
	"bytes"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"testing"

	model2 "github.com/kyma-incubator/compass/tests/pkg/model"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"

	"github.com/stretchr/testify/require"
)

type SecuredClient interface {
	GetMgmInfo(t *testing.T, url string) (*model2.ManagementInfoResponse, *model2.Error)
	RenewCertificate(t *testing.T, url string, csr string) (*model2.CrtResponse, *model2.Error)
	RevokeCertificate(t *testing.T, url string) *model2.Error

	ListServices(t *testing.T, url string) ([]model.Service, *model2.Error)
	CreateService(t *testing.T, url string, service model.ServiceDetails) (*model2.CreateServiceResponse, *model2.Error)
	GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *model2.Error)
	UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *model2.Error)
	DeleteService(t *testing.T, url string, id string) *model2.Error
}

type securedConnectorClient struct {
	httpClient *http.Client
	tenant     string
}

func NewSecuredClient(skipVerify bool, key *rsa.PrivateKey, certs []byte, tenant string) (SecuredClient, error) {
	client, err := newTLSClientWithCert(skipVerify, key, certs)
	if err != nil {
		return nil, err
	}

	return &securedConnectorClient{
		httpClient: client,
		tenant:     tenant,
	}, err
}

func newTLSClientWithCert(skipVerify bool, key *rsa.PrivateKey, certs []byte) (*http.Client, error) {
	pemEncodedKey := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	tlsCert, err := tls.X509KeyPair(certs, pemEncodedKey)
	if err != nil {
		return nil, err
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
	}, nil
}

func (cc *securedConnectorClient) GetMgmInfo(t *testing.T, url string) (*model2.ManagementInfoResponse, *model2.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var mgmInfoResponse model2.ManagementInfoResponse
	errorResp := cc.secureConnectorRequest(t, request, &mgmInfoResponse, http.StatusOK)

	return &mgmInfoResponse, errorResp
}

func (cc *securedConnectorClient) RenewCertificate(t *testing.T, url string, csr string) (*model2.CrtResponse, *model2.Error) {
	body, err := json.Marshal(model2.CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)

	var certificateResponse model2.CrtResponse
	errorResp := cc.secureConnectorRequest(t, request, &certificateResponse, http.StatusCreated)

	return &certificateResponse, errorResp
}

func (cc *securedConnectorClient) RevokeCertificate(t *testing.T, url string) *model2.Error {
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	request.Close = true

	return cc.secureConnectorRequest(t, request, nil, http.StatusCreated)
}

func (cc *securedConnectorClient) ListServices(t *testing.T, url string) ([]model.Service, *model2.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	var services []model.Service
	errorResp := cc.secureConnectorRequest(t, request, &services, http.StatusOK)

	return services, errorResp
}

func (cc *securedConnectorClient) CreateService(t *testing.T, url string, service model.ServiceDetails) (*model2.CreateServiceResponse, *model2.Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, url, http.MethodPost, service)

	var createServiceResponse model2.CreateServiceResponse
	errorResp := cc.secureConnectorRequest(t, request, &createServiceResponse, http.StatusOK)

	return &createServiceResponse, errorResp
}

func (cc *securedConnectorClient) GetService(t *testing.T, url string, id string) (*model.ServiceDetails, *model2.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodGet)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc *securedConnectorClient) UpdateService(t *testing.T, url string, id string, service model.ServiceDetails) (*model.ServiceDetails, *model2.Error) {
	request := requestWithTenantHeadersAndBody(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodPut, service)

	var serviceDetails model.ServiceDetails
	errorResp := cc.secureConnectorRequest(t, request, &serviceDetails, http.StatusOK)

	return &serviceDetails, errorResp
}

func (cc *securedConnectorClient) DeleteService(t *testing.T, url string, id string) *model2.Error {
	request := requestWithTenantHeaders(t, cc.tenant, fmt.Sprintf("%s/%s", url, id), http.MethodDelete)

	return cc.secureConnectorRequest(t, request, nil, http.StatusNoContent)
}

func (cc *securedConnectorClient) secureConnectorRequest(t *testing.T, request *http.Request, data interface{}, expectedStatus int) *model2.Error {
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
