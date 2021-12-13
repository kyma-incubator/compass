package tests

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	ScenariosLabel  = "scenarios"
	TestScenario    = "test-scenario"
	DefaultScenario = "DEFAULT"
)

func TestAppsForRuntimeWithCertificates(t *testing.T) {
	scenarios := []string{DefaultScenario, TestScenario}
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, scenarios)

	appIdToCommonName := make(map[string]string)

	// Register first Application
	firstApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-first-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}},
	})
	defer func() {
		fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, firstApp.ID, true)
		fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, &firstApp)
	}()
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
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, &graphql.RuntimeInput{
		Name: "test-runtime",
		Labels: map[string]interface{}{
			ScenariosLabel:         []string{TestScenario},
			"global_subaccount_id": runtimeSubaccountTenantID,
		},
	})
	defer fixtures.CleanupRuntime(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, &runtime)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	// Register second Application
	secondApp, err := fixtures.RegisterApplicationFromInput(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, graphql.ApplicationRegisterInput{
		Name:   "test-second-app",
		Labels: map[string]interface{}{ScenariosLabel: []string{TestScenario}},
	})
	defer func() {
		fixtures.UnassignApplicationFromScenarios(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, secondApp.ID, true)
		fixtures.CleanupApplication(t, ctx, directorAppsForRuntimeClient.DexGraphqlClient, appsForRuntimeTenantID, &secondApp)
	}()
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

	testCases := []struct {
		name      string
		getClient func() *gcli.Client
	}{
		{
			name: "successful 'ApplicationsForRuntime' call with client certificate issued by Connector",
			getClient: func() *gcli.Client {
				ott, err := directorAppsForRuntimeClient.GenerateRuntimeToken(t, runtime.ID)
				require.NoError(t, err)
				rtCertResult, rtConfiguration := clients.GenerateRuntimeCertificate(t, &ott, connectorClient, clientKey)
				certs.AssertCertificate(t, rtConfiguration.CertificateSigningRequestInfo.Subject, rtCertResult)
				defer certs.Cleanup(t, configmapCleaner, rtCertResult)

				rtCertChain := certs.DecodeCertChain(t, rtCertResult.CertificateChain)
				securedClient := clients.NewCertificateSecuredConnectorClient(cfg.DirectorMtlsURL, clientKey, rtCertChain...)
				return securedClient.GraphQlClient
			},
		},
		{
			name: "successful 'ApplicationsForRuntime' call with externally issued client certificate",
			getClient: func() *gcli.Client {
				clientKey, rawCertChain := certs.ClientCertPair(t, cfg.ExternallyIssuedCert.Certificate, cfg.ExternallyIssuedCert.Key)
				directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(cfg.DirectorExternalCertSecuredURL, clientKey, rawCertChain)
				return directorCertSecuredClient
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			client := test.getClient()
			apps, err := fixtures.AppsForRuntime(ctx, client, appsForRuntimeTenantID, runtime.ID)
			require.NoError(t, err)
			require.Equal(t, 2, len(apps.Data))
			for _, app := range apps.Data {
				certCN := appIdToCommonName[app.ID]
				require.NotEmpty(t, app.Auths)
				systemAuthID := app.Auths[0].ID
				require.Equal(t, certCN, systemAuthID)
			}
		})
	}
}
