package clients

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/require"
)

type HydratorClient struct {
	httpClient   *http.Client
	validatorURL string
}

func NewHydratorClient(validatorURL string) *HydratorClient {
	httpClient := &http.Client{
		Transport: httputil.NewServiceAccountTokenTransport(httputil.NewHTTPTransportWrapper(&http.Transport{ // Needed because hydrators are behind PeerAuthentication
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})),
	}

	return &HydratorClient{
		httpClient:   httpClient,
		validatorURL: validatorURL,
	}
}

func (vc *HydratorClient) ExecuteHydratorRequest(t *testing.T, path string, headers map[string][]string) oathkeeper.AuthenticationSession {
	authSession := oathkeeper.AuthenticationSession{}
	marshalledSession, err := json.Marshal(authSession)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", vc.validatorURL, path), bytes.NewBuffer(marshalledSession))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	for h, vals := range headers {
		for _, v := range vals {
			req.Header.Set(h, v)
		}
	}

	response, err := vc.httpClient.Do(req)
	require.NoError(t, err)
	defer func() {
		err := response.Body.Close()
		require.NoError(t, err)
	}()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var authSessionResponse oathkeeper.AuthenticationSession
	err = json.NewDecoder(response.Body).Decode(&authSessionResponse)
	require.NoError(t, err)

	return authSessionResponse
}

func (vc *HydratorClient) ResolveCertificateData(t *testing.T, headers map[string][]string) oathkeeper.AuthenticationSession {
	return vc.ExecuteHydratorRequest(t, "/v1/certificate/data/resolve", headers)
}

func (vc *HydratorClient) ResolveToken(t *testing.T, headers map[string][]string) oathkeeper.AuthenticationSession {
	return vc.ExecuteHydratorRequest(t, "/v1/tokens/resolve", headers)
}
