package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

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

	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, conf.ExternalCertTestIntSystemOUSubaccount, conf.ExternalCertCommonName, conf.ExternalCertTestIntSystemCommonName)

	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               replacer.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
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
			appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
			at, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, test.tenant, appTmplInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, test.tenant, at)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, at.ID)
				require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, at.Labels[tenantfetcher.RegionKey])
			}
		})
	}
}
