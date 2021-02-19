package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OpenAPISpec = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/beb8e5b6d8f3a644b8380e667a9376bc353e54dd/github-connector/internal/registration/configs/githubopenAPI.json"

func Test_FetchRequestAddApplicationWithAPI(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

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

	app := pkg.RegisterApplicationFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, appInput)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, app.ID)

	api := app.Bundles.Data[0].APIDefinitions.Data[0]

	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func Test_FetchRequestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

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
	api := pkg.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenant, bndl.ID, apiInput)
	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func TestFetchRequestAddBundleWithAPI(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

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

	bndl := pkg.CreateBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, bndlInput)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, bndl.APIDefinitions.Data[0].Spec.FetchRequest.Status.Condition)
}

func TestRefetchAPISpec(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

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

	bndl := pkg.CreateBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, bndlInput)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	spec := bndl.APIDefinitions.Data[0].Spec.Data

	var refetchedSpec graphql.APISpecExt
	req := pkg.FixRefetchAPISpecRequest(bndl.APIDefinitions.Data[0].ID)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &refetchedSpec)
	require.NoError(t, err)
	assert.Equal(t, spec, refetchedSpec.Data)

	saveExample(t, req.Query(), "refetch api spec")
}
