package application

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/tests/director/tests/example"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

var (
	subject                    = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	subjectTwo                 = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=3c0fe289-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	consumerType               = "Integration System"          // should be a valid consumer type
	tenantAccessLevels         = []string{"account", "global"} // should be a valid tenant access level
	region1                    = "us10"
	region2                    = "us20"
	regionPlaceholderJSONPath1 = "$.path"
	regionPlaceholderJSONPath2 = "$.path-new"
)

func TestCreateApplicationTemplate(t *testing.T) {
	ctx := context.Background()

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "int-system-ord-service-consumption")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Run("Success for global template", func(t *testing.T) {
		// GIVEN

		// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test to fail if it executes in less than a second
		testStartTime := time.Now().Add(-1 * time.Minute)

		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)

		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationNoTenant(ctx, oauthGraphQLClient, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", output)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, output.ID)
		require.NotEmpty(t, output.Name)

		t.Log("Check if application template was created")

		getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(output.ID)
		appTemplateOutput := graphql.ApplicationTemplate{}

		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)

		appTemplateInput.ApplicationInput.Labels["applicationType"] = appTemplateName

		require.NoError(t, err)
		require.NotEmpty(t, appTemplateOutput)
		assert.True(t, time.Time(appTemplateOutput.CreatedAt).After(testStartTime))
		assert.True(t, time.Time(appTemplateOutput.UpdatedAt).After(testStartTime))

		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	})

	t.Run("Success for global template with product label created with certificate", func(t *testing.T) {
		// GIVEN
		cn := "app-template-global-product-cn"
		directorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, cn, conf.SkipSSLValidation)

		productLabelValue := []interface{}{"productLabelValue"}
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-product")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
		appTemplateInput.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, directorCertSecuredClient, output)
		require.NoError(t, err)

		csm := fixtures.FindCertSubjectMappingForApplicationTemplate(t, ctx, certSecuredGraphQLClient, output.ID, cn)

		// THEN
		require.Equal(t, output.Labels[conf.ApplicationTemplateProductLabel], productLabelValue)
		require.NotEmpty(t, output.ID)
		require.NotNil(t, csm, "Certificate subject mapping should be present but is missing")
	})

	t.Run("Success for regional template with product label created", func(t *testing.T) {
		// GIVEN
		productLabelValue := []interface{}{"productLabelValue"}
		appTemplateName := "app-template-name-product-1"

		appTemplateName1 := fixtures.CreateAppTemplateName(appTemplateName)
		appTemplateInput1 := fixtures.FixRegionalApplicationTemplate(appTemplateName1, region1, regionPlaceholderJSONPath1)
		appTemplateInput1.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := fixtures.CreateAppTemplateName(appTemplateName)
		appTemplateInput2 := fixtures.FixRegionalApplicationTemplate(appTemplateName2, region2, regionPlaceholderJSONPath1)
		appTemplateInput2.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create first regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, output1)
		require.NoError(t, err)

		t.Log("Create second regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, output2)
		require.NoError(t, err)

		// THEN
		require.Equal(t, productLabelValue, output1.Labels[conf.ApplicationTemplateProductLabel])
		require.Equal(t, region1, output1.Labels[tenantfetcher.RegionKey])
		require.Equal(t, productLabelValue, output1.Labels[conf.ApplicationTemplateProductLabel])
		require.Equal(t, region2, output2.Labels[tenantfetcher.RegionKey])
		require.NotEmpty(t, output1.ID)
		require.NotEmpty(t, output2.ID)
	})

	t.Run("Error for regional template with product label when the same product app template have different JSONPaths for their region placeholder", func(t *testing.T) {
		// GIVEN

		productLabelValue := []interface{}{"productLabelValue"}
		appTemplateName1 := fixtures.CreateAppTemplateName("app-template-name-product-1")
		appTemplateInput1 := fixtures.FixRegionalApplicationTemplate(appTemplateName1, region1, regionPlaceholderJSONPath1)
		appTemplateInput1.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := fixtures.CreateAppTemplateName("app-template-name-product-2")
		appTemplateInput2 := fixtures.FixRegionalApplicationTemplate(appTemplateName2, region2, regionPlaceholderJSONPath2)
		appTemplateInput2.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create first regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, output1)
		require.NoError(t, err)

		t.Log("Create second regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, output2)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf(`Regional Application Template input with "%s" label has a different "region" placeholder from the other Application Templates with the same label`, conf.ApplicationTemplateProductLabel))

	})

	t.Run("Error for regional template with product label but missing region label on ApplicationInputJSON", func(t *testing.T) {
		// GIVEN

		productLabelValue := []interface{}{"productLabelValue"}
		appTemplateName1 := fixtures.CreateAppTemplateName("app-template-name-product-1")
		appTemplateInput1 := fixtures.FixRegionalApplicationTemplate(appTemplateName1, region1, regionPlaceholderJSONPath1)
		appTemplateInput1.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplateInput1.ApplicationInput.Labels = graphql.Labels{
			"nonRegionLabel": "{{region}}",
		}
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, output1)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), `App Template with "region" label has a missing "region" label in the applicationInput`)
	})

	t.Run("Error when mixing global and regional App Templates for a single product label", func(t *testing.T) {
		// GIVEN

		// Create Regional and then Global
		productLabelValue1 := []interface{}{"productLabelValue1"}
		appTemplateName1 := fixtures.CreateAppTemplateName("app-template-name-product-1")
		regionalAppTemplateInput1 := fixtures.FixRegionalApplicationTemplate(appTemplateName1, region1, regionPlaceholderJSONPath1)
		regionalAppTemplateInput1.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue1
		regionalAppTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(regionalAppTemplateInput1)
		require.NoError(t, err)

		createRegionalApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(regionalAppTemplate1)
		regionalOutput1 := graphql.ApplicationTemplate{}

		appTemplateName2 := fixtures.CreateAppTemplateName("app-template-name-product-2")
		globalAppTemplateInput1 := fixtures.FixApplicationTemplate(appTemplateName2)
		globalAppTemplateInput1.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue1
		globalAppTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(globalAppTemplateInput1)
		require.NoError(t, err)

		createGlobalApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(globalAppTemplate1)
		globalOutput1 := graphql.ApplicationTemplate{}

		t.Log("Create regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createRegionalApplicationTemplateRequest1, &regionalOutput1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, regionalOutput1)
		require.NoError(t, err)

		t.Log("Create global application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createGlobalApplicationTemplateRequest1, &globalOutput1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, globalOutput1)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf(`Existing application template with "%s" label is regional. The input application template should contain a "region" label`, conf.ApplicationTemplateProductLabel))

		// Create Global and then Regional
		productLabelValue2 := []interface{}{"productLabelValue2"}
		appTemplateName3 := fixtures.CreateAppTemplateName("app-template-name-product-3")
		regionalAppTemplateInput2 := fixtures.FixRegionalApplicationTemplate(appTemplateName3, region1, regionPlaceholderJSONPath1)
		regionalAppTemplateInput2.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue2
		regionalAppTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(regionalAppTemplateInput2)
		require.NoError(t, err)

		createRegionalApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(regionalAppTemplate2)
		regionalOutput2 := graphql.ApplicationTemplate{}

		appTemplateName4 := fixtures.CreateAppTemplateName("app-template-name-product-4")
		globalAppTemplateInput2 := fixtures.FixApplicationTemplate(appTemplateName4)
		globalAppTemplateInput2.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue2
		globalAppTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(globalAppTemplateInput2)
		require.NoError(t, err)

		createGlobalApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(globalAppTemplate2)
		globalOutput2 := graphql.ApplicationTemplate{}

		t.Log("Create global application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createGlobalApplicationTemplateRequest2, &globalOutput2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, globalOutput2)
		require.NoError(t, err)

		t.Log("Create regional application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, createRegionalApplicationTemplateRequest2, &regionalOutput2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, regionalOutput2)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf(`Application Template with "%s" label is global and already exists`, conf.ApplicationTemplateProductLabel))
	})

	t.Run("Error for self register when distinguished label or product label have not been defined and the call is made with a certificate", func(t *testing.T) {
		// GIVEN
		appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "app-template-error-self-reg-cn", conf.SkipSSLValidation)

		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-invalid")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, appProviderDirectorCertSecuredClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient, output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("missing %q or %q label", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.ApplicationTemplateProductLabel))
	})

	t.Run("Error for self register when distinguished label and product label have been defined and the call is made with a certificate", func(t *testing.T) {
		// GIVEN
		appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "app-template-error-self-reg-product-label-cn", conf.SkipSSLValidation)

		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-invalid")
		appTemplateInputInvalid := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplateInputInvalid.Labels[conf.ApplicationTemplateProductLabel] = []string{"test1"}

		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInputInvalid)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, appProviderDirectorCertSecuredClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient, output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("should provide either %q or %q label - providing both at the same time is not allowed", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.ApplicationTemplateProductLabel))
	})

	t.Run("Error when Self Registered Application Template already exists for a given region and distinguished label key", func(t *testing.T) {
		// GIVEN
		appProviderDirectorCertSecuredClient1 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "app-template-region-exists-1-cn", conf.SkipSSLValidation)
		appProviderDirectorCertSecuredClient2 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "app-template-region-exists-2-cn", conf.SkipSSLValidation)

		ctx := context.Background()
		appTemplateName1 := fixtures.CreateAppTemplateName("app-template-name-self-reg-1")
		appTemplateInput1 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName1, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := fixtures.CreateAppTemplateName("app-template-name-self-reg-2")
		appTemplateInput2 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName2, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		t.Log("Create first application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, appProviderDirectorCertSecuredClient1, tenantID, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient1, output1)

		require.NoError(t, err)
		require.NotEmpty(t, output1.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, output1.Labels[tenantfetcher.RegionKey])

		// WHEN
		t.Log("Create second application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, appProviderDirectorCertSecuredClient2, tenantID, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient2, output2)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("Cannot have more than one application template with labels %q: %q and %q: %q", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion))
		require.Empty(t, output2.ID)
	})

	t.Run("Error for self register when not using the certificate flow", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		tenantId := tenant.TestTenants.GetDefaultTenantID()
		name := "app-template-name-invalid-flow"

		appTemplateName := fixtures.CreateAppTemplateName(name)
		appTemplateInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)

		// WHEN
		t.Log("Creating application template")
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("label %s is forbidden when creating Application Template in a non-cert flow", conf.SubscriptionConfig.SelfRegDistinguishLabelKey))
	})
}

