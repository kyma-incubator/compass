package apitests

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"

	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

	"github.com/sirupsen/logrus"

	connectorTestkit "github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	"github.com/stretchr/testify/require"
)

// Todo: add tests for revoked and invalid certs
func TestTokens(t *testing.T) {
	token, err := idtokenprovider.GetDexToken()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	dexGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(token, testCtx.DirectorURL)
	runtimeInput := &graphql.RuntimeInput{
		Name: "test-runtime",
	}
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, context.TODO(), dexGraphQLClient, testCtx.Tenant, runtimeInput)

	runtimeToken := director.GenerateOneTimeTokenForRuntime(t, context.TODO(), dexGraphQLClient, testCtx.Tenant, runtime.ID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	clientKey, err := connectorTestkit.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	certResult, _ := connector.GenerateRuntimeCertificate(t, oneTimeToken, testCtx.ConnectorTokenSecuredClient, clientKey)
	certChain := connectorTestkit.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := createCertClient(clientKey, certChain...)

	t.Run("Successfully calls mtls broker endpoint with certificate secured client", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, testCtx.SystemBrokerURL+"/v2/catalog", nil)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Broker-API-Version", "2.15")

		resp, err := securedClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, resp.StatusCode, http.StatusOK)
		bytes, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Contains(t, string(bytes), "services")
	})

	t.Run("Gets unauthorized when calling mtls broker endpoint with default http client", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, testCtx.SystemBrokerURL+"/v2/catalog", nil)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Broker-API-Version", "2.15")

		_, err = http.DefaultClient.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "connection reset by peer")
	})
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
