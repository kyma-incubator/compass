package tests

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/authentication"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

const (
	directorPath  = "/director/graphql"
	connectorPath = "/connector/graphql"
)

var (
	runtimeInput = &graphql.RuntimeInput{
		Name: "test-runtime",
	}
	dexToken string
)

func TestCallingCompassWithTooBigRequestBody(t *testing.T) {
	var (
		ctx    = context.TODO()
		err    error
		tenant = tenant.TestTenants.GetDefaultTenantID()
	)

	t.Log("Getting Dex id_token")
	dexToken, err = idtokenprovider.GetDexToken()
	require.NoError(t, err)

	t.Run("when request passes through gateway", func(t *testing.T) {
		t.Log("Compass Gateway URL: ", conf.CompassGatewayURL)
		authoizedClient := gql.NewAuthorizedHTTPClient(dexToken)

		t.Log("Creating a request with big payload...")
		bigBodyPOSTRequest, err := fixtures.FixHTTPBigBodyPOSTRequest(conf.CompassGatewayURL+directorPath, conf.RequestPayloadLimit)
		require.NoError(t, err)

		bigBodyPOSTRequest.Header.Set("Tenant", tenant)
		bigBodyPOSTRequest = bigBodyPOSTRequest.WithContext(context.TODO())

		t.Log("Executing request with big payload...")
		resp, err := authoizedClient.Do(bigBodyPOSTRequest)
		require.NoError(t, err)
		defer idtokenprovider.CloseRespBody(resp)

		t.Log("Executed request with big payload successfully")
		require.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)

		all, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(all), "Payload Too Large")
	})

	t.Run("when request passes through MTLS gateway", func(t *testing.T) {
		t.Log("Compass MTLS Gateway URL: ", conf.CompassMTLSGatewayURL)

		dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)
		logrus.Infof("Registering runtime with name: %s, within tenant: %s", runtimeInput.Name, tenant)

		runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, runtimeInput)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

		logrus.Infof("Generating one-time token for runtime with id: %s", runtime.ID)
		runtimeToken := fixtures.RequestOneTimeTokenForRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)
		oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

		logrus.Info("Generating private key for cert...")
		clientKey, err := certs.GenerateKey()
		require.NoError(t, err)
		tokenSecuredClient := clients.NewTokenSecuredClient(conf.CompassGatewayURL + "/connector/graphql")

		logrus.Infof("Generation certificate for runtime with id: %s", runtime.ID)
		certResult, _ := clients.GenerateRuntimeCertificate(t, oneTimeToken, tokenSecuredClient, clientKey)
		certChain := certs.DecodeCertChain(t, certResult.CertificateChain)
		certSecuredClient := authentication.CreateCertClient(clientKey, certChain...)

		t.Log("Creating a request with big payload...")
		bigBodyPOSTRequest, err := fixtures.FixHTTPBigBodyPOSTRequest(conf.CompassMTLSGatewayURL+directorPath, conf.RequestPayloadLimit)
		require.NoError(t, err)

		bigBodyPOSTRequest.Header.Set("Tenant", tenant)
		bigBodyPOSTRequest = bigBodyPOSTRequest.WithContext(context.TODO())

		t.Log("Executing request with big payload...")
		resp, err := certSecuredClient.Do(bigBodyPOSTRequest)
		require.NoError(t, err)
		defer idtokenprovider.CloseRespBody(resp)

		t.Log("Executed request with big payload successfully")
		require.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)

		all, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(all), "Payload Too Large")

	})
}
