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
							Username: testConfig.BasicCredentialsUsername,
							Password: testConfig.BasicCredentialsPassword,
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
							ClientID:     testConfig.AppClientID,
							ClientSecret: testConfig.AppClientSecret,
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

			appName := "app-test-bundle"
			application := registerApplication(t, ctx, dexGraphQLClient, appName, tenant)
			defer unregisterApplication(t, dexGraphQLClient, application.ID, tenant)

			bndlName := "test-bundle"
			bndlInput := graphql.BundleCreateInput{
				Name: bndlName,
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

			bndl := createBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, bndlInput)
			defer deleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)
			bndlID := bndl.ID
			assertSpecInBundleNotNil(t, bndl)
			spec := *bndl.APIDefinitions.Data[0].Spec.APISpec.Data

			var refetchedSpec graphql.APISpecExt
			apiID := bndl.APIDefinitions.Data[0].ID
			req := fixRefetchAPISpecRequest(apiID)

			err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant, req, &refetchedSpec)
			require.NoError(t, err)

			require.NotNil(t, refetchedSpec.APISpec.Data)
			assert.NotEqual(t, spec, *refetchedSpec.APISpec.Data)

			bndl = getBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlID)

			assertSpecInBundleNotNil(t, bndl)
			assert.Equal(t, *refetchedSpec.APISpec.Data, *bndl.APIDefinitions.Data[0].Spec.Data)

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
							Username: "wrong_" + testConfig.BasicCredentialsUsername,
							Password: "wrong_" + testConfig.BasicCredentialsPassword,
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
							ClientID:     "wrong_" + testConfig.AppClientID,
							ClientSecret: "wrong_" + testConfig.AppClientSecret,
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

			appName := "app-test-bundle"
			application := registerApplication(t, ctx, dexGraphQLClient, appName, tenant)
			defer unregisterApplication(t, dexGraphQLClient, application.ID, tenant)

			bndlName := "test-bundle"
			bndlInput := graphql.BundleCreateInput{
				Name: bndlName,
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

			bndl := createBundleWithInput(t, ctx, dexGraphQLClient, tenant, application.ID, bndlInput)
			defer deleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

			assert.True(t, len(bndl.APIDefinitions.Data) > 0)
			assert.NotNil(t, bndl.APIDefinitions.Data[0])
			assert.NotNil(t, bndl.APIDefinitions.Data[0].Spec)
			assert.Nil(t, bndl.APIDefinitions.Data[0].Spec.Data)
		})
	}
}
