package tests

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	connectorTestkit "github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestSystemBrokerAuthentication(t *testing.T) {
	// setup
	runtimeInput := &graphql.RuntimeInput{
		Name: "test-runtime",
	}

	logrus.Infof("registering runtime with name: %s, within tenant: %s", runtimeInput.Name, testCtx.Tenant)
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, testCtx.Context, testCtx.DexGraphqlClient, testCtx.Tenant, runtimeInput)
	defer director.UnregisterRuntimeWithinTenant(t, testCtx.Context, testCtx.DexGraphqlClient, testCtx.Tenant, runtime.ID)

	logrus.Infof("generating one-time token for runtime with id: %s", runtime.ID)
	runtimeToken := director.GenerateOneTimeTokenForRuntime(t, testCtx.Context, testCtx.DexGraphqlClient, testCtx.Tenant, runtime.ID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	logrus.Infof("generation certificate for runtime with id: %s", runtime.ID)
	certResult, configuration := connector.GenerateRuntimeCertificate(t, oneTimeToken, testCtx.ConnectorTokenSecuredClient, testCtx.ClientKey)
	certChain := connectorTestkit.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := createCertClient(testCtx.ClientKey, certChain...)

	t.Run("Should successfully call catalog endpoint with certificate secured client", func(t *testing.T) {
		req := createCatalogRequest(t)

		resp, err := securedClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, resp.StatusCode, http.StatusOK)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Contains(t, string(body), "services")
	})

	t.Run("Should fail calling catalog endpoint without certificate", func(t *testing.T) {
		req := createCatalogRequest(t)

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		_, err := client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls: certificate required")
	})

	t.Run("Should fail calling catalog endpoint with revoked cert", func(t *testing.T) {
		logrus.Infof("revoking cert for runtime with id: %s", runtime.ID)
		connectorClient := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, testCtx.ClientKey, certChain...)
		ok, err := connectorClient.RevokeCertificate()
		require.NoError(t, err)
		require.Equal(t, ok, true)

		req := createCatalogRequest(t)

		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "unauthorized: insufficient scopes")
	})

	t.Run("Should fail calling catalog endpoint with invalid certificate", func(t *testing.T) {
		req := createCatalogRequest(t)

		fakedClientKey, err := connectorTestkit.GenerateKey()
		require.NoError(t, err)
		client := createCertClient(fakedClientKey, certChain...)

		_, err = client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls: error decrypting message")
	})
}

func createCatalogRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, testCtx.SystemBrokerURL+"/v2/catalog", nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Broker-API-Version", "2.15")
	return req
}

func createCertClient(key *rsa.PrivateKey, certificates ...*x509.Certificate) *http.Client {
	rawCerts := make([][]byte, len(certificates))
	for i, c := range certificates {
		rawCerts[i] = c.Raw
	}

	tlsCert := tls.Certificate{
		Certificate: rawCerts,
		PrivateKey:  key,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{tlsCert},
		ClientAuth:         tls.RequireAndVerifyClientCert,
		InsecureSkipVerify: true,
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}
