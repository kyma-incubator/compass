package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

type RequestBody struct {
	State         string          `json:"state"`
	Configuration json.RawMessage `json:"configuration"`
	Error         string          `json:"error"`
}

const (
	formationIDPathParam           = "ucl-formation-id"
	formationAssignmentIDPathParam = "ucl-assignment-id"
)

func Test_UpdateStatus(baseT *testing.T) {
	t := testingx.NewT(baseT)
	ctx := context.Background()

	intSysName := "async-formation-int-system"
	testConfig := `{"testCfgKey":"testCfgValue"}`

	t.Run("Caller successfully updates formation assignment for himself", func(t *testing.T) {
		// Prepare provider external client certificate and secret, and build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
		certSecuredHTTPClient := gql.NewCertAuthorizedHTTPClient(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		parentTenantID := conf.TestProviderAccountID
		subaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant

		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "selfRegisterRuntimeAsync",
			Description: ptr.String("selfRegisterRuntimeAsync-description"),
		}

		t.Log("Register runtime via certificate secured client")
		runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, runtimeInput, conf.GatewayOauth)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, &runtime)

		appName := "testAsyncApp"
		appType := "async-app-type-1"
		t.Logf("Register application with name: %q", appName)
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, appName, conf.ApplicationTypeLabelKey, appType, subaccountID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subaccountID, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		asyncFormationTmplName := "async-formation-template-name"
		ft := createFormationTemplate(t, ctx, asyncFormationTmplName, conf.KymaRuntimeTypeLabelValue, []string{appType}, graphql.ArtifactTypeEnvironmentInstance)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		asyncFormationName := "async-formation-name"
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName, &asyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName)
		formationID := formation.ID
		require.NotEmpty(t, formationID)

		t.Logf("Assign application with name: %q to formation %q", app.Name, asyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), asyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: asyncFormationName}, app.ID, graphql.FormationObjectTypeApplication, parentTenantID)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)

		t.Logf("Assign tenant: %q to formation: %q", parentTenantID, asyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), asyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subaccountID, parentTenantID)

		t.Logf("Listing formation assignments for formation with ID: %q", formationID)
		listFormationAssignmentsReq := fixtures.FixListFormationAssignmentRequest(formationID, 100)
		assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, parentTenantID, listFormationAssignmentsReq)
		require.Len(t, assignmentsPage.Data, 2)
		require.Equal(t, 2, assignmentsPage.TotalCount)

		t.Run("Runtime caller successfully updates his formation assignment", func(t *testing.T) {
			formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeRuntime, app.ID)
			executeStatusUpdateReqWithExpectedStatusCode(t, certSecuredHTTPClient, testConfig, formationID, formationAssignmentID, http.StatusOK)
		})

		t.Run("Application caller successfully updates his formation assignment", func(t *testing.T) {
			formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeApplication, runtime.ID)
			executeStatusUpdateReqWithExpectedStatusCode(t, certSecuredHTTPClient, testConfig, formationID, formationAssignmentID, http.StatusOK)
		})
	})

	t.Run("Runtime caller successfully updates formation assignment with target type rtm context", func(t *testing.T) {
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// Prepare provider external client certificate and secret, and build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
		directorCertSecuredGQLClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		// the package is gql but the client that is build is a "standard" http client, not a GraphQL one
		certSecuredHTTPClient := gql.NewCertAuthorizedHTTPClient(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "providerRuntimeAsync",
			Description: ptr.String("providerRuntimeAsync-description"),
			Labels:      graphql.Labels{conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue},
		}

		providerRuntime := fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredGQLClient, &providerRuntimeInput)
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredGQLClient, &providerRuntime)
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

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, response.StatusCode)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
		require.NoError(t, err)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
		subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
		subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

		t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
		resp, err := httpClient.Do(subscribeReq)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		require.NoError(t, err)
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

		subJobStatusPath := resp.Header.Get(subscription.LocationHeader)
		require.NotEmpty(t, subJobStatusPath)
		subJobStatusURL := conf.SubscriptionConfig.URL + subJobStatusPath
		require.Eventually(t, func() bool {
			return subscription.GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == subscription.JobSucceededStatus
		}, subscription.EventuallyTimeout, subscription.EventuallyTick)
		t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)

		rtmRequest := fixtures.FixGetRuntimeContextsRequest(providerRuntime.ID)
		rtm := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, rtmRequest, &rtm)
		require.NoError(t, err)
		require.Len(t, rtm.RuntimeContexts.Data, 1)

		applicationType := "provider-async-app-type-1"
		providerAsyncFormationTmplName := "provider-async-formation-template-name"
		ft := createFormationTemplate(t, ctx, providerAsyncFormationTmplName, conf.SubscriptionProviderAppNameValue, []string{applicationType}, graphql.ArtifactTypeSubscription)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		providerAsyncFormationName := "provider-async-formation-name"
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerAsyncFormationName, &providerAsyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerAsyncFormationName)
		require.NotEmpty(t, formation.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSysName)
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, "", intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		appRegion := "test-async-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-async-tenant-id"
		t.Logf("Create application template for type: %q", applicationType)
		appTemplateInput := graphql.ApplicationTemplateInput{
			Name:        applicationType,
			Description: &applicationType,
			ApplicationInput: &graphql.ApplicationRegisterInput{
				Name:          "{{name}}",
				ProviderName:  str.Ptr("compass"),
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
			TemplateName: applicationType,
			Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       "async-app-formation-mapping-tests",
				},
				{
					Placeholder: "display-name",
					Value:       "async-app-display-name",
				},
			},
		}

		t.Logf("Create application from template: %q", applicationType)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		asyncApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &asyncApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &asyncApp)
		require.NoError(t, err)
		require.NotEmpty(t, asyncApp.ID)
		t.Logf("App ID: %q", asyncApp.ID)

		t.Logf("Assign application to formation: %q", providerAsyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(asyncApp.ID, string(graphql.FormationObjectTypeApplication), providerAsyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerAsyncFormationName}, asyncApp.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerAsyncFormationName, assignedFormation.Name)

		assertFormationAssignmentsCount(t, ctx, formation.ID, subscriptionConsumerAccountID, 0)

		t.Logf("Assign tenant: %q to formation: %q", subscriptionConsumerSubaccountID, providerAsyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerAsyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerAsyncFormationName, assignedFormation.Name)

		assignmentsPage := assertFormationAssignmentsCount(t, ctx, formation.ID, subscriptionConsumerAccountID, 2)

		formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeRuntimeContext, asyncApp.ID)
		executeStatusUpdateReqWithExpectedStatusCode(t, certSecuredHTTPClient, testConfig, formation.ID, formationAssignmentID, http.StatusOK)
	})

	t.Run("Application caller successfully updates formation assignment with target type application made through subscription", func(t *testing.T) {
		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
		subscriptionConsumerAccountID := conf.TestConsumerAccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// We need an externally issued cert with a subject that is not part of the access level mappings
		externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
			ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
			ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
			CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
			ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
			ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
			TestExternalCertSubject:               strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "app-template-subscription-cn", -1),
			ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
			ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
			ExternalCertProvider:                  certprovider.CertificateService,
		}

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
		appProviderDirectorCertSecuredGQLClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		appProviderCertSecuredHTTPClient := gql.NewCertAuthorizedHTTPClient(providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)

		// Create Application Template
		appTemplateName := createAppTemplateName("app-template-name-subscription-async")
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhooks(appTemplateName)
		appTemplateInput.Labels["applicationType"] = appTemplateName
		appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorCertSecuredGQLClient, subscriptionConsumerAccountID, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, appProviderDirectorCertSecuredGQLClient, subscriptionConsumerAccountID, appTmpl)
		require.NoError(t, err)
		require.NotEmpty(t, appTmpl.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

		selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

		regionLbl, ok := appTmpl.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, regionLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer([]byte(selfRegLabelValue)))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		require.NoError(t, err)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.Equal(t, http.StatusOK, response.StatusCode)

		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)
		createSubscription(t, ctx, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID)

		actualAppPage := graphql.ApplicationPage{}
		getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
		require.NoError(t, err)

		require.Len(t, actualAppPage.Data, 1)
		require.Equal(t, appTmpl.ID, *actualAppPage.Data[0].ApplicationTemplateID)

		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "runtimeAsync",
			Description: ptr.String("runtimeAsync-description"),
		}

		t.Log("Register runtime via certificate secured client")
		runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, runtimeInput, conf.GatewayOauth)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, &runtime)

		providerAsyncFormationTmplName := "provider-async-formation-template-name"
		providerAppTypes := []string{appTemplateName}
		ft := createFormationTemplate(t, ctx, providerAsyncFormationTmplName, conf.KymaRuntimeTypeLabelValue, providerAppTypes, graphql.ArtifactTypeEnvironmentInstance)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		providerAsyncFormationName := "provider-async-formation-name"
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerAsyncFormationName, &providerAsyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, providerAsyncFormationName)
		require.NotEmpty(t, formation.ID)

		appID := actualAppPage.Data[0].ID
		t.Logf("Assign application to formation: %q", providerAsyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(appID, string(graphql.FormationObjectTypeApplication), providerAsyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: providerAsyncFormationName}, appID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerAsyncFormationName, assignedFormation.Name)

		assertFormationAssignmentsCount(t, ctx, formation.ID, subscriptionConsumerAccountID, 0)

		t.Logf("Assign tenant: %q to formation: %q", subscriptionConsumerSubaccountID, providerAsyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), providerAsyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
		require.NoError(t, err)
		require.Equal(t, providerAsyncFormationName, assignedFormation.Name)

		assignmentsPage := assertFormationAssignmentsCount(t, ctx, formation.ID, subscriptionConsumerAccountID, 2)
		formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeApplication, runtime.ID)

		executeStatusUpdateReqWithExpectedStatusCode(t, appProviderCertSecuredHTTPClient, testConfig, formation.ID, formationAssignmentID, http.StatusOK)
	})

	t.Run("Unauthorized call", func(t *testing.T) {
		parentTenantID := tenant.TestTenants.GetDefaultTenantID()
		subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)

		runtimeInput := graphql.RuntimeRegisterInput{
			Name:        "selfRegisterRuntimeAsync",
			Description: ptr.String("selfRegisterRuntimeAsync-description"),
		}

		t.Log("Register runtime via certificate secured client")
		runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccountID, runtimeInput, conf.GatewayOauth)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, parentTenantID, &runtime)

		appName := "testAsyncApp"
		appType := "async-app-type-1"
		t.Logf("Register application with name: %q", appName)
		app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, appName, conf.ApplicationTypeLabelKey, appType, subaccountID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subaccountID, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)

		asyncFormationTmplName := "async-formation-template-name"
		ft := createFormationTemplate(t, ctx, asyncFormationTmplName, conf.KymaRuntimeTypeLabelValue, []string{appType}, graphql.ArtifactTypeEnvironmentInstance)
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, ft.ID)

		asyncFormationName := "async-formation-name"
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName, &asyncFormationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, parentTenantID, asyncFormationName)
		formationID := formation.ID
		require.NotEmpty(t, formationID)

		t.Logf("Assign application with name: %q to formation %q", app.Name, asyncFormationName)
		assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), asyncFormationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: asyncFormationName}, app.ID, graphql.FormationObjectTypeApplication, parentTenantID)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)

		t.Logf("Assign tenant: %q to formation: %q", parentTenantID, asyncFormationName)
		assignReq = fixtures.FixAssignFormationRequest(subaccountID, string(graphql.FormationObjectTypeTenant), asyncFormationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, parentTenantID, assignReq, &assignedFormation)
		defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subaccountID, parentTenantID)
		require.NoError(t, err)
		require.Equal(t, asyncFormationName, assignedFormation.Name)

		t.Logf("List formation assignments for formation with ID: %q", formationID)
		listFormationAssignmentsReq := fixtures.FixListFormationAssignmentRequest(formationID, 100)
		assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, parentTenantID, listFormationAssignmentsReq)
		require.Len(t, assignmentsPage.Data, 2)
		require.Equal(t, 2, assignmentsPage.TotalCount)
		formationAssignmentID := getFormationAssignmentIDByTargetTypeAndSourceID(t, assignmentsPage, graphql.FormationAssignmentTypeRuntime, app.ID)
		t.Logf("successfully listed FAs for formation ID: %q", formationID)

		certSecuredHTTPClientWithDifferentTenant := getHTTPCertClientWithCustomSubject(t, ctx, conf, conf.ExternalCertTestIntSystemOUSubaccount, conf.ExternalCertTestIntSystemCommonName)

		executeStatusUpdateReqWithExpectedStatusCode(t, certSecuredHTTPClientWithDifferentTenant, testConfig, formation.ID, formationAssignmentID, http.StatusUnauthorized)
	})
}

