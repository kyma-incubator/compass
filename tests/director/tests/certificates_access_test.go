package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSystemAccess(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestDirectorCertificateAccess Integration System consumer: manage account tenant entities", func(t *testing.T) {
		ctx := context.Background()

		// defaultTenantId is the parent of the subaccountID
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
		//subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemSubaccount)

		// Register integration system
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, defaultTenantId, "self-registered-integration-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, dexGraphQLClient, defaultTenantId, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.ClientCertPair(t, conf.ExternalCA.Certificate, conf.ExternalCA.Key)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, clientKey, rawCertChain)

		applicationName := "int-sys-managed-app"
		intSysManagedTenant := tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemManagedAccount)

		t.Log(fmt.Sprintf("Trying to create applications in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		app, err := fixtures.RegisterApplication(t, ctx, directorCertSecuredClient, applicationName, intSysManagedTenant)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, intSysManagedTenant, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)
		require.Equal(t, applicationName, app.Name)

		t.Log(fmt.Sprintf("Trying to list applications in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		apps := fixtures.GetApplicationPage(t, ctx, directorCertSecuredClient, intSysManagedTenant)
		require.Len(t, apps.Data, 1)
		require.Equal(t, applicationName, apps.Data[0].Name)

		t.Log(fmt.Sprintf("Trying to register runtime in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		rt, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, intSysManagedTenant, &graphql.RuntimeInput{
			Name: "test-runtime",
		})
		require.NoError(t, err)
		require.NotEmpty(t, rt.ID)

		t.Log(fmt.Sprintf("Trying to create application template in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		at, err := fixtures.CreateApplicationTemplate(t, ctx, directorCertSecuredClient, defaultTenantId, "test-app-template")
		require.NoError(t, err)
		require.NotEmpty(t, at.ID)
	})

	t.Run("TestDirectorCertificateAccess Integration System consumer: cannot manage entities in non-managed tenant types", func(t *testing.T) {
		ctx := context.Background()

		// defaultTenantId is the parent of the subaccountID
		defaultTenantId := tenant.TestTenants.GetDefaultTenantID()
		subaccountID := tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemSubaccount)

		// Build graphql director client configured with certificate
		clientKey, rawCertChain := certs.IssueExternalIssuerCertificate(t, conf.ExternalCA.Certificate, conf.ExternalCA.Key, subaccountID)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, clientKey, rawCertChain)

		// Register integration system
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, defaultTenantId, "self-registered-integration-system")
		defer fixtures.CleanupIntegrationSystem(t, ctx, dexGraphQLClient, defaultTenantId, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		applicationName := "int-sys-managed-app"
		intSysManagedTenant := tenant.TestTenants.GetIDByName(t, tenant.TestIntegrationSystemManagedAccount)

		t.Log(fmt.Sprintf("Trying to create applications in subaccount tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		app, err := fixtures.RegisterApplication(t, ctx, directorCertSecuredClient, applicationName, intSysManagedTenant)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, intSysManagedTenant, &app)
		require.Error(t, err)

		t.Log(fmt.Sprintf("Trying to list applications in subaccount tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		getAppReq := fixtures.FixGetApplicationsRequestWithPagination()
		apps := graphql.ApplicationPage{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, intSysManagedTenant, getAppReq, &apps)
		require.Error(t, err)

		t.Log(fmt.Sprintf("Trying to register runtime in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		_, err = fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorCertSecuredClient, intSysManagedTenant, &graphql.RuntimeInput{
			Name: "test-runtime",
		})
		require.Error(t, err)

		t.Log(fmt.Sprintf("Trying to create application template in account tenant %s via client certificate of integration system with ID %s", intSysManagedTenant, intSys.ID))
		_, err = fixtures.CreateApplicationTemplate(t, ctx, directorCertSecuredClient, defaultTenantId, "test-app-template")
		require.Error(t, err)
	})
}
