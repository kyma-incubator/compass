package tests

import (
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/require"
)

const (
	TestScenario = "test-scenario"
)

func TestAppsForRuntimeWithCertificates(t *testing.T) {
	defer fixtures.DeleteFormationWithinTenant(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, TestScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, TestScenario)

	appIdToCommonName := make(map[string]string)

	// Register first Application
	firstApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-first-app",
		Labels: map[string]interface{}{cfg.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C)},
	})
	defer fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &firstApp)
	require.NoError(t, err)
	require.NotEmpty(t, firstApp.ID)

	defer fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, firstApp.ID, []string{TestScenario})
	fixtures.AssignApplicationInScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, firstApp.ID, []string{TestScenario})

	// Issue a certificate for the first Application
	certResult, configuration := clients.GenerateApplicationCertificate(t, directorAppsForRuntimeClient, connectorClient, firstApp.ID, clientKey)
	certs.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)

	decodedClientCert := certs.DecodeCert(t, certResult.ClientCertificate)
	firstAppCommonName := decodedClientCert.Subject.CommonName
	require.NotEmpty(t, firstAppCommonName)
	appIdToCommonName[firstApp.ID] = firstAppCommonName
	defer certs.Cleanup(t, configmapCleaner, certResult)

	// Register Runtime
	input := fixtures.FixRuntimeRegisterInput("test-runtime")

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, input, cfg.GatewayOauth)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	defer fixtures.UnassignRuntimeFromScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, runtime.ID, []string{TestScenario})
	fixtures.AssignRuntimeInScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, runtime.ID, []string{TestScenario})

	// Issue a certificate for the Runtime
	ott, err := directorAppsForRuntimeClient.GenerateRuntimeToken(t, runtime.ID)
	require.NoError(t, err)
	rtCertResult, rtConfiguration := clients.GenerateRuntimeCertificate(t, &ott, connectorClient, clientKey)
	certs.AssertCertificate(t, rtConfiguration.CertificateSigningRequestInfo.Subject, rtCertResult)
	defer certs.Cleanup(t, configmapCleaner, rtCertResult)

	// Register second Application
	secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-second-app",
		Labels: map[string]interface{}{cfg.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C)},
	})
	defer fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &secondApp)
	require.NoError(t, err)
	require.NotEmpty(t, secondApp.ID)

	defer fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, secondApp.ID, []string{TestScenario})
	fixtures.AssignApplicationInScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, secondApp.ID, []string{TestScenario})

	// Issue a certificate for the second Application
	secondCertResult, secondConfiguration := clients.GenerateApplicationCertificate(t, directorAppsForRuntimeClient, connectorClient, secondApp.ID, clientKey)
	certs.AssertCertificate(t, secondConfiguration.CertificateSigningRequestInfo.Subject, secondCertResult)

	decodedClientCertSecondApp := certs.DecodeCert(t, secondCertResult.ClientCertificate)
	secondAppCommonName := decodedClientCertSecondApp.Subject.CommonName
	require.NotEmpty(t, secondAppCommonName)
	appIdToCommonName[secondApp.ID] = secondAppCommonName
	defer certs.Cleanup(t, configmapCleaner, secondCertResult)

	// Call "ApplicationsForRuntime" and assert that sys_auth ids from the query match the Apps' certificate CN
	rtCertChain := certs.DecodeCertChain(t, rtCertResult.CertificateChain)
	securedClient := clients.NewCertificateSecuredConnectorClient(cfg.DirectorMtlsURL, clientKey, rtCertChain...)

	apps, err := fixtures.AppsForRuntime(ctx, securedClient.GraphQlClient, appsForRuntimeTenantID, runtime.ID)
	require.NoError(t, err)
	require.Equal(t, 2, len(apps.Data))
	for _, app := range apps.Data {
		certCN := appIdToCommonName[app.ID]
		require.NotEmpty(t, app.Auths)
		systemAuthID := app.Auths[0].ID
		require.Equal(t, certCN, systemAuthID)
	}
}
