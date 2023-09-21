package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	queryAPIForApplication := fixtures.FixAPIForApplicationWithDefaultPaginationRequest(application.ID)

	apiDefPage := graphql.APIDefinitionPageExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryAPIForApplication, &apiDefPage)
	require.NoError(t, err)

	assert.Equal(t, 1, apiDefPage.TotalCount)
	assert.Equal(t, api.ID, apiDefPage.Data[0].ID)
	example.SaveExample(t, queryAPIForApplication.Query(), "query api definition for application")
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

	apiAddRequest := fixtures.FixAddAPIToApplicationRequest(application.ID, inputGQL)
	apiDef := graphql.APIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, apiAddRequest, &apiDef)
	require.NoError(t, err)

	assert.NotNil(t, apiDef.ID)
	example.SaveExample(t, apiAddRequest.Query(), "add api definition to application")
}

func TestUpdateAPIDefinitionToApplication(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	newName := "TestUpdatePIDefinitionToApplication"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	addedAPI := fixtures.AddAPIToApplication(t, ctx, certSecuredGraphQLClient, application.ID)
	assert.NotNil(t, addedAPI.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(fixtures.FixAPIDefinitionInputWithName(newName))
	require.NoError(t, err)

	apiUpdateRequest := fixtures.FixUpdateAPIToApplicationRequest(addedAPI.ID, inputGQL)
	updatedAPI := graphql.APIDefinitionExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, apiUpdateRequest, &updatedAPI)
	require.NoError(t, err)

	assert.Equal(t, newName, updatedAPI.Name)
	assert.NotEqual(t, addedAPI.Name, updatedAPI.Name)
	example.SaveExample(t, apiUpdateRequest.Query(), "update api definition to application")
}
