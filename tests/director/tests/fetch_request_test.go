package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/tests/example"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	OpenAPISpec   = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/beb8e5b6d8f3a644b8380e667a9376bc353e54dd/github-connector/internal/registration/configs/githubopenAPI.json"
	AsyncAPISpec  = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/2fcc4728f63cdf23656417d9f85679a02367df97/slack-connector/internal/registration/configs/slackasyncapi.json"
	MDDocumentURL = "https://raw.githubusercontent.com/kyma-incubator/compass/v1.4.2/README.md"
)

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

	app, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appInput)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
	assert.NoError(t, err)

	api := app.Bundles.Data[0].APIDefinitions.Data[0]

	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func Test_FetchRequestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

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
	api := fixtures.AddAPIToBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID, apiInput)
	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func TestFetchRequestAddBundleWithAPI(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

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

	bndl := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlInput)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, bndl.APIDefinitions.Data[0].Spec.FetchRequest.Status.Condition)
}

func TestRefetchAPISpec(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

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

	bndl := fixtures.CreateBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlInput)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	spec := bndl.APIDefinitions.Data[0].Spec.Data

	var refetchedSpec graphql.APISpecExt
	req := fixtures.FixRefetchAPISpecRequest(bndl.APIDefinitions.Data[0].ID)

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &refetchedSpec)
	require.NoError(t, err)
	assert.Equal(t, spec, refetchedSpec.Data)

	example.SaveExample(t, req.Query(), "refetch api spec")
}
