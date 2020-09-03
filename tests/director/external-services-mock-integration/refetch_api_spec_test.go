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

	pkgName := "test-bundle"
	pkgInput := graphql.BundleCreateInput{
		Name: pkgName,
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

	pkg := createBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, pkgInput)
	defer deleteBundle(t, ctx, dexGraphQLClient, tenant, pkg.ID)
	bundleID := pkg.ID
	assertSpecInBundleNotNil(t, pkg)
	spec := *pkg.APIDefinitions.Data[0].Spec.APISpec.Data

	var refetchedSpec graphql.APISpecExt
	apiID := pkg.APIDefinitions.Data[0].ID
	req := fixRefetchAPISpecRequest(apiID)

	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant, req, &refetchedSpec)
	require.NoError(t, err)

	require.NotNil(t, refetchedSpec.APISpec.Data)
	assert.NotEqual(t, spec, *refetchedSpec.APISpec.Data)

	pkg = getBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bundleID)

	assertSpecInBundleNotNil(t, pkg)
	assert.Equal(t, *refetchedSpec.APISpec.Data, *pkg.APIDefinitions.Data[0].Spec.Data)

}
