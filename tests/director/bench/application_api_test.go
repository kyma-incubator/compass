package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
)

func BenchmarkApplicationsForRuntime(b *testing.B) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	appsCount := 5
	for i := 0; i < appsCount; i++ {
		app := fixtures.CreateApp(fmt.Sprintf("director-%d", i))
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, dexGraphQLClient, tenantID, app)
		defer fixtures.CleanupApplication(b, ctx, dexGraphQLClient, tenantID, &appResp)
		require.NoError(b, err)
	}

	//create runtime without normalization
	runtime := fixtures.FixRuntimeInput("runtime")
	(runtime.Labels)["scenarios"] = []string{conf.DefaultScenario}
	(runtime.Labels)["isNormalized"] = "false"

	rt, err := fixtures.RegisterRuntimeFromInputWithinTenant(b, ctx, dexGraphQLClient, tenantID, &runtime)
	defer fixtures.CleanupRuntime(b, ctx, dexGraphQLClient, tenantID, &rt)
	require.NoError(b, err)
	require.NotEmpty(b, rt.ID)

	request := fixtures.FixApplicationForRuntimeRequestWithPageSize(rt.ID, appsCount)
	request.Header.Set("Tenant", tenantID)

	res := struct {
		Result interface{} `json:"result"`
	}{}

	b.ResetTimer() // Reset timer after the initialization

	for i := 0; i < b.N; i++ {
		res.Result = &graphql.ApplicationPage{}

		err := dexGraphQLClient.Run(ctx, request, &res)

		//THEN
		require.NoError(b, err)
		require.Len(b, res.Result.(*graphql.ApplicationPage).Data, appsCount)
	}

	b.StopTimer() // Stop timer in order to exclude defers from the time
}
