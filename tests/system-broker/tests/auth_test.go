package tests

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/kyma-incubator/compass/tests/pkg/authentication"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx/broker"

	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	runtimeInput = &graphql.RuntimeRegisterInput{
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
	runtimeInput = &graphql.RuntimeRegisterInput{
		Name: testRuntimeName,
	}

	logrus.Infof("registering runtime with name: %s, within tenant: %s", runtimeInput.Name, testCtx.Tenant)
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, runtimeInput)
	defer fixtures.CleanupRuntime(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

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
		connectorClient := clients.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, testCtx.ClientKey, certChain...)
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

		fakedClientKey, err := certs.GenerateKey()
		require.NoError(t, err)
		client := authentication.CreateCertClient(fakedClientKey, certChain...)

		_, err = client.Do(req)
		require.Error(t, err)
		require.Contains(t, err.Error(), "tls: error decrypting message")
	})
}

func TestCallingORDServiceWithCert(t *testing.T) {
	logrus.Infof("registering runtime with name: %s, within tenant: %s", runtimeInput.Name, testCtx.Tenant)
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, runtimeInput)
	defer fixtures.CleanupRuntime(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	app, err := fixtures.RegisterApplicationFromInput(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, applicationInput)
	defer fixtures.CleanupApplication(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, &app)
	require.NoError(t, err)

	bundle := fixtures.CreateBundle(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, app.ID, testBundleName)
	defer fixtures.DeleteBundle(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, bundle.ID)

	api := fixtures.AddAPIToBundleWithInput(t, testCtx.Context, testCtx.CertSecuredGraphQLClient, testCtx.Tenant, bundle.ID, apiDefInput)

	securedClient, configuration, certChain := getSecuredClientByContext(t, testCtx, runtime.ID)

	t.Run("Should succeed calling ORD service with the same cert used for calling catalog", func(t *testing.T) {
		req := createCatalogRequest(t, testCtx)
		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, resp.StatusCode, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "services")

		reqORD := createORDServiceRequest(t, testCtx, api.ID, api.Spec.ID)
		respORD, err := securedClient.Do(reqORD)

		require.NoError(t, err)
		require.NotNil(t, respORD)
		require.Equal(t, http.StatusOK, respORD.StatusCode)
	})

	t.Run("Should fail calling ORD service with revoked cert", func(t *testing.T) {

		logrus.Infof("revoking cert for runtime with id: %s", runtime.ID)
		connectorClient := clients.NewCertificateSecuredConnectorClient(*configuration.ManagementPlaneInfo.CertificateSecuredConnectorURL, testCtx.ClientKey, certChain...)
		ok, err := connectorClient.RevokeCertificate()
		require.NoError(t, err)
		require.Equal(t, ok, true)

		req := createORDServiceRequest(t, testCtx, api.ID, api.Spec.ID)

		resp, err := securedClient.Do(req)

		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode) //oathkeeper cannot be configured to return 401 the way we use it

		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "invalid tenantID")
	})

}

func getSecuredClientByContext(t *testing.T, ctx *broker.SystemBrokerTestContext, runtimeID string) (*http.Client, externalschema.Configuration, []*x509.Certificate) {
	logrus.Infof("generating one-time token for runtime with id: %s", runtimeID)
	runtimeToken := fixtures.RequestOneTimeTokenForRuntime(t, ctx.Context, ctx.CertSecuredGraphQLClient, ctx.Tenant, runtimeID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	logrus.Infof("generation certificate for runtime with id: %s", runtimeID)
	certResult, configuration := clients.GenerateRuntimeCertificate(t, oneTimeToken, ctx.ConnectorTokenSecuredClient, ctx.ClientKey)
	certChain := certs.DecodeCertChain(t, certResult.CertificateChain)
	securedClient := authentication.CreateCertClient(ctx.ClientKey, certChain...)

	return securedClient, configuration, certChain
}

func createCatalogRequest(t *testing.T, ctx *broker.SystemBrokerTestContext) *http.Request {
	req, err := http.NewRequest(http.MethodGet, ctx.SystemBrokerURL+"/v2/catalog", nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Broker-API-Version", "2.15")
	return req
}

func createORDServiceRequest(t *testing.T, ctx *broker.SystemBrokerTestContext, apiId, specId string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, ctx.ORDServiceURL+fmt.Sprintf("/api/%s/specification/%s", apiId, specId), nil)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)

	return req
}