func TestCreateApplicationTemplate_WhenApplicationTypeLabelIsSameAsApplicationTemplateName(t *testing.T) {
	ctx := context.Background()

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "int-system-ord-service-consumption")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	// GIVEN
	appTemplateName := "SAP app-template"
	appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplate)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)

	t.Log("Check if application template was created")
	appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, appTemplate.ID)

	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
}

func TestCreateApplicationTemplate_WhenApplicationTypeLabelIsDifferentFromApplicationTemplateName(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appTemplateInput := fixtures.FixApplicationTemplate("SAP app-template")
	appTemplateInput.ApplicationInput.Labels["applicationType"] = "random-app-type"

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "int-system-ord-service-consumption")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplate)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)

	t.Log("Check if application template was created")
	appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, appTemplate.ID)

	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
}

func TestCreateApplicationTemplate_SameNamesAndRegion(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	region := "region-02"

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "create-app-template-same-region")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplateName := "SAP app-template"
	appTemplateOneInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = region

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, appTemplateOne.ID)
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = region

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateTwo)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name \"SAP app-template\" and region %s already exists", region))
}

func TestCreateApplicationTemplate_SameNamesAndDifferentRegions(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "create-app-template-diff-region")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplateName := "SAP app-template"
	appTemplateOneInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = region1

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, appTemplateOne.ID)

	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = region2

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Check if application template two was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, appTemplateTwo.ID)
	appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	require.NotEmpty(t, appTemplateTwoOutput)
	assertions.AssertApplicationTemplate(t, appTemplateTwoInput, appTemplateTwoOutput)
}

