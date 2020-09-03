package external_services_mock_integration

import (
	"context"
	"net/url"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefetchAPISpecDifferentSpec(t *testing.T) {
	ctx := context.Background()
	tenant := testConfig.DefaultTenant

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer unregisterApplication(t, dexGraphQLClient, application.ID, tenant)

	externalServicesURL, err := url.Parse(testConfig.ExternalServicesMockBaseURL)
	require.NoError(t, err)
	externalServicesURL.Path = "external-api/spec"

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
					URL: externalServicesURL.String(),
				},
			},
		},
		},
	}

	bundle := createBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, bundleInput)
	defer deleteBundle(t, ctx, dexGraphQLClient, tenant, bundle.ID)
	bundleID := bundle.ID
	assertSpecInBundleNotNil(t, bundle)
	spec := *bundle.APIDefinitions.Data[0].Spec.APISpec.Data

	var refetchedSpec graphql.APISpecExt
	apiID := bundle.APIDefinitions.Data[0].ID
	req := fixRefetchAPISpecRequest(apiID)

	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant, req, &refetchedSpec)
	require.NoError(t, err)

	require.NotNil(t, refetchedSpec.APISpec.Data)
	assert.NotEqual(t, spec, *refetchedSpec.APISpec.Data)

	bundle = getBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bundleID)

	assertSpecInBundleNotNil(t, bundle)
	assert.Equal(t, *refetchedSpec.APISpec.Data, *bundle.APIDefinitions.Data[0].Spec.Data)

}
