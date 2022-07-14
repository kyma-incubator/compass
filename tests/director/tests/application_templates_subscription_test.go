package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionApplicationTemplateFlow(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("When creating app template with a certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		subscriptionProviderSubaccountID := conf.TestProviderSubaccountID
		subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID
		subscriptionConsumerTenantID := conf.TestConsumerTenantID

		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")

		// Create Application Template
		appTemplateName := createAppTemplateName("app-template-name-subscription")
		appTemplateInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), &appTmpl)
		require.NoError(t, err)
		require.NotEmpty(t, appTmpl.ID)

		selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

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

		t.Run("Subscribe tenant to Application flow", func(t *testing.T) {
			// WHEN
			createSubscription(t, ctx, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			// THEN
			actualAppPage := graphql.ApplicationPage{}
			getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 1)
			require.Equal(t, appTmpl.ID, *actualAppPage.Data[0].ApplicationTemplateID)
		})

		t.Run("Subscribe tenant to Application flow and add bundle", func(t *testing.T) {
			// Subscribe
			createSubscription(t, ctx, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID)
			defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			// Ensure subscription is OK
			actualAppPage := graphql.ApplicationPage{}
			getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 1)
			subscribedApplication := actualAppPage.Data[0]
			require.Equal(t, appTmpl.ID, *subscribedApplication.ApplicationTemplateID)

			//Create Bundle
			bndlInput := fixtures.FixBundleCreateInputWithRelatedObjects(t, "bndl-app-1")
			bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
			require.NoError(t, err)
			addBndlRequest := fixtures.FixAddBundleRequest(subscribedApplication.ID, bndl)
			output := graphql.BundleExt{}

			t.Log("Try to create bundle")
			err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &output)

			//Verify that Bundle can be created
			require.NoError(t, err)
			require.NotEmpty(t, output.ID)
			assertions.AssertBundle(t, &bndlInput, &output)
			defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, subscriptionConsumerTenantID, output.ID)
		})

		t.Run("Unsubscribe tenant to Application flow and ensure that canot add bundle", func(t *testing.T) {
			// Subscribe
			createSubscription(t, ctx, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID)

			// Ensure subscription is OK
			actualAppPage := graphql.ApplicationPage{}
			getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 1)
			subscribedApplication := actualAppPage.Data[0]
			require.Equal(t, appTmpl.ID, *subscribedApplication.ApplicationTemplateID)

			// Unsubscribe
			subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			// Ensure Unsubscription is OK
			actualAppPage = graphql.ApplicationPage{}
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 0)

			//Create Bundle
			bndlInput := fixtures.FixBundleCreateInputWithRelatedObjects(t, "bndl-app-1")
			bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
			require.NoError(t, err)
			addBndlRequest := fixtures.FixAddBundleRequest(subscribedApplication.ID, bndl)
			output := graphql.BundleExt{}

			t.Log("Try to create bundle")
			err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &output)

			//Verify that Bundle cannot be created after unsubscribtion
			require.Error(t, err)
		})

		t.Run("Unsubscribe tenant to Application flow", func(t *testing.T) {
			// GIVEN
			createSubscription(t, ctx, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID)

			actualAppPage := graphql.ApplicationPage{}
			getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 1)
			require.Equal(t, appTmpl.ID, *actualAppPage.Data[0].ApplicationTemplateID)

			// WHEN
			subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

			// THEN
			actualAppPage = graphql.ApplicationPage{}
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
			require.NoError(t, err)

			require.Len(t, actualAppPage.Data, 0)
		})
	})
}

func createSubscription(t *testing.T, ctx context.Context, httpClient *http.Client, appTmpl graphql.ApplicationTemplate, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID string) {
	subscribeReq := buildSubscriptionRequest(t, ctx, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

	//unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
	//In case there isn't subscription it will fail-safe without error
	subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID)

	t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
	resp, err := httpClient.Do(subscribeReq)
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
	t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
}

func buildSubscriptionRequest(t *testing.T, ctx context.Context, subscriptionConsumerTenantID, subscriptionProviderSubaccountID string) *http.Request {
	apiPath := fmt.Sprintf("/saas-manager/v1/application/tenants/%s/subscriptions", subscriptionConsumerTenantID)
	subscribeReq, err := http.NewRequest(http.MethodPost, conf.SubscriptionConfig.URL+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
	require.NoError(t, err)
	subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, "tenantFetcherClaims")
	subscribeReq.Header.Add(subscription.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	subscribeReq.Header.Add(subscription.ContentTypeHeader, subscription.ContentTypeApplicationJson)
	subscribeReq.Header.Add(conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

	return subscribeReq
}