func TestCreateApplicationTemplate_DifferentNamesAndDistinguishLabelsAndSameRegionsAndApplicationTypeLabels(t *testing.T) {
	ctx := context.Background()
	appProviderDirectorOnboardingCertSecuredClient1 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "DifferentNamesAndDistinguishLabel1-technical", conf.SkipSSLValidation)
	appProviderDirectorOnboardingCertSecuredClient2 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "DifferentNamesAndDistinguishLabel2-technical", conf.SkipSSLValidation)
	appProviderDirectorOnboardingCertSecuredClient3 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "DifferentNamesAndDistinguishLabel3-technical", conf.SkipSSLValidation)
	appProviderDirectorOnboardingCertSecuredClient4 := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "DifferentNamesAndDistinguishLabel4-technical", conf.SkipSSLValidation)

	applicationTypeLabelValue := "SAP app-template"
	appTemplateRegion := conf.SubscriptionConfig.SelfRegRegion

	appTemplateName1 := "SAP app-template-one"
	appTemplateName2 := "SAP app-template-two"
	appTemplateName3 := "SAP app-template-three"

	distinguishLabelValue2 := "other-distinguished-label"
	distinguishLabelValue3 := "another-one-distinguished-label"

	appTemplateOneInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName1, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Logf("Create first application template with name: %q, distinguishLabel: %q and region: %q", appTemplateName1, conf.SubscriptionConfig.SelfRegDistinguishLabelValue, conf.SubscriptionConfig.SelfRegRegion)
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorOnboardingCertSecuredClient1, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorOnboardingCertSecuredClient1, appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if first application template was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, appProviderDirectorOnboardingCertSecuredClient1, tenantID, appTemplateOne.ID)

	appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateOneInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.TestProviderSubaccountID
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName2, distinguishLabelValue2)
	appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue

	t.Logf("Create second application template with name: %q, applicationType: %q, distinguishLabel: %q and region: %q", appTemplateName2, applicationTypeLabelValue, distinguishLabelValue2, conf.SubscriptionConfig.SelfRegRegion)
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorOnboardingCertSecuredClient2, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorOnboardingCertSecuredClient2, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Check if second application template was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, appProviderDirectorOnboardingCertSecuredClient2, tenantID, appTemplateTwo.ID)

	appTemplateTwoInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateTwoOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateTwoInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.TestProviderSubaccountID
	appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateTwoOutput)
	assertions.AssertApplicationTemplate(t, appTemplateTwoInput, appTemplateTwoOutput)

	appTemplateThreeInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName2, distinguishLabelValue3)
	appTemplateThreeInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue

	t.Logf("Create third application template with name: %q, applicationType: %q, distinguishLabel: %q and region: %q", appTemplateName2, applicationTypeLabelValue, distinguishLabelValue3, conf.SubscriptionConfig.SelfRegRegion)
	appTemplateThree, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorOnboardingCertSecuredClient3, tenantID, appTemplateThreeInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorOnboardingCertSecuredClient3, appTemplateThree)

	t.Log("Check if third application template was not created, because it has the same name and region as the second app template")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name %q and region %s already exists", appTemplateName2, appTemplateRegion))

	appTemplateFourInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName3, distinguishLabelValue2)
	appTemplateFourInput.ApplicationInput.Labels["applicationType"] = applicationTypeLabelValue

	t.Logf("Create fourth application template with name: %q, applicationType: %q, distinguishLabel: %q and region: %q", appTemplateName3, applicationTypeLabelValue, distinguishLabelValue2, conf.SubscriptionConfig.SelfRegRegion)
	appTemplateFour, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorOnboardingCertSecuredClient4, tenantID, appTemplateFourInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorOnboardingCertSecuredClient4, appTemplateFour)

	t.Log("Check if fourth application template was not created, because it has the same distinguish label and region as the third app template")
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("Cannot have more than one application template with labels %q: %q and %q: %q", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, distinguishLabelValue2, tenantfetcher.RegionKey, appTemplateRegion))

}

func TestCreateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"
	sapProvider := "SAP"
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	ctx := context.Background()

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "CreateApplicationTemplate_NotValid", conf.SkipSSLValidation)

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "app-template-not-valid")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	testCases := []struct {
		Name                                  string
		AppTemplateName                       string
		AppTemplateAppInputJSONNameProperty   *string
		AppTemplateAppInputJSONLabelsProperty *map[string]interface{}
		AppTemplatePlaceholders               []*graphql.PlaceholderDefinitionInput
		AppInputDescription                   *string
		ExpectedErrMessage                    string
		IsSelfReg                             bool
		Client                                *gcli.Client
	}{
		{
			Name:                                  "not compliant name",
			AppTemplateName:                       "not-compliant-name",
			AppTemplateAppInputJSONNameProperty:   str.Ptr("not-compliant-name"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "name": "{{name}}", "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNameJSONPath,
				},
			},
			AppInputDescription: nil,
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                  "missing mandatory applicationInput name property",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr(""),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "Invalid data ApplicationTemplateInput [appInput=name: cannot be blank.]",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                  "missing mandatory applicationInput displayName label property",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name")},
			AppTemplatePlaceholders:               []*graphql.PlaceholderDefinitionInput{},
			AppInputDescription:                   ptr.String("test"),
			ExpectedErrMessage:                    "applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed.",
			Client:                                appProviderDirectorCertSecuredClient,
			IsSelfReg:                             true,
		},
		{
			Name:                                  "unused placeholder defined in the placeholders array",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
				{
					Name:        "unused-placeholder-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "application input does not use provided placeholder [name=unused-placeholder-name]",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                  "undefined placeholder applicationInput",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr("{{undefined-placeholder-name}}"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "Placeholder [name=undefined-placeholder-name] is used in the application input but it is not defined in the Placeholders array",
			Client:              oauthGraphQLClient,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			appTemplateInput := fixtures.FixApplicationTemplate(testCase.AppTemplateName)
			if testCase.IsSelfReg {
				appTemplateInput = fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(testCase.AppTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
			}

			if testCase.AppInputDescription != nil {
				appTemplateInput.ApplicationInput.Description = testCase.AppInputDescription
			}
			appTemplateInput.Placeholders = testCase.AppTemplatePlaceholders
			if testCase.AppTemplateAppInputJSONNameProperty != nil {
				appTemplateInput.ApplicationInput.Name = *testCase.AppTemplateAppInputJSONNameProperty
			}
			appTemplateInput.ApplicationInput.ProviderName = &sapProvider

			if testCase.AppTemplateAppInputJSONLabelsProperty != nil {
				appTemplateInput.ApplicationInput.Labels = *testCase.AppTemplateAppInputJSONLabelsProperty
			}
			appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
			require.NoError(t, err)

			createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
			output := graphql.ApplicationTemplate{}

			// WHEN
			t.Log("Create application template")
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, testCase.Client, tenantID, createApplicationTemplateRequest, &output)
			defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, testCase.Client, output)

			//THEN
			require.NotNil(t, err)
			if testCase.ExpectedErrMessage != "" {
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestUpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)
	oldWebhookCount := len(appTemplate.Webhooks)
	oldWebhookID := appTemplate.Webhooks[0].ID
	oldWebhookUrl := appTemplate.Webhooks[0].URL

	newAppCreateInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url2.com"),
	}}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template without override and one webhook")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	require.NotEmpty(t, updateOutput.Webhooks)
	newWebhookCount := len(updateOutput.Webhooks)
	newWebhookID := updateOutput.Webhooks[0].ID
	newWebhookUrl := updateOutput.Webhooks[0].URL

	require.Equal(t, oldWebhookCount, newWebhookCount)
	require.NotEqual(t, oldWebhookID, newWebhookID)
	require.NotEqual(t, oldWebhookUrl, newWebhookUrl)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))

	example.SaveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestUpdateApplicationTemplateWithProductLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultProductLabel(appTemplateName, conf.ApplicationTemplateProductLabel, []string{"E2E_TEST"})
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	newAppCreateInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)
}

