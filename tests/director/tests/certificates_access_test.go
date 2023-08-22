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

	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, conf.ExternalCertTestIntSystemOUSubaccount, conf.TestExternalCertCN, conf.ExternalCertTestIntSystemCommonName)

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
		tenantName     string
		resourceSuffix string
		expectErr      bool
	}{
		{
			name:           "Integration System can manage account tenant entities",
			tenantName:     tenant.TestIntegrationSystemManagedAccount,
			resourceSuffix: "account-owned",
		},
		{
			name:           "Integration System can manage subaccount tenant entities",
			tenantName:     tenant.TestIntegrationSystemManagedSubaccount,
			resourceSuffix: "subaccount-owned",
		},
		{
			name:           "Integration System cannot manage customer tenant entities",
			tenantName:     tenant.TestDefaultCustomerTenant,
			resourceSuffix: "customer-owned",
			expectErr:      true,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {

			tenantObj := tenant.TestTenants.GetTenantByName(test.tenantName)
			tenantID := tenantObj.ExternalTenant

			t.Log(fmt.Sprintf("Trying to create application in account tenant %s", tenantID))
			app, err := fixtures.RegisterApplication(t, ctx, directorCertSecuredClient, fmt.Sprintf("app-%s", test.resourceSuffix), tenantID)
			defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &app)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, app.ID)
			}

			t.Log(fmt.Sprintf("Trying to list applications in account tenant %s", tenantID))
			getAppReq := fixtures.FixGetApplicationsRequestWithPagination()
			apps := graphql.ApplicationPage{}
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, tenantID, getAppReq, &apps)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, apps.Data)
			}

			t.Log(fmt.Sprintf("Trying to register runtime in account tenant %s", tenantID))
			rtmInput := fixRuntimeInput(fmt.Sprintf("runtime-%s", test.resourceSuffix))
			rt, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, tenantID, &rtmInput)
			defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &rt)
			if test.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, rt.ID)
			}

			if tenantObj.Type == tenant.Subaccount {
				t.Log(fmt.Sprintf("Trying to create application template in account tenant %s via client certificate", tenantID))

				name := fmt.Sprintf("app-template-%s", test.resourceSuffix)
				appTemplateName := createAppTemplateName(name)
				appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
				at, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantID, appTmplInput)
				defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, at)
				if test.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.NotEmpty(t, at.ID)
					require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, at.Labels[tenantfetcher.RegionKey])
				}
			}
		})
	}
}
