package bench

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/request"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	tenantHeader = "Tenant"
)

func BenchmarkSystemBundles(b *testing.B) {
	//GIVEN
	ctx := context.Background()
	defaultTestTenant := tenant.TestTenants.GetDefaultTenantID()

	appsCount := 15
	for i := 0; i < appsCount; i++ {
		appInput := fixtures.CreateApp(fmt.Sprintf("ord-service-%d", i))
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, certSecuredGraphQLClient, defaultTestTenant, appInput)
		defer fixtures.CleanupApplication(b, ctx, certSecuredGraphQLClient, defaultTestTenant, &appResp)
		require.NoError(b, err)
	}

	b.Log("Create integration system")
	var intSys directorSchema.IntegrationSystemExt // needed so the 'defer' can be above the integration system registration
	defer fixtures.CleanupIntegrationSystem(b, ctx, certSecuredGraphQLClient, "", &intSys)
	intSys = fixtures.RegisterIntegrationSystem(b, ctx, certSecuredGraphQLClient, "", "test-int-system")

	var intSystemCredentials directorSchema.IntSysSystemAuth // needed so the 'defer' can be above the integration system auth creation
	defer fixtures.DeleteSystemAuthForIntegrationSystem(b, ctx, certSecuredGraphQLClient, &intSystemCredentials)
	intSystemCredentials = fixtures.RequestClientCredentialsForIntegrationSystem(b, ctx, certSecuredGraphQLClient, "", intSys.ID)
	require.NotEmpty(b, intSystemCredentials)

	intSystemHttpClient, err := clients.NewIntegrationSystemClient(ctx, &intSystemCredentials)
	require.NoError(b, err)

	b.ResetTimer() // Reset timer after the initialization

	for i := 0; i < b.N; i++ {
		//WHEN
		respBody := makeRequestWithHeaders(b, intSystemHttpClient, fmt.Sprintf("%s/systemInstances?$expand=consumptionBundles($expand=apis,events)&$format=json", testConfig.ORDServiceURL), map[string][]string{tenantHeader: {defaultTestTenant}})

		//THEN
		require.Len(b, (gjson.Get(respBody, "value")).Array(), appsCount)
	}

	b.StopTimer() // Stop timer in order to exclude defers from the time
}

func makeRequestWithHeaders(b *testing.B, httpClient *http.Client, url string, headers map[string][]string) string {
	return request.MakeRequestWithHeadersAndStatusExpect(b, httpClient, url, headers, http.StatusOK, testConfig.ORDServiceDefaultResponseType)
}
