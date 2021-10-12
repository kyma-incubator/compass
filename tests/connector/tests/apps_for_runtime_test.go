package tests

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/require"
)

const (
	ScenariosLabel  = "scenarios"
	TestScenario    = "test-scenario"
	DefaultScenario = "DEFAULT"
)

func TestAppsForRuntimeWithCertificates(t *testing.T) {
	scenarios := []string{DefaultScenario, TestScenario}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, []string{DefaultScenario})

	appIdToCommonName := make(map[string]string)

	// Register first Application
	firstApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-first-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}},
	})
	defer func() {
		fixtures.CleanupApplication(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, &firstApp)
		fixtures.UnassignApplicationFromScenarios(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, firstApp.ID, true)
	}()
	require.NoError(t, err)
	require.NotEmpty(t, firstApp.ID)

	// Issue a certificate for the first Application
	certResult, configuration := clients.GenerateApplicationCertificate(t, directorClient, connectorClient, firstApp.ID, clientKey)
	certs.AssertCertificate(t, configuration.CertificateSigningRequestInfo.Subject, certResult)

	decodedClientCert := certs.DecodeCert(t, certResult.ClientCertificate)
	firstAppCommonName := decodedClientCert.Subject.CommonName
	require.NotEmpty(t, firstAppCommonName)
	appIdToCommonName[firstApp.ID] = firstAppCommonName
	defer certs.Cleanup(t, configmapCleaner, certResult)

	// Register Runtime
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, &graphql.RuntimeInput{
		Name:   "test-runtime",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}},
	})
	defer fixtures.CleanupRuntime(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	// Issue a certificate for the Runtime
	ott, err := directorClient.GenerateRuntimeToken(t, runtime.ID)
	require.NoError(t, err)
	rtCertResult, rtConfiguration := clients.GenerateRuntimeCertificate(t, &ott, connectorClient, clientKey)
	certs.AssertCertificate(t, rtConfiguration.CertificateSigningRequestInfo.Subject, rtCertResult)
	defer certs.Cleanup(t, configmapCleaner, rtCertResult)

	// Register second Application
	secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-second-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}},
	})
	defer func() {
		fixtures.CleanupApplication(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, &secondApp)
		fixtures.UnassignApplicationFromScenarios(t, ctx, directorClient.DexGraphqlClient, appsForRuntimeTenantID, secondApp.ID, true)
	}()
	require.NoError(t, err)
	require.NotEmpty(t, secondApp.ID)

	// Issue a certificate for the second Application
	secondCertResult, secondConfiguration := clients.GenerateApplicationCertificate(t, directorClient, connectorClient, secondApp.ID, clientKey)
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
