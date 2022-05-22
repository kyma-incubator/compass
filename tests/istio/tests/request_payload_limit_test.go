package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/authentication"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	directorPath  = "/director/graphql"
	connectorPath = "/connector/graphql"
)

func TestCallingCompassGateways(t *testing.T) {
	var (
		ctx          = context.TODO()
		err          error
		tenant       = conf.DefaultTenant
		runtimeInput = &graphql.RuntimeRegisterInput{
			Name: "test-runtime",
			Labels: graphql.Labels{
				conf.SelfRegDistinguishLabelKey: []interface{}{conf.SelfRegDistinguishLabelValue},
				tenantfetcher.RegionKey:         conf.SelfRegRegion,
			},
		}
	)

	authorizedClient := gql.NewCertAuthorizedHTTPClient(certCache.Get().PrivateKey, certCache.Get().Certificate, conf.SkipSSLValidation)

	logrus.Infof("Registering runtime with name: %s, within tenant: %s", runtimeInput.Name, tenant)

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenant, runtimeInput)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	logrus.Infof("Generating one-time token for runtime with id: %s", runtime.ID)
	runtimeToken := fixtures.RequestOneTimeTokenForRuntime(t, ctx, certSecuredGraphQLClient, tenant, runtime.ID)
	oneTimeToken := &externalschema.Token{Token: runtimeToken.Token}

	logrus.Info("Generating private key for cert...")
	clientKey, err := certs.GenerateKey()
	require.NoError(t, err)
	tokenSecuredClient := clients.NewTokenSecuredClient(conf.CompassGatewayURL + connectorPath)

	logrus.Infof("Generation certificate for runtime with id: %s", runtime.ID)
	certResult, _ := clients.GenerateRuntimeCertificate(t, oneTimeToken, tokenSecuredClient, clientKey)
	certChain := certs.DecodeCertChain(t, certResult.CertificateChain)
	certSecuredClient := authentication.CreateCertClient(clientKey, certChain...)

	tests := []struct {
		negativeDescription string
		positiveDescription string
		client              *http.Client
		url                 string
		requestSize         int
	}{
		{
			negativeDescription: "fails when request is over 5MB and passes through gateway",
			positiveDescription: "succeeds for a regular applications request passing through gateway",
			url:                 conf.DirectorExternalCertSecuredURL,
			client:              authorizedClient,
			requestSize:         conf.RequestPayloadLimit,
		},
		{
			negativeDescription: "fails when request is over 2MB and reaches director",
			positiveDescription: "succeeds for a regular applications request reaching director",
			url:                 conf.DirectorExternalCertSecuredURL,
			client:              authorizedClient,
			requestSize:         2097152,
		},
		{
			negativeDescription: "fails when request is over 5MB and passes through MTLS gateway",
			positiveDescription: "succeeds for a regular applications request passing through MTLS gateway",
			url:                 conf.CompassMTLSGatewayURL + directorPath,
			client:              certSecuredClient,
			requestSize:         conf.RequestPayloadLimit,
		},
	}

	for _, test := range tests {
		t.Run(test.negativeDescription, func(t *testing.T) {
			t.Log("Creating a request with big payload...")
			bigBodyPOSTRequest := getHTTPBigBodyPOSTRequest(t, test.url, tenant, test.requestSize)

			t.Log("Executing request with big payload...")
			resp, err := test.client.Do(bigBodyPOSTRequest)
			require.NoError(t, err)

			defer func() {
				err := resp.Body.Close()
				if err != nil {
					logrus.Printf("WARNING: Unable to close response body. Cause: %v", err)
				}
			}()

			t.Log("Successfully executed request with big payload")

			require.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)

			t.Log("Response checking for error message ...")
			all, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(all), "Payload Too Large")
		})

		t.Run(test.positiveDescription, func(t *testing.T) {
			t.Log("Creating a request for applications...")
			applicationsPOSTRequest := getHTTPApplicationsPOSTRequest(t, test.url, tenant)

			t.Log("Executing request for applications...")
			appsResp, err := test.client.Do(applicationsPOSTRequest)
			require.NoError(t, err)
			defer func() {
				err := appsResp.Body.Close()
				if err != nil {
					logrus.Printf("WARNING: Unable to close response body. Cause: %v", err)
				}
			}()
			t.Log("Successfully executed request for applications")

			require.Equal(t, http.StatusOK, appsResp.StatusCode)
		})
	}
}

func getHTTPBigBodyPOSTRequest(t *testing.T, url, tenant string, bodySize int) *http.Request {
	var b strings.Builder
	b.Grow(bodySize)
	for i := 0; i < bodySize; i++ {
		b.WriteByte('a')
	}
	s := b.String()
	applicationsGQLRequest := fixtures.FixGetApplicationRequest(s)
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     applicationsGQLRequest.Query(),
		Variables: applicationsGQLRequest.Vars(),
	}

	var requestBuffer bytes.Buffer
	err := json.NewEncoder(&requestBuffer).Encode(requestBodyObj)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, &requestBuffer)
	require.NoError(t, err)

	req.Header.Set("Tenant", tenant)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req = req.WithContext(context.TODO())
	return req
}

func getHTTPApplicationsPOSTRequest(t *testing.T, url, tenant string) *http.Request {
	applicationsGQLRequest := fixtures.FixGetApplicationsRequestWithPagination()
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     applicationsGQLRequest.Query(),
		Variables: applicationsGQLRequest.Vars(),
	}

	var requestBuffer bytes.Buffer
	err := json.NewEncoder(&requestBuffer).Encode(requestBodyObj)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, url, &requestBuffer)
	require.NoError(t, err)

	req.Header.Set("Tenant", tenant)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req = req.WithContext(context.TODO())
	return req
}
