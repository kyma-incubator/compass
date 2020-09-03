package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OpenAPISpec = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/beb8e5b6d8f3a644b8380e667a9376bc353e54dd/github-connector/internal/registration/configs/githubopenAPI.json"

func Test_FetchRequestAddApplicationWithAPI(t *testing.T) {
	ctx := context.Background()

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

	app := registerApplicationFromInput(t, ctx, appInput)
	defer unregisterApplication(t, app.ID)

	api := app.Bundles.Data[0].APIDefinitions.Data[0]

	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func Test_FetchRequestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

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
	api := addAPIToBundleWithInput(t, ctx, bundle.ID, apiInput)
	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func TestFetchRequestAddBundleWithAPI(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundleInput := graphql.BundleCreateInput{
		Name: bundleName,
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

	bundle := createBundleWithInput(t, ctx, application.ID, bundleInput)
	defer deleteBundle(t, ctx, bundle.ID)

	assert.NotNil(t, bundle.APIDefinitions.Data[0].Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, bundle.APIDefinitions.Data[0].Spec.FetchRequest.Status.Condition)
}

func TestRefetchAPISpec(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundleInput := graphql.BundleCreateInput{
		Name: bundleName,
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

	bundle := createBundleWithInput(t, ctx, application.ID, bundleInput)
	defer deleteBundle(t, ctx, bundle.ID)

	spec := bundle.APIDefinitions.Data[0].Spec.Data

	var refetchedSpec graphql.APISpecExt
	req := fixRefetchAPISpecRequest(bundle.APIDefinitions.Data[0].ID)

	err := tc.RunOperation(ctx, req, &refetchedSpec)
	require.NoError(t, err)
	assert.Equal(t, spec, refetchedSpec.Data)

	saveExample(t, req.Query(), "refetch api spec")
}
