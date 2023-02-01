package clients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/model"

	"github.com/stretchr/testify/require"
)

const (
	ApplicationHeader = "Application"
	GroupHeader       = "Group"
	TenantHeader      = "Tenant"
	Tenant            = "testkit-tenant"
	Extensions        = ""
	KeyAlgorithm      = "rsa2048"
)

type ConnectorClient interface {
	CreateToken(t *testing.T) model.TokenResponse
	GetInfo(t *testing.T, url string) (*model.InfoResponse, *model.Error)
	CreateCertChain(t *testing.T, csr, url string) (*model.CrtResponse, *model.Error)
}

type connectorClient struct {
	httpClient     *http.Client
	directorClient Client
	appID          string
	tenant         string
}

func NewConnectorClient(directorClient Client, appID, tenant string, skipVerify bool) ConnectorClient {
	client := NewHttpClient(skipVerify)

	return &connectorClient{
		httpClient:     client,
		directorClient: directorClient,
		appID:          appID,
		tenant:         tenant,
	}
}

func NewHttpClient(skipVerify bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
	}
	client := &http.Client{Transport: tr}
	return client
}

func (cc *connectorClient) CreateToken(t *testing.T) model.TokenResponse {
	url, token, err := cc.directorClient.GetOneTimeTokenUrl(cc.appID)
	require.NoError(t, err)
	tokenResponse := model.TokenResponse{
		URL:   url,
		Token: token,
	}

	return tokenResponse
}

func (cc *connectorClient) GetInfo(t *testing.T, url string) (*model.InfoResponse, *model.Error) {
	request := requestWithTenantHeaders(t, cc.tenant, url, http.MethodGet)

	response, err := cc.httpClient.Do(request)
	defer func() {
		err := response.Body.Close()
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	if response.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(t, response)
	}

	require.Equal(t, http.StatusOK, response.StatusCode)

	infoResponse := &model.InfoResponse{}

	err = json.NewDecoder(response.Body).Decode(&infoResponse)
	require.NoError(t, err)

	return infoResponse, nil
}

func (cc *connectorClient) CreateCertChain(t *testing.T, csr, url string) (*model.CrtResponse, *model.Error) {
	body, err := json.Marshal(model.CsrRequest{Csr: csr})
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	require.NoError(t, err)
	request.Close = true
	request.Header.Add("Content-Type", "application/json")

	response, err := cc.httpClient.Do(request)
	defer func() {
		err := response.Body.Close()
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	if response.StatusCode != http.StatusCreated {
		return nil, parseErrorResponse(t, response)
	}

	require.Equal(t, http.StatusCreated, response.StatusCode)

	crtResponse := &model.CrtResponse{}

	err = json.NewDecoder(response.Body).Decode(&crtResponse)
	require.NoError(t, err)

	return crtResponse, nil
}

func requestWithTenantHeaders(t *testing.T, tenant, url, method string) *http.Request {
	return requestWithTenantHeadersAndBody(t, tenant, url, method, nil)
}

func requestWithTenantHeadersAndBody(t *testing.T, tenant, url, method string, rawBody interface{}) *http.Request {
	jsonBody, err := json.Marshal(rawBody)
	require.NoError(t, err)

	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
	require.NoError(t, err)

	request.Header.Set("Tenant", tenant)
	request.Close = true

	return request
}

func parseErrorResponse(t *testing.T, response *http.Response) *model.Error {
	logResponse(t, response)
	errorResponse := model.ErrorResponse{}
	err := json.NewDecoder(response.Body).Decode(&errorResponse)
	require.NoError(t, err)

	return &model.Error{response.StatusCode, errorResponse}
}

func logResponse(t *testing.T, resp *http.Response) {
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		t.Logf("failed to dump response, %s", err)
	}

	reqDump, err := httputil.DumpRequest(resp.Request, true)
	if err != nil {
		t.Logf("failed to dump request, %s", err)
	}

	if err == nil {
		t.Logf("\n--------------------------------\n%s\n--------------------------------\n%s\n--------------------------------", reqDump, respDump)
	}
}
