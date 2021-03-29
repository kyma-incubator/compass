package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OpenAPISpec = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/beb8e5b6d8f3a644b8380e667a9376bc353e54dd/github-connector/internal/registration/configs/githubopenAPI.json"

func Test_FetchRequestAddApplicationWithAPI(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appInput := graphql.ApplicationRegisterInput{
		Name: "test",
		Bundles: []*graphql.BundleCreateInput{{
			Name: "test",
			APIDefinitions: []*graphql.APIDefinitionInput{{
				Name:      "test",
				TargetURL: "https://target.url",
				Spec: &graphql.APISpecInput{
					Format: graphql.SpecFormatJSON,
					Type:   graphql.APISpecTypeOpenAPI,
					FetchRequest: &graphql.FetchRequestInput{
						URL: OpenAPISpec,
					},
				},
			}},
		}},
	}

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantId, appInput)
	assert.NoError(t, err)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	api := app.Bundles.Data[0].APIDefinitions.Data[0]

	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func Test_FetchRequestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, dexGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	apiInput := graphql.APIDefinitionInput{
		Name:      "test",
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: OpenAPISpec,
			},
		},
	}
	api := fixtures.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenantId, bndl.ID, apiInput)
	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func TestFetchRequestAddBundleWithAPI(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndlName := "test-bundle"
	bndlInput := graphql.BundleCreateInput{
		Name: bndlName,
		APIDefinitions: []*graphql.APIDefinitionInput{{
			Name:      "test",
			TargetURL: "https://target.url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: OpenAPISpec,
				},
			},
		},
		},
	}

	bndl := fixtures.CreateBundleWithInput(t, ctx, dexGraphQLClient, tenantId, application.ID, bndlInput)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, bndl.APIDefinitions.Data[0].Spec.FetchRequest.Status.Condition)
}

func TestRefetchAPISpec(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenantId)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, application.ID)

	bndlName := "test-bundle"
	bndlInput := graphql.BundleCreateInput{
		Name: bndlName,
		APIDefinitions: []*graphql.APIDefinitionInput{{
			Name:      "test",
			TargetURL: "https://target.url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: OpenAPISpec,
				},
			},
		},
		},
	}

	bndl := fixtures.CreateBundleWithInput(t, ctx, dexGraphQLClient, tenantId, application.ID, bndlInput)
	defer fixtures.DeleteBundle(t, ctx, dexGraphQLClient, tenantId, bndl.ID)

	spec := bndl.APIDefinitions.Data[0].Spec.Data

	var refetchedSpec graphql.APISpecExt
	req := fixtures.FixRefetchAPISpecRequest(bndl.APIDefinitions.Data[0].ID)

	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, req, &refetchedSpec)
	require.NoError(t, err)
	assert.Equal(t, spec, refetchedSpec.Data)

	saveExample(t, req.Query(), "refetch api spec")
}
