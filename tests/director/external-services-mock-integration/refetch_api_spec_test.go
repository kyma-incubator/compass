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
		Name          string
		FetchRequest  *graphql.FetchRequestInput
		ShouldRefetch bool
	}{
		{
			Name: "Success without credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "/external-api/unsecured/spec",
			},
			ShouldRefetch: true,
		},
		{
			Name: "Success with basic credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "/external-api/secured/basic/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: "admin",
							Password: "admin",
						},
					},
				},
			},
			ShouldRefetch: true,
		},
		{
			Name: "Wrong basic credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "/external-api/secured/basic/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Basic: &graphql.BasicCredentialDataInput{
							Username: "admin",
							Password: "",
						},
					},
				},
			},
			ShouldRefetch: false,
		},
		{
			Name: "Success with oauth",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "/external-api/secured/oauth/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							ClientID:     "client_id",
							ClientSecret: "client_secret",
							URL:          testConfig.ExternalServicesMockBaseURL + "/external-api/secured/oauth/token",
						},
					},
				},
			},
			ShouldRefetch: true,
		},
		{
			Name: "Wrong client credentials",
			FetchRequest: &graphql.FetchRequestInput{
				URL: testConfig.ExternalServicesMockBaseURL + "/external-api/secured/oauth/spec",
				Auth: &graphql.AuthInput{
					Credential: &graphql.CredentialDataInput{
						Oauth: &graphql.OAuthCredentialDataInput{
							ClientID:     "wrong_id",
							ClientSecret: "wrong_secret",
							URL:          testConfig.ExternalServicesMockBaseURL + "/external-api/secured/oauth/token",
						},
					},
				},
			},
			ShouldRefetch: false,
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
			if testCase.ShouldRefetch {
				assert.NotEqual(t, spec, *refetchedSpec.APISpec.Data)
			} else {
				assert.Equal(t, spec, *refetchedSpec.APISpec.Data)
			}

			pkg = getPackage(t, ctx, dexGraphQLClient, tenant, application.ID, pkgID)

			assertSpecInPackageNotNil(t, pkg)
			if testCase.ShouldRefetch {
				assert.Equal(t, *refetchedSpec.APISpec.Data, *pkg.APIDefinitions.Data[0].Spec.Data)
			} else {
				assert.NotEqual(t, *refetchedSpec.APISpec.Data, *pkg.APIDefinitions.Data[0].Spec.Data)
			}
		})
	}

}
