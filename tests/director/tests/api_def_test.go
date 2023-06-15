package tests

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAPIDefinitionInApplication(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	api := fixtures.AddAPIToApplication(t, ctx, certSecuredGraphQLClient, application.ID)

	queryApiForApplication := fixtures.FixAPIForApplicationWithDefaultPaginationRequest(application.ID)
	apiDefPage := graphql.APIDefinitionPage{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryApiForApplication, &apiDefPage)
	require.NoError(t, err)

	assert.Equal(t, 1, apiDefPage.TotalCount)
	assert.Equal(t, api.ID, apiDefPage.Data[0].ID)
	saveExample(t, queryApiForApplication.Query(), "query api definition for application")
}

func TestAddAPIDefinitionToApplication(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(fixtures.FixAPIDefinitionInput())
	require.NoError(t, err)

	apiRequest := fixtures.FixAddAPIToApplicationRequest(application.ID, inputGQL)
	apiDef := graphql.APIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, apiRequest, &apiDef)
	require.NoError(t, err)

	assert.NotNil(t, apiDef.ID)
	saveExample(t, apiRequest.Query(), "add api definition to application")
}

func TestUpdatePIDefinitionToApplication(t *testing.T) {
	ctx := context.Background()

	newTargetURL := "http://localhost-new:6789"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	addedAPI := fixtures.AddAPIToApplication(t, ctx, certSecuredGraphQLClient, application.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(fixtures.FixAPIDefinitionInputWithTargetURL(newTargetURL))
	require.NoError(t, err)

	apiRequest := fixtures.FixUpdateAPIToApplicationRequest(application.ID, inputGQL)
	updatedAPI := graphql.APIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, apiRequest, &updatedAPI)
	require.NoError(t, err)

	assert.Equal(t, newTargetURL, updatedAPI.TargetURL)
	assert.NotEqual(t, addedAPI.TargetURL, updatedAPI.TargetURL)
	saveExample(t, apiRequest.Query(), "update api definition to application")
}
