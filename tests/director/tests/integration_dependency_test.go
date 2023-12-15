package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddIntegrationDependencyToApplication(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	appName := "app-test"
	appNamespace := "test.ns"

	t.Log("Register application with application namespace")
	application, err := fixtures.RegisterApplicationWithApplicationNamespace(t, ctx, certSecuredGraphQLClient, appName, appNamespace, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)
	t.Logf("Successfully registered application with id %q", application.ID)

	inputGQL, err := testctx.Tc.Graphqlizer.IntegrationDependencyInputToGQL(fixIntegrationDependencyInput())
	require.NoError(t, err)

	t.Logf("Add integration dependency to application with id %q", application.ID)
	IntegrationDependencyAddRequest := fixAddIntegrationDependencyToApplicationRequest(application.ID, inputGQL)
	integrationDependency := graphql.IntegrationDependency{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, IntegrationDependencyAddRequest, &integrationDependency)
	require.NoError(t, err)

	assert.NotNil(t, integrationDependency.ID)
	t.Logf("Successfully added integration dependency with id %q to application with id %q", integrationDependency.ID, application.ID)

	t.Logf("Delete integration dependency wiht id %q", integrationDependency.ID)
	var deletedIntegrationDependency graphql.IntegrationDependency
	deleteReq := fixDeleteIntegrationDependencyRequest(integrationDependency.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteReq, &deletedIntegrationDependency)
	require.NoError(t, err)

	t.Logf("Check if integration dependency with id %q was deleted", integrationDependency.ID)
	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	require.Empty(t, app.IntegrationDependencies.Data)
	t.Logf("Integration Dependency with id %q was sucessfully deleted from application with id %q", integrationDependency.ID, application.ID)

	example.SaveExample(t, IntegrationDependencyAddRequest.Query(), "add integration dependency to application")
	example.SaveExample(t, deleteReq.Query(), "delete integration dependency")
}

func fixIntegrationDependencyInput() graphql.IntegrationDependencyInput {
	mandatory := false
	return graphql.IntegrationDependencyInput{
		Name:          "Int dep name",
		Mandatory:     &mandatory,
		Visibility:    str.Ptr("public"),
		ReleaseStatus: str.Ptr("active"),
		Description:   str.Ptr("int dep desc"),
		Aspects: []*graphql.AspectInput{
			{
				Name:        "Aspect name",
				Description: str.Ptr("aspect desc"),
				Mandatory:   &mandatory,
				APIResources: []*graphql.AspectAPIDefinitionInput{
					{
						OrdID: "ns:apiResource:API_ID:v1",
					},
				},
				EventResources: []*graphql.AspectEventDefinitionInput{
					{
						OrdID: "ns:eventResource:EVENT_ID:v1",
						Subset: []*graphql.AspectEventDefinitionSubsetInput{
							{
								EventType: str.Ptr("sap.billing.sb.Subscription.Created.v1"),
							},
						},
					},
				},
			},
		},
	}
}

func fixAddIntegrationDependencyToApplicationRequest(appID, integrationDependencyInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addIntegrationDependencyToApplication(appID: "%s", in: %s) {
				%s
			}
		}
		`, appID, integrationDependencyInputGQL, testctx.Tc.GQLFieldsProvider.ForIntegrationDependency()))
}

func fixDeleteIntegrationDependencyRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteIntegrationDependency(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForIntegrationDependency()))
}
