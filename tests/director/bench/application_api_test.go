package bench

import (
	"context"
	"fmt"
	"testing"

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
	require.NoError(b, err)
	require.Equal(b, testScenario, formation.Name)

	defer func() {
		b.Logf("Deleting formation with name: %q", testScenario)
		deleteRequest := fixtures.FixDeleteFormationRequest(testScenario)
		var deleteFormation graphql.Formation
		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, deleteRequest, &deleteFormation)
		assertions.AssertNoErrorForOtherThanNotFound(b, err)
	}()

	appsCount := 5
	for i := 0; i < appsCount; i++ {
		appInput := fixtures.CreateApp(fmt.Sprintf("director-%d", i))
		appInput.Labels = map[string]interface{}{"scenarios": []string{testScenario}}
		appResp, err := fixtures.RegisterApplicationFromInput(b, ctx, certSecuredGraphQLClient, tenantID, appInput)
		defer fixtures.CleanupApplication(b, ctx, certSecuredGraphQLClient, tenantID, &appResp)
		require.NoError(b, err)
	}

	//create runtime without normalization
	runtime := fixtures.FixRuntimeRegisterInput("runtime")
	(runtime.Labels)["scenarios"] = []string{testScenario}
	(runtime.Labels)["isNormalized"] = "false"

	rt, err := fixtures.RegisterRuntimeFromInputWithinTenant(b, ctx, certSecuredGraphQLClient, tenantID, &runtime)
	defer fixtures.CleanupRuntime(b, ctx, certSecuredGraphQLClient, tenantID, &rt)
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

		err := certSecuredGraphQLClient.Run(ctx, request, &res)

		//THEN
		require.NoError(b, err)
		require.Len(b, res.Result.(*graphql.ApplicationPage).Data, appsCount)
	}

	b.StopTimer() // Stop timer in order to exclude defers from the time
}
