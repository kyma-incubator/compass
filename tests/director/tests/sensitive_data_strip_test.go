package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveDataStrip(t *testing.T) {
	const (
		appName     = "application-test"
		runtimeName = "runtime-test"
		intSysName  = "integration-system-test"

		appAuthsReadScope       = "application.auths:read"
		runtimeAuthsReadScope   = "runtime.auths:read"
		intSystemAuthsReadScope = "integration_system.auths:read"
		webhooksReadScopes      = "application.webhooks:read application_template.webhooks:read"
		bundleInstanceAuthScope = "bundle.instance_auths:read"
		fetchRequestsReadScopes = "document.fetch_request:read event_spec.fetch_request:read api_spec.fetch_request:read"
	)

	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)
	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// REGISTER RUNTIME
	t.Log(fmt.Sprintf("Registering runtime %q", runtimeName))
	runtimeRegInput := fixtures.FixRuntimeInput(runtimeName)
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &runtimeRegInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	// register runtime oauth client
	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	// register application
	t.Log(fmt.Sprintf("Registering application %q", appName))
	appInput := appWithAPIsAndEvents(appName)
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantId, appInput)
	assert.NoError(t, err)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	// assert document, event and api definitions are present
	assert.Len(t, app.Bundles.Data, 1)
	bndl := app.Bundles.Data[0]

	assert.Len(t, bndl.EventDefinitions.Data, 1)
	assert.Len(t, bndl.APIDefinitions.Data, 1)
	assert.Len(t, bndl.Documents.Data, 1)

	// register application oauth client
	appAuth := fixtures.RequestClientCredentialsForApplication(t, context.Background(), dexGraphQLClient, tenantId, app.ID)
	appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, appOauthCredentialData.ClientSecret)
	require.NotEmpty(t, appOauthCredentialData.ClientID)

	intSysInput := graphql.IntegrationSystemInput{Name: intSysName}
	intSys, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(intSysInput)
	require.NoError(t, err)

	registerIntegrationSystemRequest := fixtures.FixRegisterIntegrationSystemRequest(intSys)
	output := graphql.IntegrationSystemExt{}

	// WHEN
	t.Log("Register integration system")

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerIntegrationSystemRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, output.ID)

	// register integration system oauth client
	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, context.Background(), dexGraphQLClient, tenantId, app.ID)
	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, appOauthCredentialData.ClientSecret)
	require.NotEmpty(t, appOauthCredentialData.ClientID)

	// assign runtime and app to the same scenario
	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios[:1])

	// set application scenarios label
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, ScenariosLabel, scenarios[1:])
	defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, ScenariosLabel, scenarios[:1])

	// set runtime scenarios label
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
	defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

	// create bundle instance auth
	t.Log(fmt.Sprintf("Creating bundle instance auths %q with bundle with APIDefinition and ", appName))
	authCtx, inputParams := fixtures.FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := fixtures.FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := fixtures.FixRequestBundleInstanceAuthCreationRequest(bndl.ID, bndlInstanceAuthRequestInputStr)

	t.Log("Issue a Hydra token with Client Credentials")
	runtimeOAuthGraphQLClient := gqlClient(t, rtmOauthCredentialData, "runtime:write application:read")
	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Log("Request bundle instance auth creation")
	output := graphql.BundleInstanceAuth{}
	err = runtimeConsumer.Run(bndlInstanceAuthCreationRequestReq, runtimeOAuthGraphQLClient, &output)

	require.NoError(t, err)
	assertions.AssertBundleInstanceAuthInput(t, bndlInstanceAuthRequestInput, output)

	// Fetch Application with bundles
	bundlesForApplicationReq := fixtures.FixGetBundlesRequest(app.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = runtimeConsumer.Run(bundlesForApplicationReq, runtimeOAuthGraphQLClient, &bndlFromAPI)
	require.NoError(t, err)

	// Assert the bundle instance auths exists
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data))
	require.Equal(t, 1, len(bndlFromAPI.Bundles.Data[0].InstanceAuths))

}

func gqlClient(t *testing.T, creds *graphql.OAuthCredentialData, scopes string) *gcli.Client {
	accessToken := token.GetAccessToken(t, creds, scopes)
	return gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)
}

func appWithAPIsAndEvents(name string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name: name,
		Bundles: []*graphql.BundleCreateInput{{
			Name: "test",
			APIDefinitions: []*graphql.APIDefinitionInput{{
				Name:      "test-api-def",
				TargetURL: "https://target.url",
				Spec: &graphql.APISpecInput{
					Format: graphql.SpecFormatJSON,
					Type:   graphql.APISpecTypeOpenAPI,
					FetchRequest: &graphql.FetchRequestInput{
						URL: OpenAPISpec,
					},
				},
			}},
			EventDefinitions: []*graphql.EventDefinitionInput{{
				Name: "test-event-def",
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL: AsyncAPISpec,
					},
				},
			}},
			Documents: []*graphql.DocumentInput{{
				Title:  "test-document",
				Format: graphql.DocumentFormatMarkdown,
				FetchRequest: &graphql.FetchRequestInput{
					URL: MDDocumentURL,
				},
			}},
		}},
	}
}
