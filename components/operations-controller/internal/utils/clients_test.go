package utils_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/utils"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/utils/automock"
	sb_http "github.com/kyma-incubator/compass/components/system-broker/pkg/http"

	"github.com/stretchr/testify/require"
)

const testServerResponse = "Hello, client"

func TestPrepareMTLSClient(t *testing.T) {
	var (
		cfg         = &sb_http.Config{Timeout: time.Second, SkipSSLValidation: true}
		testCert    = &tls.Certificate{}
		mockedCache = &automock.CertificateCache{}
		ts          = httptest.NewUnstartedServer(testServerHandlerFunc(t))
	)

	mockedCache.On("Get").Return([]*tls.Certificate{testCert}, nil).Once()
	ts.TLS = &tls.Config{
		ClientAuth: tls.RequestClientCert,
	}

	ts.StartTLS()
	defer ts.Close()

	mtlsClient := utils.PrepareMtlsClient(cfg, mockedCache)

	resp, err := mtlsClient.Get(ts.URL)
	require.NoError(t, err)

	respPayload, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(respPayload), testServerResponse)

	mockedCache.AssertExpectations(t)
}

func TestPrepareHttpClient(t *testing.T) {
	var (
		cfg              = &sb_http.Config{Timeout: time.Second, SkipSSLValidation: true}
		ts               = httptest.NewUnstartedServer(testServerHandlerFunc(t))
		basicCredentials = &auth.BasicCredentials{
			Username: "testUser",
			Password: "testPass",
		}
		authContext = auth.SaveToContext(context.Background(), basicCredentials)
	)
	ts.Start()
	defer ts.Close()

	client := utils.PrepareHttpClient(cfg)

	req, err := http.NewRequest(http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	req = req.WithContext(authContext)

	resp, err := client.Do(req)
	require.NoError(t, err)

	respPayload, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(respPayload), testServerResponse)
}

func testServerHandlerFunc(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, testServerResponse)
		require.NoError(t, err)
	}
}
