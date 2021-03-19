package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func TestSensitiveDataStrip(t *testing.T) {
	const (
		appName     = "application-test"
		runtimeName = "runtime-test"
		intSysName  = "integration-system-test"
	)

	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)
	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// CREATE APP TEMPLATE
	t.Log("Creating application template")
	appTmpInput := fixtures.FixApplicationTemplateWithWebhook("app-template-test")
	appTemplate := fixtures.CreateApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTmpInput)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)

	// REGISTER RUNTIME
	t.Log(fmt.Sprintf("Registering runtime %q", runtimeName))
	runtimeRegInput := fixtures.FixRuntimeInput(runtimeName)
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &runtimeRegInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	// REQUEST RUNTIME OAUTH CLIENT
	t.Log(fmt.Sprintf("Requesting OAuth client for runtime %q", runtimeName))
	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)
	runtimeOAuthGraphQLClient := gqlClient(t, rtmOauthCredentialData, token.RuntimeScopes)

	// REGISTER APPLICATION
	t.Log(fmt.Sprintf("Registering application %q", appName))
	appInput := appWithAPIsAndEvents(appName)
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantId, appInput)
	require.NoError(t, err)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)

	// assert document, event and api definitions are present
	require.Len(t, app.Bundles.Data, 1)
	bndl := app.Bundles.Data[0]

	require.Len(t, bndl.EventDefinitions.Data, 1)
	require.Len(t, bndl.APIDefinitions.Data, 1)
	require.Len(t, bndl.Documents.Data, 1)

	// register application oauth client
	t.Log(fmt.Sprintf("Requesting application OAuth client for application %q", appName))
	appAuth := fixtures.RequestClientCredentialsForApplication(t, context.Background(), dexGraphQLClient, tenantId, app.ID)
	appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, appOauthCredentialData.ClientSecret)
	require.NotEmpty(t, appOauthCredentialData.ClientID)
	applicationOAuthGraphQLClient := gqlClient(t, appOauthCredentialData, token.ApplicationScopes)

	// register integration system
	t.Log(fmt.Sprintf("Registering integration system %q", intSysName))
	integrationSystem := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSysName)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, integrationSystem.ID)

	// register integration system oauth client
	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, context.Background(), dexGraphQLClient, tenantId, integrationSystem.ID)
	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)
	intSystemOAuthGraphQLClient := gqlClient(t, intSysOauthCredentialData, token.IntegrationSystemScopes)

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
	instanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)
	require.NotNil(t, instanceAuth)

	type accessRequired struct {
		appWebhooks               bool
		appAuths                  bool
		appTemplateWebhooks       bool
		bundleInstanceAuth        bool
		bundleInstanceAuths       bool
		bundleDefaultInstanceAuth bool
		documentFetchRequest      bool
		eventSpecFetchRequest     bool
		apiSpecFetchRequest       bool
		integrationSystemAuths    bool
		runtimeAuths              bool
	}

	type consumerType string
	const (
		applicationConsumer       consumerType = "application"
		runtimeConsumer           consumerType = "runtime"
		integrationSystemConsumer consumerType = "integrationSystem"
	)

	type testCase struct {
		name              string
		consumer          *gcli.Client
		consumerType      consumerType
		fieldExpectations accessRequired
	}
	testCases := []testCase{
		{
			name:         "Runtime access",
			consumer:     runtimeOAuthGraphQLClient,
			consumerType: runtimeConsumer,
			fieldExpectations: accessRequired{
				bundleInstanceAuth:        true,
				bundleInstanceAuths:       true,
				bundleDefaultInstanceAuth: true,
				runtimeAuths:              true,
			},
		},
		{
			name:         "Integration system access",
			consumer:     intSystemOAuthGraphQLClient,
			consumerType: integrationSystemConsumer,
			fieldExpectations: accessRequired{
				appTemplateWebhooks:    true,
				integrationSystemAuths: true,
			},
		},
		{
			name:         "Application access",
			consumer:     applicationOAuthGraphQLClient,
			consumerType: applicationConsumer,
			fieldExpectations: accessRequired{
				appWebhooks:               true,
				appAuths:                  true,
				bundleInstanceAuth:        true,
				bundleInstanceAuths:       true,
				bundleDefaultInstanceAuth: true,
				documentFetchRequest:      true,
				eventSpecFetchRequest:     true,
				apiSpecFetchRequest:       true,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			// applications
			application := fixtures.GetApplication(t, ctx, test.consumer, tenantId, app.ID)
			require.Equal(t, application.Webhooks != nil, test.fieldExpectations.appWebhooks)
			require.Equal(t, application.Auths != nil, test.fieldExpectations.appAuths)
			//=== RUN   TestSensitiveDataStrip/Application_access
			// X=false true
			require.Equal(t, application.Bundles.Data[0].APIDefinitions.Data[0].Spec.FetchRequest != nil, test.fieldExpectations.apiSpecFetchRequest)
			require.Equal(t, application.Bundles.Data[0].EventDefinitions.Data[0].Spec.FetchRequest != nil, test.fieldExpectations.eventSpecFetchRequest)
			require.Equal(t, application.Bundles.Data[0].Documents.Data[0].FetchRequest != nil, test.fieldExpectations.documentFetchRequest)

			require.Equal(t, application.Bundles.Data[0].InstanceAuths != nil, test.fieldExpectations.bundleInstanceAuths)

			// TODO app templates
			if test.consumerType == integrationSystemConsumer {
				// integration systems
				is := fixtures.GetIntegrationSystem(t, ctx, test.consumer, integrationSystem.ID)
				require.Equal(t, is.Auths != nil, test.fieldExpectations.integrationSystemAuths)

				appTemplate := fixtures.GetApplicationTemplate(t, ctx, test.consumer, tenantId, appTemplate.ID)
				t.Log(fmt.Sprintf("APP TEMPLATE: %v", appTemplate.Webhooks))
				require.Equal(t, appTemplate.Webhooks != nil, test.fieldExpectations.appTemplateWebhooks)
			}

			if test.consumerType == integrationSystemConsumer || test.consumerType == runtimeConsumer {
				// runtimes
				rt := fixtures.GetRuntime(t, ctx, test.consumer, tenantId, runtime.ID)
				require.Equal(t, rt.Auths != nil, test.fieldExpectations.runtimeAuths)
			}
		})
	}
}

func gqlClient(t *testing.T, creds *graphql.OAuthCredentialData, scopes string) *gcli.Client {
	accessToken := token.GetAccessToken(t, creds, scopes)
	return gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)
}

func appWithAPIsAndEvents(name string) graphql.ApplicationRegisterInput {
	webhookURL := "http://test-url.com"
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
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.WebhookTypeUnregisterApplication,
			URL:  &webhookURL,
		}},
	}
}
