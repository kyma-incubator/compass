package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

func BenchmarkApplicationsForRuntime(b *testing.B) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	appsCount := 5
	apps := make([]graphql.ApplicationRegisterInput, 0, appsCount)
	for i := 0; i < appsCount; i++ {
		apps = append(apps, fixtures.CreateApp(fmt.Sprintf("%d", i)))
	}

	for _, app := range apps {
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, dexGraphQLClient, tenantID, app)
		require.NoError(b, err)
		defer fixtures.UnregisterApplication(b, ctx, dexGraphQLClient, tenantID, appResp.ID)
	}

	//create runtime without normalization
	runtime := fixtures.FixRuntimeInput("runtime")
	(runtime.Labels)["scenarios"] = []string{conf.DefaultScenario}
	(runtime.Labels)["isNormalized"] = "false"

	rt := fixtures.RegisterRuntimeFromInputWithinTenant(b, ctx, dexGraphQLClient, tenantID, &runtime)
	defer fixtures.UnregisterRuntime(b, ctx, dexGraphQLClient, tenantID, rt.ID)

	request := fixtures.FixApplicationForRuntimeRequestWithPageSize(rt.ID, appsCount)

	b.ResetTimer() // Reset timer after the initialization

	for i := 0; i < b.N; i++ {
		applicationPage := graphql.ApplicationPage{}

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(b, err)
		require.Len(b, applicationPage.Data, appsCount)
	}
}