func TestUpdateApplicationTemplateWithOverride(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)

	newAppCreateInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateInput.Webhooks = []*graphql.WebhookInput{}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateWithOverrideRequest(appTemplate.ID, true, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template with override and empty list of webhooks")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)
	require.Empty(t, updateOutput.Webhooks)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))
}

func TestUpdateApplicationTemplateWithoutOverrideWithoutWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-without-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)
	oldWebhookCount := len(appTemplate.Webhooks)
	oldWebhookID := appTemplate.Webhooks[0].ID
	oldWebhookUrl := appTemplate.Webhooks[0].URL

	newAppCreateInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateInput.Webhooks = []*graphql.WebhookInput{}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template without override and empty list of webhooks")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	require.NotEmpty(t, updateOutput.Webhooks)
	newWebhookCount := len(updateOutput.Webhooks)
	newWebhookID := updateOutput.Webhooks[0].ID
	newWebhookUrl := updateOutput.Webhooks[0].URL

	require.Equal(t, oldWebhookCount, newWebhookCount)
	require.Equal(t, oldWebhookID, newWebhookID)
	require.Equal(t, oldWebhookUrl, newWebhookUrl)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))
}

func TestUpdateLabelsOfApplicationTemplateFailsWithInsufficientScopes(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "create-app-template-same-region")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTmplInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, Labels: map[string]interface{}{"label1": "test"}, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	t.Log("Should return error because there is no application_template.labels:write scope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient scopes provided")
}

func TestUpdateApplicationTypeLabelOfApplicationsWhenAppTemplateNameIsUpdated(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	firstTenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	secondTenantId := tenant.TestTenants.List()[1].ExternalTenant

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, firstTenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, firstTenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, firstTenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, firstTenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, firstTenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Create application from template for the first tenant")
	appFromTmplFirstTenant := graphql.ApplicationFromTemplateInput{
		TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app1-e2e-update-applicationType-label",
			},
			{
				Placeholder: "display-name",
				Value:       "app1 description",
			},
		},
	}

	appFromTmplGQLFirstTenant, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplFirstTenant)
	require.NoError(t, err)

	createAppFromTmplRequestFirstTenant := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQLFirstTenant)
	outputAppFirstTenant := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, firstTenantId, createAppFromTmplRequestFirstTenant, &outputAppFirstTenant)
	defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, firstTenantId, &outputAppFirstTenant)
	require.NoError(t, err)

	t.Log("Create application from template for the second tenant")
	appFromTmplSecondTenant := graphql.ApplicationFromTemplateInput{
		TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app2-e2e-update-applicationType-label",
			},
			{
				Placeholder: "display-name",
				Value:       "app2 description",
			},
		},
	}

	appFromTmplGQLSecondTenant, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSecondTenant)
	require.NoError(t, err)

	createAppFromTmplRequestSecondTenant := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQLSecondTenant)
	outputAppSecondTenant := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, secondTenantId, createAppFromTmplRequestSecondTenant, &outputAppSecondTenant)
	defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, secondTenantId, &outputAppSecondTenant)
	require.NoError(t, err)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	t.Log("Get updated application for the first tenant")
	app1 := fixtures.GetApplication(t, ctx, oauthGraphQLClient, firstTenantId, outputAppFirstTenant.ID)
	assert.Equal(t, outputAppFirstTenant.ID, app1.ID)

	t.Log("Get updated application for the second tenant")
	app2 := fixtures.GetApplication(t, ctx, oauthGraphQLClient, secondTenantId, outputAppSecondTenant.ID)
	assert.Equal(t, outputAppSecondTenant.ID, app2.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	t.Log("Check if applicationType label of application for the first tenant was updated")
	assert.Equal(t, app1.Labels["applicationType"], newName)

	t.Log("Check if applicationType label of application for the second tenant was updated")
	assert.Equal(t, app2.Labels["applicationType"], newName)
}

func TestUpdateApplicationTemplate_AlreadyExistsInTheSameRegion(t *testing.T) {
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	region := "region-01"

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "int-system-update-app-template")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplateOneInput := fixtures.FixApplicationTemplate("SAP app-template")
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = region

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	appTemplateTwoInput := fixtures.FixApplicationTemplate("SAP app-template-two")
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = region

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Update the name of second application template to be the same as the name of the first one")
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{
		Name:             appTemplateOne.Name,
		Description:      appTemplateOne.Description,
		ApplicationInput: appTemplateOneInput.ApplicationInput,
		Placeholders:     appTemplateOneInput.Placeholders,
		AccessLevel:      appTemplateOne.AccessLevel,
	}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)
	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplateTwo.ID, appTemplateGQL)

	updateOutput := graphql.ApplicationTemplate{}
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateAppTemplateRequest, &updateOutput)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name \"SAP app-template\" and region %s already exists", region))
}

func TestUpdateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"
	sapProvider := "SAP"
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	ctx := context.Background()

	appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "CreateApplicationTemplate_NotValid", conf.SkipSSLValidation)

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, "app-template-not-valid")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)
	testCases := []struct {
		Name                                     string
		NewAppTemplateName                       string
		NewAppTemplateAppInputJSONNameProperty   *string
		NewAppTemplateAppInputJSONLabelsProperty *map[string]interface{}
		NewAppTemplatePlaceholders               []*graphql.PlaceholderDefinitionInput
		AppInputDescription                      *string
		ExpectedErrMessage                       string
		IsSelfReg                                bool
		Client                                   *gcli.Client
	}{
		{
			Name:               "not compliant name",
			NewAppTemplateName: "not-compliant-name",
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test {{display-name}}"),
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                     "missing mandatory applicationInput name property",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SubscriptionConfig.SelfRegRegion),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr(""),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"name": "{{name}}", "displayName": "{{display-name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "Invalid data ApplicationTemplateUpdateInput [appInput=name: cannot be blank.]",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                     "missing mandatory applicationInput displayName label property",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SubscriptionConfig.SelfRegRegion),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"name": "{{name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed.",
			IsSelfReg:           true,
			Client:              appProviderDirectorCertSecuredClient,
		},
		{
			Name:                                     "unused placeholder defined in the placeholders array",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
				{
					Name:        "unused-placeholder-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "application input does not use provided placeholder [name=unused-placeholder-name]",
			Client:              oauthGraphQLClient,
		},
		{
			Name:                                     "undefined placeholder applicationInput",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr("{{undefined-placeholder-name}}"),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test"),
			ExpectedErrMessage:  "Placeholder [name=undefined-placeholder-name] is used in the application input but it is not defined in the Placeholders array",
			Client:              oauthGraphQLClient,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appTemplateName := fixtures.CreateAppTemplateName("app-template")

			t.Log("Create application template")
			appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
			if testCase.IsSelfReg {
				appTemplateInput = fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
			}
			appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, testCase.Client, tenantID, appTemplateInput)
			defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, testCase.Client, appTemplate)

			require.NoError(t, err)
			require.NotEmpty(t, appTemplate.ID)

			// WHEN
			t.Log("Update application template")
			appJSONInput := &graphql.ApplicationJSONInput{
				Name:         "{{name}}",
				ProviderName: ptr.String("compass-tests"),
				Labels: graphql.Labels{
					"a": []string{"b", "c"},
					"d": []string{"e", "f"},
				},
				Webhooks: []*graphql.WebhookInput{{
					Type: graphql.WebhookTypeConfigurationChanged,
					URL:  ptr.String("http://url.com"),
				}},
				HealthCheckURL: ptr.String("http://url.valid"),
			}

			if testCase.NewAppTemplateAppInputJSONNameProperty != nil {
				appJSONInput.Name = *testCase.NewAppTemplateAppInputJSONNameProperty
			}
			appJSONInput.Description = testCase.AppInputDescription
			appJSONInput.ProviderName = &sapProvider
			if testCase.NewAppTemplateAppInputJSONLabelsProperty != nil {
				appJSONInput.Labels = *testCase.NewAppTemplateAppInputJSONLabelsProperty
			}

			appTemplateUpdateInput := graphql.ApplicationTemplateUpdateInput{Name: testCase.NewAppTemplateName, ApplicationInput: appJSONInput, Placeholders: testCase.NewAppTemplatePlaceholders, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
			appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateUpdateInput)

			updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
			updateOutput := graphql.ApplicationTemplate{}

			err = testctx.Tc.RunOperation(ctx, testCase.Client, updateAppTemplateRequest, &updateOutput)

			//THEN
			require.NotNil(t, err)
			if testCase.ExpectedErrMessage != "" {
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestDeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Delete application template")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application template was deleted")

	out := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)

	require.Empty(t, out)
	example.SaveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestDeleteApplicationTemplateWithCertSubjMapping(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	cn := "DeleteAppTemplateWithCSM"

	appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, cn, conf.SkipSSLValidation)

	t.Logf("Create application template with name %q", appTemplateName)
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorCertSecuredClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	// WHEN
	t.Logf("Delete application template with id %q", appTemplate.ID)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	err = testctx.Tc.RunOperation(ctx, appProviderDirectorCertSecuredClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if all certificate subject mappings were deleted")
	csm := fixtures.FindCertSubjectMappingForApplicationTemplate(t, ctx, certSecuredGraphQLClient, appTemplate.ID, cn)
	require.Nil(t, csm)
}

func TestDeleteApplicationTemplateBeforeDeletingAssociatedApplicationsWithIt(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	name := "template"

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplateName := fixtures.CreateAppTemplateName(name)
	appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)

	t.Log("Creating application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)

	appFromTemplate := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "name",
			Value:       "new-name",
		},
		{
			Placeholder: "display-name",
			Value:       "new-display-name",
		}}}
	appFromTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTemplate)
	require.NoError(t, err)
	createAppFromTemplateRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateGQL)
	outputApp := graphql.ApplicationExt{}

	t.Logf("Creating application from application template with id %s", appTemplate.ID)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, createAppFromTemplateRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)
	require.NoError(t, err)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Logf("Deleting application template with id %s when application with id %s is associated with it", appTemplate.ID, outputApp.ID)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.Error(t, err)

	//THEN
	t.Logf("Checking if application template with id %s was deleted", appTemplate.ID)

	out := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	require.NotEmpty(t, out)

	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	require.Equal(t, "new-name", outputApp.Application.Name)
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := fixtures.CreateAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(appTemplate.ID)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Get application template")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, getApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	//THEN
	t.Log("Check if application template was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := fixtures.CreateAppTemplateName("app-template-1")
	name2 := fixtures.CreateAppTemplateName("app-template-2")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application templates")
	appTmplInput1 := fixtures.FixApplicationTemplate(name1)
	appTemplate1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate1)
	require.NoError(t, err)

	directorCertClientRegion2 := createDirectorCertClientForAnotherRegion(t, ctx, "query_app_template")

	appTmplInput2 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name2, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplate2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertClientRegion2, tenantId, appTmplInput2)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, directorCertClientRegion2, appTemplate2)
	require.NoError(t, err)

	pageSize := 200
	pageCursor := ""
	hasNextPage := true

	var applicationTemplates []*graphql.ApplicationTemplate
	for hasNextPage {
		getApplicationTemplatesRequest := fixtures.FixGetApplicationTemplatesWithPagination(pageSize, pageCursor)
		if pageCursor == "" {
			example.SaveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
		}

		output := graphql.ApplicationTemplatePage{}

		// WHEN
		t.Logf("List application templates page with size %d and cursor %s", pageSize, pageCursor)
		err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, getApplicationTemplatesRequest, &output)
		require.NoError(t, err)

		applicationTemplates = append(applicationTemplates, output.Data...)

		pageCursor = string(output.PageInfo.EndCursor)
		hasNextPage = output.PageInfo.HasNextPage
	}

	t.Log("Check if application templates were received")
	appTemplateIDs := []string{appTemplate1.ID, appTemplate2.ID}
	t.Logf("Created templates are with IDs: %v ", appTemplateIDs)
	found := 0
	for _, tmpl := range applicationTemplates {
		t.Logf("Checked template from query response is: %s ", tmpl.ID)
		if str.ContainsInSlice(appTemplateIDs, tmpl.ID) {
			found++
		}
	}
	assert.Equal(t, 2, found)
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{display-name}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "name",
			Value:       "new-name",
		},
		{
			Placeholder: "display-name",
			Value:       "new-display-name",
		}}}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestRegisterApplicationFromTemplateWithOrdWebhook(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeOpenResourceDiscovery,
		URL:  ptr.String("http://test.test"),
	}}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "int-system-ord-webhook")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)

	appFromTmpl := fixtures.FixApplicationFromTemplateInput(appTemplateName, "name", "new-name", "display-name", "new-display-name")
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)

	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}

	//WHEN
	t.Log("Create application from application template")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.Equal(t, 1, len(outputApp.Operations))
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, outputApp.Operations[0].OperationType)
}

