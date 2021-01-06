package external_services_mock_integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefetchAPISpecDifferentSpec(t *testing.T) {

	testCases := []struct {
		Name         string
		FetchRequest *graphql.FetchRequestInput
	}{
		{
			Name: "Success without credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "external-api/unsecured/spec",
			},
		},
		{
			Name: "Success with basic credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "external-api/secured/basic/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: "admin",
							Password: "admin",
						},
					},
				},
			},
		},
		{
			Name: "Success with oauth",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "external-api/secured/oauth/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							ClientID:     "client_id",
							ClientSecret: "client_secret",
							URL:          testConfig.ExternalServicesMockBaseURL + "oauth/token",
						},
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			ctx := context.Background()
			tenant := testConfig.DefaultTenant

			t.Log("Get Dex id_token")
			dexToken, err := idtokenprovider.GetDexToken()
			require.NoError(t, err)

			dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

			appName := "app-test-package"
			application := registerApplication(t, ctx, dexGraphQLClient, appName, tenant)
			defer unregisterApplication(t, dexGraphQLClient, application.ID, tenant)

			pkgName := "test-package"
			pkgInput := graphql.PackageCreateInput{
				Name: pkgName,
				APIDefinitions: []*graphql.APIDefinitionInput{{
					Name:      "test",
					TargetURL: "https://target.url",
					Spec: &graphql.APISpecInput{
						Format:       graphql.SpecFormatJSON,
						Type:         graphql.APISpecTypeOpenAPI,
						FetchRequest: testCase.FetchRequest,
					},
				},
				},
			}

			pkg := createPackageWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, pkgInput)
			defer deletePackage(t, ctx, dexGraphQLClient, tenant, pkg.ID)
			pkgID := pkg.ID
			assertSpecInPackageNotNil(t, pkg)
			spec := *pkg.APIDefinitions.Data[0].Spec.APISpec.Data

			var refetchedSpec graphql.APISpecExt
			apiID := pkg.APIDefinitions.Data[0].ID
			req := fixRefetchAPISpecRequest(apiID)

			err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant, req, &refetchedSpec)
			require.NoError(t, err)

			require.NotNil(t, refetchedSpec.APISpec.Data)
			assert.NotEqual(t, spec, *refetchedSpec.APISpec.Data)

			pkg = getPackage(t, ctx, dexGraphQLClient, tenant, application.ID, pkgID)

			assertSpecInPackageNotNil(t, pkg)
			assert.Equal(t, *refetchedSpec.APISpec.Data, *pkg.APIDefinitions.Data[0].Spec.Data)

		})
	}

}

func TestCreateAPIWithFetchRequestWithWrongCredentials(t *testing.T) {

	testCases := []struct {
		Name         string
		FetchRequest *graphql.FetchRequestInput
	}{
		{
			Name: "API creation fails when fetch request has wrong basic credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "external-api/secured/basic/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: "admin",
							Password: "",
						},
					},
				},
			},
		},
		{
			Name: "API creation fails when fetch request has wrong oauth client credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "external-api/secured/oauth/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							ClientID:     "wrong_id",
							ClientSecret: "wrong_secret",
							URL:          testConfig.ExternalServicesMockBaseURL + "oauth/token",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			ctx := context.Background()
			tenant := testConfig.DefaultTenant

			t.Log("Get Dex id_token")
			dexToken, err := idtokenprovider.GetDexToken()
			require.NoError(t, err)

			dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

			appName := "app-test-package"
			application := registerApplication(t, ctx, dexGraphQLClient, appName, tenant)
			defer unregisterApplication(t, dexGraphQLClient, application.ID, tenant)

			pkgName := "test-package"
			pkgInput := graphql.PackageCreateInput{
				Name: pkgName,
				APIDefinitions: []*graphql.APIDefinitionInput{{
					Name:      "test",
					TargetURL: "https://target.url",
					Spec: &graphql.APISpecInput{
						Format:       graphql.SpecFormatJSON,
						Type:         graphql.APISpecTypeOpenAPI,
						FetchRequest: testCase.FetchRequest,
					},
				},
				},
			}

			pkg := createPackageWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, pkgInput)
			defer deletePackage(t, ctx, dexGraphQLClient, tenant, pkg.ID)

			assert.True(t, len(pkg.APIDefinitions.Data) > 0)
			assert.NotNil(t, pkg.APIDefinitions.Data[0])
			assert.NotNil(t, pkg.APIDefinitions.Data[0].Spec)
			assert.Nil(t, pkg.APIDefinitions.Data[0].Spec.Data)
		})
	}
}