func getHTTPCertClientWithCustomSubject(t *testing.T, ctx context.Context, conf *DirectorConfig, certTenant, commonName string) *http.Client {
	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, certTenant, conf.ExternalCertCommonName, commonName)

	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestIntSystemSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               replacer.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	return gql.NewCertAuthorizedHTTPClient(pk, cert, conf.SkipSSLValidation)
}

func executeStatusUpdateReqWithExpectedStatusCode(t *testing.T, certSecuredHTTPClient *http.Client, testConfig, formationID, formationAssignmentID string, expectedStatusCode int) {
	reqBody := RequestBody{
		State:         "READY",
		Configuration: json.RawMessage(testConfig),
	}
	marshalBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	formationMappingEndpoint := resolveFormationMappingURL(formationID, formationAssignmentID)
	request, err := http.NewRequest(http.MethodPatch, formationMappingEndpoint, bytes.NewBuffer(marshalBody))
	require.NoError(t, err)
	request.Header.Add("Content-Type", "application/json")
	response, err := certSecuredHTTPClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
}

func assertFormationAssignmentsCount(t *testing.T, ctx context.Context, formationID, parentTenantID string, expectedAssignmentsCount int) *graphql.FormationAssignmentPage {
	t.Logf("List formation assignments for formation with ID: %q", formationID)
	listFormationAssignmentsReq := fixtures.FixListFormationAssignmentRequest(formationID, 100)
	assignmentsPage := fixtures.ListFormationAssignments(t, ctx, certSecuredGraphQLClient, parentTenantID, listFormationAssignmentsReq)
	require.Len(t, assignmentsPage.Data, expectedAssignmentsCount)
	require.Equal(t, expectedAssignmentsCount, assignmentsPage.TotalCount)
	return assignmentsPage
}

func getFormationAssignmentIDByTargetTypeAndSourceID(t *testing.T, assignmentsPage *graphql.FormationAssignmentPage, targetType graphql.FormationAssignmentType, sourceID string) string {
	var formationAssignmentID string
	for _, a := range assignmentsPage.Data {
		if a.TargetType == targetType && a.Source == sourceID {
			formationAssignmentID = a.ID
		}
	}
	require.NotEmpty(t, formationAssignmentID, "The formation assignment could not be empty")
	return formationAssignmentID
}

func resolveFormationMappingURL(formationID, formationAssignmentID string) string {
	formationMappingEndpoint := strings.Replace(conf.DirectorExternalCertFormationMappingURL, fmt.Sprintf("{%s}", formationIDPathParam), formationID, 1)
	formationMappingEndpoint = strings.Replace(formationMappingEndpoint, fmt.Sprintf("{%s}", formationAssignmentIDPathParam), formationAssignmentID, 1)
	return formationMappingEndpoint
}