func TestRegisterApplicationFromTemplateWithTemplateID(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTemplateName2 := fixtures.CreateAppTemplateName("template-1")
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "int-system-ord-service-consumption")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template in the first region")
	appTemplateOneInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = region1
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.Equal(t, appTemplateName, appTemplateOne.Name)

	t.Log("Check if application template in the first region was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOneOutput)

	t.Log("Create application template in the second region")
	appTemplateTwoInput := fixtures.FixApplicationTemplate(appTemplateName2)
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = region2

	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.Equal(t, appTemplateName2, appTemplateTwo.Name)

	t.Log("Check if application template in the second region was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwoOutput)

	require.NotEqual(t, appTemplateOne.ID, appTemplateTwo.ID)

	t.Log("Create application using template id")
	appFromTmpl := graphql.ApplicationFromTemplateInput{
		ID:           &appTemplateTwo.ID,
		TemplateName: appTemplateTwo.Name,
		Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app-name",
			},
			{
				Placeholder: "display-name",
				Value:       "app-display-name",
			},
		},
	}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputWithTemplateIDToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}

	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.Equal(t, appTemplateTwo.ID, *outputApp.Application.ApplicationTemplateID)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test app-display-name", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template using template name and id")
}

func TestRegisterApplicationFromTemplateWithPlaceholderPayload(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name"
	displayNameJSONPath := "$.displayName"
	placeholdersPayload := `{\"name\": \"appName\", \"displayName\":\"appDisplayName\"}`
	appTemplateName := fixtures.CreateAppTemplateName("templateForPlaceholdersPayload")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, PlaceholdersPayload: &placeholdersPayload}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "appName", outputApp.Application.Name)
	require.Equal(t, "test appDisplayName", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template with placeholder payload")
}

