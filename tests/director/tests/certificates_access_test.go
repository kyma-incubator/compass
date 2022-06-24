package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemAccess(t *testing.T) {
	// Build graphql director client configured with certificate
	ctx := context.Background()
	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	testCases := []struct {
		name           string
		tenant         string
		resourceSuffix string
		expectErr      bool
	}{
		{
			name:           "Integration System can manage account tenant entities",
			tenant:         tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemManagedAccount),
			resourceSuffix: "account-owned",
		},
		{
			name:           "Integration System can manage subaccount tenant entities",
			tenant:         tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemManagedSubaccount),
			resourceSuffix: "subaccount-owned",
		},
		{
			name:           "Integration System cannot manage customer tenant entities",
			tenant:         tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant),
			resourceSuffix: "customer-owned",
			expectErr:      true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Log(fmt.Sprintf("Trying to create application in account tenant %s", test.tenant))
			app, err := fixtures.RegisterApplication(t, ctx, directorCertSecuredClient, fmt.Sprintf("app-%s", test.resourceSuffix), test.tenant)
			defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, test.tenant, &app)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, app.ID)
			}

			t.Log(fmt.Sprintf("Trying to list applications in account tenant %s", test.tenant))
			getAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			apps := graphql.ApplicationPage{}
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, test.tenant, getAppReq, &apps)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, apps.Data)
			}

			t.Log(fmt.Sprintf("Trying to register runtime in account tenant %s", test.tenant))
			rtmInput := fixRuntimeInput(fmt.Sprintf("runtime-%s", test.resourceSuffix))
			rt, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, test.tenant, &rtmInput)
			defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, test.tenant, &rt)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, rt.ID)
			}

			t.Log(fmt.Sprintf("Trying to create application template in account tenant %s via client certificate", test.tenant))

			name := fmt.Sprintf("app-template-%s", test.resourceSuffix)
			appTemplateName := createAppTemplateName(name)
			appTmplInput := fixAppTemplateInput(appTemplateName)
			at, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, test.tenant, appTmplInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, test.tenant, &at)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, at.ID)
			}
		})
	}
}

func TestApplicationTemplateWithExternalCertificate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	// Build graphql director client configured with certificate
	externalCertProviderConfig := buildExternalCertProviderConfig()

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Run("Create Application Template with external certificate", func(t *testing.T) {
		// WHEN
		t.Log("Create application template")
		name := createAppTemplateName("create-app-template-with-external-cert-name")
		appTemplateInput := fixAppTemplateInput(name)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, &appTemplate)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate.ID)

		t.Log("Check if application template was created")
		appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)
		require.NotEmpty(t, appTemplateOutput)

		// Enhance input to match the newly created labels
		appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
		appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID
		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	})

	t.Run("Delete Application Template with external certificate", func(t *testing.T) {
		t.Log("Create application template")
		name := createAppTemplateName("delete-app-template-with-external-cert-name")
		appTemplateInput := fixAppTemplateInput(name)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, &appTemplate)

		require.NoError(t, err)
		require.NotEmpty(t, appTemplate.ID)

		// WHEN
		t.Log("Delete application template")
		fixtures.DeleteApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)

		//THEN
		t.Log("Check if application template was deleted")
		out := fixtures.GetApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)

		require.Empty(t, out)
	})
}

func TestAddBundleToApplicationWithExternalCertificate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	// Build graphql director client configured with certificate
	externalCertProviderConfig := buildExternalCertProviderConfig()

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := createAppTemplateName("app-template-with-external-cert-name")

	t.Log("Create application template")
	appTemplateInput := fixAppTemplateInput(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantId, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, &appTemplate)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	// Register application from app template
	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: name, Values: []*graphql.TemplateValueInput{{Placeholder: "name", Value: "test-name"}, {Placeholder: "display-name", Value: "test-name"}}}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, directorCertSecuredClient, tenantId, outputApp.ID)

	t.Run("Success", func(t *testing.T) {
		//WHEN
		bndlName := "test-bundle"
		bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID, bndlName)
		defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, outputApp)
	})

	t.Run("Error when no header is present", func(t *testing.T) {
		//WHEN
		bndlName := "test-bundle"
		in, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(fixtures.FixBundleCreateInput(bndlName))
		require.NoError(t, err)

		req := fixtures.FixAddBundleRequest(outputApp.ID, in)
		var resp graphql.BundleExt

		err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, req, &resp)

		//THEN
		require.Error(t, err)
	})
}

func buildExternalCertProviderConfig() certprovider.ExternalCertProviderConfig {
	return certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.ExternalCertProviderConfig.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, "integration-system-test", "external-cert", -1),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
	}
}
