package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/stretchr/testify/require"
)

type RequestBody struct {
	State         ConfigurationState `json:"state"`
	Configuration json.RawMessage    `json:"configuration"`
	Error         string             `json:"error"`
}

type ConfigurationState string

const (
	ReadyConfigurationState         ConfigurationState = "READY"
	CreateErrorConfigurationState   ConfigurationState = "CREATE_ERROR"
	DeleteErrorConfigurationState   ConfigurationState = "DELETE_ERROR"
	ConfigPendingConfigurationState ConfigurationState = "CONFIG_PENDING"
)

func Test_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	parentTenantID := tenant.TestTenants.GetDefaultTenantID()
	subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultSubaccount)

	intSysName := "async-formation-int-system"
	testConfig := `{"testCfgKey":"testCfgValue"}`
	certAuthorizedClient := gql.NewCertAuthorizedHTTPClient(cc.Get()[conf.ExternalClientCertSecretName].PrivateKey, cc.Get()[conf.ExternalClientCertSecretName].Certificate, conf.SkipSSLValidation)

	t.Run("Integration system caller successfully updates formation assignment with target type application", func(t *testing.T) {
		t.Logf("Creating integration system with name: %q", intSysName)
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSysName)
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
		require.NotEmpty(t, intSysOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		asyncFormationTmplName := "async-formation-template-name"
		t.Logf("Creating formation template with name: %q", asyncFormationTmplName)
		ft := createFormationTemplate(t, ctx, "async", asyncFormationTmplName, conf.SubscriptionProviderAppNameValue, graphql.ArtifactTypeEnvironmentInstance)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		asyncFormationName := "async-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", asyncFormationName, asyncFormationTmplName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName, &asyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName)
		formationID := formation.ID
		require.NotEmpty(t, formationID)

		applicationType := "async-app-type-1"
		appRegion := "test-async-app-region"
		appNamespace := "compass.async.test"
		localTenantID := "local-async-tenant-id"
		t.Logf("Create application template for type %q", applicationType)
		appTemplateInput := graphql.ApplicationTemplateInput{
			Name:        applicationType,
			Description: &applicationType,
			ApplicationInput: &graphql.ApplicationRegisterInput{
				Name:          "{{name}}",
				ProviderName:  str.Ptr("compassAsyncTest"),
				Description:   ptr.String("test {{display-name}}"),
				LocalTenantID: &localTenantID,
				Labels: graphql.Labels{
					"applicationType": applicationType,
					"region":          appRegion,
				},
			},
			Placeholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name: "name",
				},
				{
					Name: "display-name",
				},
			},
			ApplicationNamespace: &appNamespace,
			AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
		}
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := graphql.ApplicationFromTemplateInput{
			TemplateName: applicationType, Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "async-app-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "Async App",
				},
			},
		}

		t.Logf("Create application from template: %q", applicationType)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmpl := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		asyncApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, createAppFromTmpl, &asyncApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, parentTenantID, &asyncApp)
		require.NoError(t, err)
		require.NotEmpty(t, asyncApp.ID)
		t.Logf("Async app ID: %q", asyncApp.ID)

		t.Logf("Assign application with name: %q to formation %q", asyncApp.Name, asyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(asyncApp.ID, string(graphql.FormationObjectTypeApplication), asyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)

		// todo::: do we need the ASA here?
		t.Logf("Assign tenant: %q to formation: %q", parentTenantID, asyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), asyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subaccountID, parentTenantID)

		// todo::: get FA + ID
		t.Logf("List formations for ID: %q and their formation assignments", formationID)
		listFormationsReq := fixtures.FixListFormationsRequestWithPageSize(100)
		formationPage := fixtures.ListFormations(t, ctx, certSecuredGraphQLClient, listFormationsReq, 1)
		require.Empty(t, formationPage.Data)

		reqBody := RequestBody{
			State:         ReadyConfigurationState,
			Configuration: json.RawMessage(testConfig),
		}
		marshalBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := certAuthorizedClient.Post(conf.DirectorExternalCertFormationMappingURL, util.ContentTypeApplicationJSON, bytes.NewBuffer(marshalBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

	})

	t.Run("Runtime caller successfully updates formation assignment with target type runtime", func(t *testing.T) {
		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "selfRegisterRuntimeAsync",
			Description: ptr.String("selfRegisterRuntimeAsync-description"),
		}

		t.Log("Register runtime via certificate secured client")
		runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, parentTenantID, runtimeInput, conf.GatewayOauth)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, parentTenantID, &runtime)

		// Register application
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "testingApp", conf.ApplicationTypeLabelKey, string(util.ApplicationTypeC4C), parentTenantID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, parentTenantID, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		asyncFormationTmplName := "async-formation-template-name"
		t.Logf("Creating formation template with name: %q", asyncFormationTmplName)
		ft := createFormationTemplate(t, ctx, "async", asyncFormationTmplName, conf.SubscriptionProviderAppNameValue, graphql.ArtifactTypeEnvironmentInstance)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		asyncFormationName := "async-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", asyncFormationName, asyncFormationTmplName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName, &asyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName)
		formationID := formation.ID
		require.NotEmpty(t, formationID)

		t.Logf("Assign application with name: %q to formation %q", app.Name, asyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), asyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)

		t.Logf("Assign tenant: %q to formation: %q", parentTenantID, asyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), asyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subaccountID, parentTenantID)

		reqBody := RequestBody{
			State:         ReadyConfigurationState,
			Configuration: json.RawMessage(testConfig),
		}
		marshalBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := certAuthorizedClient.Post(conf.DirectorExternalCertFormationMappingURL, util.ContentTypeApplicationJSON, bytes.NewBuffer(marshalBody))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Runtime caller successfully updates formation assignment with target type runtime context", func(t *testing.T) {
		// Prepare provider external client certificate and secret, and build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		//certSecuredClient := gql.NewCertAuthorizedHTTPClient(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntimeAsync",
			Description: ptr.String("providerRuntimeAsync-description"),
			Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
		}

		providerRuntime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		regionLbl, ok := providerRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)
	})

	t.Run("Runtime caller successfully updates formation assignment with target type application", func(t *testing.T) {

	})

	t.Run("Unauthorized call", func(t *testing.T) {

	})
}
