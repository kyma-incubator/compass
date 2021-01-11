package apitests

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pivotal-cf/brokerapi/v7/domain"

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
		Name: "test-runtime1",
	}
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, context.TODO(), dexGraphQLClient, testCtx.Tenant, runtimeInput)
	defer director.UnregisterRuntimeWithinTenant(t, context.TODO(), dexGraphQLClient, testCtx.Tenant, runtime.ID)

	runtimeToken := director.GenerateOneTimeTokenForRuntime(t, context.TODO(), dexGraphQLClient, testCtx.Tenant, runtime.ID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	clientKey, err := connectorTestkit.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	certResult, configuration := connector.GenerateRuntimeCertificate(t, oneTimeToken, testCtx.ConnectorTokenSecuredClient, clientKey)
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

		client := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		_, err = client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls: certificate required")
	})

	t.Run("Gets unauthorized when calling with revoked cert", func(t *testing.T) {
		connectorClient := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, clientKey, certChain...)
		ok, err := connectorClient.RevokeCertificate()
		require.NoError(t, err)
		require.Equal(t, ok, true)

		details, _ := json.Marshal(domain.BindDetails{
			ServiceID: "serviceID",
			PlanID:    "planID",
		})

		req, err := http.NewRequest(http.MethodPut, testCtx.SystemBrokerURL+"/v2/service_instances/2be0980c-92d2-460f-9568-ffcbb98155c7/service_bindings/043ccdb4-0ebc-475b-849f-6afec54fdd95?accepts_incomplete=true", bytes.NewBuffer(details))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Broker-API-Version", "2.15")

		resp, err := securedClient.Do(req)
		logrus.Infof("Response is: %+v", resp)
		require.Error(t, err)
		require.Contains(t, err.Error(), "abc")
		require.Nil(t, resp)
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
