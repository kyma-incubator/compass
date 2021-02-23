package tests

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/system-broker/pkg"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	connectorTestkit "github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	director "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testRuntimeName = "test-runtime"
	testAppName     = "test-application"
	testBundleName  = "test-bundle"
	testApiDefName  = "test-api-def"
	testTargetURL   = "http://example.com"
)

var (
	runtimeInput = &graphql.RuntimeInput{
		Name: testRuntimeName,
	}
	applicationInput = graphql.ApplicationRegisterInput{
		Name: testAppName,
	}
	fri = graphql.FetchRequestInput{
		URL: testTargetURL,
	}
	apiDefInput = graphql.APIDefinitionInput{
		Name:      testApiDefName,
		TargetURL: testTargetURL,
		Spec: &graphql.APISpecInput{
			Type:         graphql.APISpecTypeOpenAPI,
			Format:       graphql.SpecFormatYaml,
			FetchRequest: &fri,
		},
	}
)

func TestSystemBrokerAuthentication(t *testing.T) {
	testCtx, err := pkg.NewTestContext(config)
	require.NoError(t, err)

	logrus.Infof("registering runtime with name: %s, within tenant: %s", runtimeInput.Name, testCtx.Tenant)
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, testCtx.Context, testCtx.DexGraphqlClient, testCtx.Tenant, runtimeInput)
	defer director.UnregisterRuntimeWithinTenant(t, testCtx.Context, testCtx.DexGraphqlClient, testCtx.Tenant, runtime.ID)

	securedClient, configuration, certChain := getSecuredClientByContext(t, testCtx, runtime.ID)

	t.Run("Should successfully call catalog endpoint with certificate secured client", func(t *testing.T) {
		req := createCatalogRequest(t, testCtx)

		resp, err := securedClient.Do(req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, resp.StatusCode, http.StatusOK)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Contains(t, string(body), "services")
	})

	t.Run("Should fail calling catalog endpoint without certificate", func(t *testing.T) {
		req := createCatalogRequest(t, testCtx)

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

		req := createCatalogRequest(t, testCtx)

		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "unauthorized: insufficient scopes")
	})

	t.Run("Should fail calling catalog endpoint with invalid certificate", func(t *testing.T) {
		req := createCatalogRequest(t, testCtx)

		fakedClientKey, err := connectorTestkit.GenerateKey()
		require.NoError(t, err)
		client := createCertClient(fakedClientKey, certChain...)

		_, err = client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls: error decrypting message")
	})
}

func TestCallingORDServiceWithCert(t *testing.T) {
	testContext, err := pkg.NewTestContext(config)
	require.NoError(t, err)

	logrus.Infof("registering runtime with name: %s, within tenant: %s", runtimeInput.Name, testContext.Tenant)
	runtime := director.RegisterRuntimeFromInputWithinTenant(t, testContext.Context, testContext.DexGraphqlClient, testContext.Tenant, runtimeInput)
	defer director.UnregisterRuntimeWithinTenant(t, testContext.Context, testContext.DexGraphqlClient, testContext.Tenant, runtime.ID)

	app, err := director.RegisterApplicationWithinTenant(t, testContext.Context, testContext.DexGraphqlClient, testContext.Tenant, applicationInput)
	require.NoError(t, err)
	defer director.UnregisterApplication(t, testContext.Context, testContext.DexGraphqlClient, testContext.Tenant, app.ID)

	bundle := director.CreateBundle(t, testContext.Context, testContext.DexGraphqlClient, app.ID, testContext.Tenant, testBundleName)
	api := director.AddAPIToBundleWithInput(t, testContext.Context, testContext.DexGraphqlClient, bundle.ID, testContext.Tenant, apiDefInput)

	securedClient, configuration, certChain := getSecuredClientByContext(t, testContext, runtime.ID)

	t.Run("Should succeed calling ORD service with the same cert used for calling catalog", func(t *testing.T) {
		req := createCatalogRequest(t, testContext)
		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "services")

		reqORD := createORDServiceRequest(t, testContext, api.ID, api.Spec.ID)
		respORD, err := securedClient.Do(reqORD)

		require.NoError(t, err)
		require.NotNil(t, respORD)
		require.Equal(t, http.StatusOK, respORD.StatusCode)
	})

	t.Run("Should fail calling ORD service with revoked cert", func(t *testing.T) {

		logrus.Infof("revoking cert for runtime with id: %s", runtime.ID)
		connectorClient := connector.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, testContext.ClientKey, certChain...)
		ok, err := connectorClient.RevokeCertificate()
		require.NoError(t, err)
		require.Equal(t, ok, true)

		req := createORDServiceRequest(t, testContext, api.ID, api.Spec.ID)

		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode) //oathkeeper cannot be configured to return 401 the way we use it

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "invalid tenantID")
	})

}

func createCatalogRequest(t *testing.T, ctx *pkg.TestContext) *http.Request {
	req, err := http.NewRequest(http.MethodGet, ctx.SystemBrokerURL+"/v2/catalog", nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Broker-API-Version", "2.15")
	return req
}

func createORDServiceRequest(t *testing.T, ctx *pkg.TestContext, apiId, specId string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, ctx.ORDServiceURL+fmt.Sprintf("/api/%s/specification/%s", apiId, specId), nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	return req
}

func getSecuredClientByContext(t *testing.T, ctx *pkg.TestContext, runtimeID string) (*http.Client, externalschema.Configuration, []*x509.Certificate) {
	logrus.Infof("generating one-time token for runtime with id: %s", runtimeID)
	runtimeToken := director.GenerateOneTimeTokenForRuntime(t, ctx.Context, ctx.DexGraphqlClient, ctx.Tenant, runtimeID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	logrus.Infof("generation certificate for runtime with id: %s", runtimeID)
	certResult, configuration := connector.GenerateRuntimeCertificate(t, oneTimeToken, ctx.ConnectorTokenSecuredClient, ctx.ClientKey)
	certChain := connectorTestkit.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := createCertClient(ctx.ClientKey, certChain...)

	return securedClient, configuration, certChain
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
