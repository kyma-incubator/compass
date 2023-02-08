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
	ScenariosLabel = "scenarios"
	TestScenario   = "test-scenario"
)

func TestAppsForRuntimeWithCertificates(t *testing.T) {
	defer fixtures.DeleteFormationWithinTenant(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, TestScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, TestScenario)

	appIdToCommonName := make(map[string]string)

	// Register first Application
	firstApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-first-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}, cfg.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C)},
	})
	defer fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &firstApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, firstApp.ID)
	require.NoError(t, err)
	require.NotEmpty(t, firstApp.ID)

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
	input.Labels[ScenariosLabel] = []string{TestScenario}
	runtime := fixtures.RegisterKymaRuntime(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, input, cfg.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	// Issue a certificate for the Runtime
	ott, err := directorAppsForRuntimeClient.GenerateRuntimeToken(t, runtime.ID)
	require.NoError(t, err)
	rtCertResult, rtConfiguration := clients.GenerateRuntimeCertificate(t, &ott, connectorClient, clientKey)
	certs.AssertCertificate(t, rtConfiguration.CertificateSigningRequestInfo.Subject, rtCertResult)
	defer certs.Cleanup(t, configmapCleaner, rtCertResult)

	// Register second Application
	secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-second-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}, cfg.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C)},
	})
	defer fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, &secondApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.CertSecuredGraphqlClient, appsForRuntimeTenantID, secondApp.ID)
	require.NoError(t, err)
	require.NotEmpty(t, secondApp.ID)

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
