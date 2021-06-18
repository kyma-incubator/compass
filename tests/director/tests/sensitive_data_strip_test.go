package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
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

	t.Log("Creating application template")
	appTmpInput := fixtures.FixApplicationTemplateWithWebhook("app-template-test")
	appTemplate := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, appTmpInput)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)

	t.Log(fmt.Sprintf("Registering runtime %q", runtimeName))
	runtimeRegInput := fixtures.FixRuntimeInput(runtimeName)
	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &runtimeRegInput)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	t.Log(fmt.Sprintf("Requesting OAuth client for runtime %q", runtimeName))
	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)
	runtimeOAuthGraphQLClient := gqlClient(t, rtmOauthCredentialData, token.RuntimeScopes)

	t.Log(fmt.Sprintf("Registering application %q", appName))
	appInput := appWithAPIsAndEvents(appName)
	app, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, tenantId, appInput)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, app.ID)
	require.NoError(t, err)

	t.Log(fmt.Sprintf("Asserting document, event and api definitions are present"))
	require.Len(t, app.Bundles.Data, 1)
	bndl := app.Bundles.Data[0]

	require.Len(t, bndl.EventDefinitions.Data, 1)
	require.Len(t, bndl.APIDefinitions.Data, 1)
	require.Len(t, bndl.Documents.Data, 1)

	t.Log(fmt.Sprintf("Requesting application OAuth client for application %q", appName))
	appAuth := fixtures.RequestClientCredentialsForApplication(t, context.Background(), dexGraphQLClient, tenantId, app.ID)
	appOauthCredentialData, ok := appAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, appOauthCredentialData.ClientSecret)
	require.NotEmpty(t, appOauthCredentialData.ClientID)
	applicationOAuthGraphQLClient := gqlClient(t, appOauthCredentialData, token.ApplicationScopes)

	t.Log(fmt.Sprintf("Registering integration system %q", intSysName))
	integrationSystem := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSysName)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, integrationSystem.ID)

	t.Log(fmt.Sprintf("Registering OAuth client for integration system %q", intSysName))
	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, context.Background(), dexGraphQLClient, tenantId, integrationSystem.ID)
	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)
	intSystemOAuthGraphQLClient := gqlClient(t, intSysOauthCredentialData, token.IntegrationSystemScopes)

	t.Log(fmt.Sprintf("assign runtime and app to scenario: %s", "'test-scenario'"))
	scenarios := []string{conf.DefaultScenario, "test-scenario"}
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios[:1])
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenantId, scenarios)

	t.Log(fmt.Sprintf("Setting application scenarios label: %s", ScenariosLabel))
	defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, ScenariosLabel, scenarios[:1])
	fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, app.ID, ScenariosLabel, scenarios[1:])

	t.Log(fmt.Sprintf("Setting runtime scenarios label: %s", ScenariosLabel))
	defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])
	fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])

	t.Log(fmt.Sprintf("Creating bundle instance auths %q with bundle with APIDefinition and ", appName))
	instanceAuth := fixtures.CreateBundleInstanceAuth(t, ctx, dexGraphQLClient, bndl.ID)
	require.NotNil(t, instanceAuth)

	t.Run("Application access", func(t *testing.T) {
		type accessRequired struct {
			appWebhooks               bool
			appWebhooksAuth           bool
			appAuths                  bool
			bundleInstanceAuth        bool
			bundleInstanceAuths       bool
			bundleDefaultInstanceAuth bool
			documentFetchRequest      bool
			eventSpecFetchRequest     bool
			apiSpecFetchRequest       bool
			fetchRequestAuth          bool
		}
		type testCase struct {
			name              string
			consumer          *gcli.Client
			fieldExpectations accessRequired
		}
		testCases := []testCase{
			{
				name:     "Runtime consumer",
				consumer: runtimeOAuthGraphQLClient,
				fieldExpectations: accessRequired{
					bundleInstanceAuth:        true,
					bundleInstanceAuths:       true,
					bundleDefaultInstanceAuth: true,
				},
			},
			{
				name:              "Integration system consumer",
				consumer:          intSystemOAuthGraphQLClient,
				fieldExpectations: accessRequired{},
			},
			{
				name:     "Application consumer",
				consumer: applicationOAuthGraphQLClient,
				fieldExpectations: accessRequired{
					appWebhooks:               true,
					appAuths:                  true,
					bundleInstanceAuth:        true,
					bundleInstanceAuths:       true,
					bundleDefaultInstanceAuth: true,
					documentFetchRequest:      true,
					eventSpecFetchRequest:     true,
					apiSpecFetchRequest:       true,
					fetchRequestAuth:          true,
				},
			},
			{
				name:     "Admin user consumer",
				consumer: dexGraphQLClient,
				fieldExpectations: accessRequired{
					appWebhooks:               true,
					appAuths:                  true,
					bundleInstanceAuth:        true,
					bundleInstanceAuths:       true,
					bundleDefaultInstanceAuth: true,
					documentFetchRequest:      true,
					eventSpecFetchRequest:     true,
					apiSpecFetchRequest:       true,
					appWebhooksAuth:           true,
					fetchRequestAuth:          true,
				},
			},
		}

		for _, test := range testCases {
			t.Run(test.name, func(t *testing.T) {
				application := fixtures.GetApplication(t, ctx, test.consumer, tenantId, app.ID)

				require.Greater(t, len(application.Auths), 0)
				for _, applicationAuth := range application.Auths {
					require.NotEmpty(t, applicationAuth.ID)
					require.Equal(t, applicationAuth.Auth != nil, test.fieldExpectations.appAuths)

				}
				require.Equal(t, application.Bundles.Data[0].InstanceAuths != nil, test.fieldExpectations.bundleInstanceAuths)

				require.Equal(t, application.Webhooks != nil, test.fieldExpectations.appWebhooks)
				if application.Webhooks != nil {
					require.Equal(t, application.Webhooks[0].Auth != nil, test.fieldExpectations.appWebhooksAuth)
				}

				require.Equal(t, application.Bundles.Data[0].APIDefinitions.Data[0].Spec.FetchRequest != nil, test.fieldExpectations.apiSpecFetchRequest)
				if application.Bundles.Data[0].APIDefinitions.Data[0].Spec.FetchRequest != nil {
					require.Equal(t, application.Bundles.Data[0].APIDefinitions.Data[0].Spec.FetchRequest.Auth != nil, test.fieldExpectations.fetchRequestAuth)
				}

				require.Equal(t, application.Bundles.Data[0].EventDefinitions.Data[0].Spec.FetchRequest != nil, test.fieldExpectations.eventSpecFetchRequest)
				if application.Bundles.Data[0].EventDefinitions.Data[0].Spec.FetchRequest != nil {
					require.Equal(t, application.Bundles.Data[0].EventDefinitions.Data[0].Spec.FetchRequest.Auth != nil, test.fieldExpectations.fetchRequestAuth)
				}

				require.Equal(t, application.Bundles.Data[0].Documents.Data[0].FetchRequest != nil, test.fieldExpectations.documentFetchRequest)
				if application.Bundles.Data[0].Documents.Data[0].FetchRequest != nil {
					require.Equal(t, application.Bundles.Data[0].Documents.Data[0].FetchRequest.Auth != nil, test.fieldExpectations.fetchRequestAuth)
				}
			})
		}
	})

	t.Run("Application template access", func(t *testing.T) {
		t.Run("from integration system", func(t *testing.T) {
			appTemplate := fixtures.GetApplicationTemplate(t, ctx, intSystemOAuthGraphQLClient, tenantId, appTemplate.ID)
			require.NotNil(t, appTemplate.Webhooks, "app template webhooks should be visible")
			require.Nil(t, appTemplate.Webhooks[0].Auth, "app template webhook auths should not be visible")
		})
		t.Run("from admin user", func(t *testing.T) {
			appTemplate := fixtures.GetApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)
			require.NotNil(t, appTemplate.Webhooks, "app template webhooks should be visible")
			require.NotNil(t, appTemplate.Webhooks[0].Auth, "app template webhook auths should be visible")
		})
	})

	t.Run("Runtime access", func(t *testing.T) {
		t.Run("from runtime", func(t *testing.T) {
			rt := fixtures.GetRuntime(t, ctx, runtimeOAuthGraphQLClient, tenantId, runtime.ID)
			require.NotNil(t, rt.Auths)
		})
		t.Run("from admin user", func(t *testing.T) {
			rt := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)
			require.NotNil(t, rt.Auths)
		})
	})

	t.Run("Integration system access", func(t *testing.T) {
		t.Run("from integration system", func(t *testing.T) {
			is := fixtures.GetIntegrationSystem(t, ctx, intSystemOAuthGraphQLClient, integrationSystem.ID)
			require.NotNil(t, is.Auths)
		})
		t.Run("from admin user", func(t *testing.T) {
			is := fixtures.GetIntegrationSystem(t, ctx, dexGraphQLClient, integrationSystem.ID)
			require.NotNil(t, is.Auths)
		})
	})
}

func gqlClient(t *testing.T, creds *graphql.OAuthCredentialData, scopes string) *gcli.Client {
	accessToken := token.GetAccessToken(t, creds, scopes)
	return gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)
}

func appWithAPIsAndEvents(name string) graphql.ApplicationRegisterInput {
	webhookURL := "http://test-url.com"
	webhookOutputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
	auth := &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "username",
				Password: "password",
			},
		},
	}
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
						URL:  OpenAPISpec,
						Auth: auth,
					},
				},
			}},
			EventDefinitions: []*graphql.EventDefinitionInput{{
				Name: "test-event-def",
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL:  AsyncAPISpec,
						Auth: auth,
					},
				},
			}},
			Documents: []*graphql.DocumentInput{{
				Title:  "test-document",
				Format: graphql.DocumentFormatMarkdown,
				FetchRequest: &graphql.FetchRequestInput{
					URL:  MDDocumentURL,
					Auth: auth,
				},
			}},
		}},
		Webhooks: []*graphql.WebhookInput{{
			Type:           graphql.WebhookTypeUnregisterApplication,
			URL:            &webhookURL,
			Auth:           auth,
			OutputTemplate: &webhookOutputTemplate,
		}},
	}
}
