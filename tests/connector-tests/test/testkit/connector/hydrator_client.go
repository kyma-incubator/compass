package connector

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/require"
)

type HydratorClient struct {
	httpClient   *http.Client
	validatorURL string
}

func NewHydratorClient(validatorURL string) *HydratorClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return &HydratorClient{
		httpClient:   httpClient,
		validatorURL: validatorURL,
	}
}

func (vc *HydratorClient) ResolveToken(t *testing.T, headers map[string][]string) oathkeeper.AuthenticationSession {
	return vc.executeHydratorRequest(t, "/v1/tokens/resolve", headers)
}

func (vc *HydratorClient) ResolveCertificateData(t *testing.T, headers map[string][]string) oathkeeper.AuthenticationSession {
	return vc.executeHydratorRequest(t, "/v1/certificate/data/resolve", headers)
}

func (vc *HydratorClient) executeHydratorRequest(t *testing.T, path string, headers map[string][]string) oathkeeper.AuthenticationSession {
	authSession := oathkeeper.AuthenticationSession{}
	body, err := json.Marshal(authSession)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", vc.validatorURL, path), bytes.NewBuffer(body))
	require.NoError(t, err)

	for h, vals := range headers {
		for _, v := range vals {
			req.Header.Set(h, v)
		}
	}

	fmt.Println(req.Header)

	response, err := vc.httpClient.Do(req)
	require.NoError(t, err)
	defer func() {
		err := response.Body.Close()
		if err != nil {
			logrus.Warnf("Failed to close body: %s", err.Error())
		}
	}()
	require.Equal(t, http.StatusOK, response.StatusCode)

	var authSessionResponse oathkeeper.AuthenticationSession
	err = json.NewDecoder(response.Body).Decode(&authSessionResponse)
	require.NoError(t, err)

	return authSessionResponse
}