func TestRegisterApplicationFromTemplate_DifferentSubaccount(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{display-name}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	appProviderDirectorCertSecuredClient := certprovider.NewDirectorCertClientWithOtherSubject(t, ctx, conf.ExternalCertProviderConfig, conf.DirectorExternalCertSecuredURL, "app-template-different-sa-cn", conf.SkipSSLValidation)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorCertSecuredClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appProviderDirectorCertSecuredClient, appTmpl)
	require.NoError(t, err)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

	directorCertSecuredClient := createDirectorCertClientForAnotherRegion(t, ctx, "register")

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "name",
			Value:       "new-name",
		},
		{
			Placeholder: "display-name",
			Value:       "new-display-name",
		}}}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	// WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, tenantId, createAppFromTmplRequest, &outputApp)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name %q and consumer id REDACTED_%x not found", appTemplateName, sha256.Sum256([]byte(conf.TestProviderSubaccountIDRegion2))))
}

func TestAddWebhookToApplicationTemplateWithTenant(t *testing.T) {
	ctx := context.Background()
	name := "app-template"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	t.Log("Request Client Credentials for Integration System")
	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(name)
	appTmplInput.Labels = graphql.Labels{
		"a":                             []string{"b", "c"},
		"d":                             []string{"e", "f"},
		"displayName":                   "{{display-name}}",
		conf.GlobalSubaccountIDLabelKey: tenantId,
	}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Add Webhook to application template with invalid tenant")
	// add
	url := "http://new-webhook.url"
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"

	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	actualWebhook := graphql.Webhook{}
	addReq := fixtures.FixAddWebhookToTemplateRequest(appTemplate.ID, webhookInStr)
	example.SaveExampleInCustomDir(t, addReq.Query(), example.AddWebhookCategory, "add application template webhook")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant.TestTenants.GetSystemFetcherTenantID(), addReq, &actualWebhook)
	require.Error(t, err)
	require.Contains(t, err.Error(), "the provided tenant c395681d-11dd-4cde-bbcf-570b4a153e79 and the parent tenant 5577cf46-4f78-45fa-b55f-a42a3bdba868 do not match")

	t.Log("Add Webhook to application template with valid tenant")
	actualWebhook = graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, addReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
	id := actualWebhook.ID
	require.NotNil(t, id)

	t.Log("Get Application Template")
	updatedAppTemplate := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	assert.Len(t, updatedAppTemplate.Webhooks, 1)

	t.Log("Update Application Template webhook with tenant")
	urlUpdated := "http://updated-webhook.url"
	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &urlUpdated,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	t.Log("Delete Application Template webhook with tenant")
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, deleteReq, &actualWebhook)
	require.NoError(t, err)
}

func TestAddWebhookToApplicationTemplateWithoutTenant(t *testing.T) {
	ctx := context.Background()
	name := "app-template"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	t.Log("Request Client Credentials for Integration System")
	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixtures.FixApplicationTemplate(name)
	appTmplInput.Labels = graphql.Labels{
		"a":                             []string{"b", "c"},
		"d":                             []string{"e", "f"},
		"displayName":                   "{{display-name}}",
		conf.GlobalSubaccountIDLabelKey: tenantId,
	}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Add Webhook to application template without tenant")
	url := "http://new-webhook.url"
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	actualWebhook := graphql.Webhook{}
	addReq := fixtures.FixAddWebhookToTemplateRequest(appTemplate.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, addReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
	id := actualWebhook.ID
	require.NotNil(t, id)

	t.Log("Get Application Template")
	updatedAppTemplate := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	assert.Len(t, updatedAppTemplate.Webhooks, 1)

	t.Log("Update Application Template webhook without tenant")
	urlUpdated := "http://updated-webhook.url"
	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &urlUpdated,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	t.Log("Delete Application Template webhook without tenant")
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, deleteReq, &actualWebhook)
	require.NoError(t, err)
}

func createDirectorCertClientForAnotherRegion(t *testing.T, ctx context.Context, subjectSuffix string) *gcli.Client {
	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:         conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace:    conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestRegion2SecretName:     conf.ExternalCertProviderConfig.CertSvcInstanceTestRegion2SecretName,
		ExternalCertCronjobContainerName:         conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:                  conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:                  conf.ExternalCertProviderConfig.TestExternalCertSubjectRegion2 + subjectSuffix,
		ExternalClientCertCertKey:                conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:                 conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalClientCertExpectedIssuerLocality: &conf.ExternalClientCertExpectedIssuerLocalityRegion2,
		ExternalCertProvider:                     certprovider.CertificateService,
	}
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	return gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)
}

func fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	input.ApplicationInput.BaseURL = str.Ptr(fmt.Sprintf(baseURLTemplate, "{{subdomain}}", "{{region}}"))
	input.Placeholders = append(input.Placeholders, &graphql.PlaceholderDefinitionInput{Name: "subdomain"}, &graphql.PlaceholderDefinitionInput{Name: "region"})
	return input
}

func fixAppTemplateInputWithDistinguishLabel(name, distinguishedLabel string) graphql.ApplicationTemplateInput {
	return fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, distinguishedLabel)
}
