package tests

import (
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/gql"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

const directorPath = "/director/graphql"

func TestCallingCompassGatewayWithTooBigRequestBody(t *testing.T) {
	t.Log("Getting Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	authoizedClient := gql.NewAuthorizedHTTPClient(dexToken)
	t.Log("Compass Gateway URL: ", conf.CompassGatewayURL)

	t.Log("Creating a request with big payload...")
	bigBodyPOSTRequest, err := fixtures.FixHTTPBigBodyPOSTRequest(conf.CompassGatewayURL+directorPath, conf.RequestPayloadLimit)
	require.NoError(t, err)

	bigBodyPOSTRequest.Header.Set("Tenant", tenant.TestTenants.GetDefaultTenantID())
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
}
