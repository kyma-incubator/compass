package bench

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
)

func BenchmarkApplicationsForRuntime(b *testing.B) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultTenantID()

	testScenario := "test-scenario"

	b.Logf("Creating formation with name: %q", testScenario)
	var formation graphql.Formation
	createFirstFormationReq := fixtures.FixCreateFormationRequest(testScenario)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createFirstFormationReq, &formation)
	defer func() {
		b.Logf("Deleting formation with name: %q", testScenario)
		deleteRequest := fixtures.FixDeleteFormationRequest(testScenario)
		var deleteFormation graphql.Formation
		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, deleteRequest, &deleteFormation)
		assertions.AssertNoErrorForOtherThanNotFound(b, err)
	}()
	require.NoError(b, err)
	require.Equal(b, testScenario, formation.Name)

	appsCount := 5
	for i := 0; i < appsCount; i++ {
		appInput := fixtures.CreateApp(fmt.Sprintf("director-%d", i))
		appInput.Labels = map[string]interface{}{
			"scenarios":                  []string{testScenario},
			conf.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C),
		}
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, certSecuredGraphQLClient, tenantID, appInput)
		defer fixtures.CleanupApplication(b, ctx, certSecuredGraphQLClient, tenantID, &appResp)
		require.NoError(b, err)
		require.NotEmpty(b, appResp.ID)
	}

	//create runtime without normalization
	runtimeInput := fixtures.FixRuntimeRegisterInput("runtime")
	(runtimeInput.Labels)["scenarios"] = []string{testScenario}
	(runtimeInput.Labels)["isNormalized"] = "false"

	runtime := fixtures.RegisterKymaRuntimeBench(b, ctx, certSecuredGraphQLClient, tenantID, runtimeInput, conf.GatewayOauth)
	defer fixtures.CleanupRuntime(b, ctx, certSecuredGraphQLClient, tenantID, &runtime)

	request := fixtures.FixApplicationForRuntimeRequestWithPageSize(runtime.ID, appsCount)
	request.Header.Set("Tenant", tenantID)

	res := struct {
		Result interface{} `json:"result"`
	}{}

	b.ResetTimer() // Reset timer after the initialization

	for i := 0; i < b.N; i++ {
		res.Result = &graphql.ApplicationPage{}

		err := certSecuredGraphQLClient.Run(ctx, request, &res)

		//THEN
		require.NoError(b, err)
		require.Len(b, res.Result.(*graphql.ApplicationPage).Data, appsCount)
	}

	b.StopTimer() // Stop timer in order to exclude defers from the time
}
