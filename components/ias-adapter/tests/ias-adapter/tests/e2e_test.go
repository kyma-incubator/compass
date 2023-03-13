package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ias"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
)

type tester struct {
	*testing.T
	ctx        context.Context
	iasService ias.Service
}

func TestEndToEnd(t *testing.T) {
	tester := tester{
		T:          t,
		ctx:        context.Background(),
		iasService: newIASService(t),
	}

	testCfg := tester.newTestConfig()

	gqlClient := tester.newGQLClient(testCfg)

	// 1. create formation
	tenantID := tenant.TestTenants.GetDefaultTenantID()
	formationName := newFormationName()
	fixtures.CreateFormation(t, tester.ctx, gqlClient, formationName)
	defer func() {
		fixtures.DeleteFormation(t, tester.ctx, gqlClient, formationName)
	}()

	// 2. assert consumer has no consumedApis
	tester.assertApplicationConsumedAPIs(testCfg.IASURL, testCfg.ConsumerAppClientID, 0)

	formationInput := graphql.FormationInput{Name: formationName}

	// 3. assign provider app in formation
	fixtures.AssignFormationWithApplicationObjectType(t, tester.ctx, gqlClient, formationInput, testCfg.ProviderAppID, tenantID)
	defer func() {
		fixtures.UnassignFormationWithApplicationObjectType(t, tester.ctx, gqlClient, formationInput, testCfg.ProviderAppID, tenantID)
	}()

	// 4. assert consumer app has no consumedApis
	tester.assertApplicationConsumedAPIs(testCfg.IASURL, testCfg.ConsumerAppClientID, 0)

	// 5. assign consumer app in formation
	fixtures.AssignFormationWithApplicationObjectType(t, tester.ctx, gqlClient, formationInput, testCfg.ConsumerAppID, tenantID)
	defer func() {
		fixtures.UnassignFormationWithApplicationObjectType(t, tester.ctx, gqlClient, formationInput, testCfg.ConsumerAppID, tenantID)
	}()

	// 6. assert consumer app has consumedApis
	tester.assertApplicationConsumedAPIs(testCfg.IASURL, testCfg.ConsumerAppClientID, 1)

	// 7. unassign consumer app from formation
	fixtures.UnassignFormationWithApplicationObjectType(t, tester.ctx, gqlClient, formationInput, testCfg.ConsumerAppID, tenantID)

	// 8. assert consumer app has no consumedApis
	tester.assertApplicationConsumedAPIs(testCfg.IASURL, testCfg.ConsumerAppClientID, 0)
}

func newIASService(t *testing.T) ias.Service {
	config, err := config.New()
	require.NoError(t, err)
	iasClient, err := ias.NewClient(config.IASConfig)
	require.NoError(t, err)
	return ias.NewService(config.IASConfig, iasClient)
}

func (t tester) newGQLClient(testCfg testConfig) *gcli.Client {
	certLoaderCache, err := certloader.StartCertLoader(t.ctx, testCfg.CertLoaderConfig)
	require.NoError(t, err)
	util.WaitForCache(certLoaderCache)
	require.NoError(t, err)
	return gql.NewCertAuthorizedGraphQLClientWithCustomURL(
		testCfg.DirectorExternalCertSecuredURL,
		certLoaderCache.Get()[testCfg.ExternalClientCertSecretName].PrivateKey,
		certLoaderCache.Get()[testCfg.ExternalClientCertSecretName].Certificate,
		testCfg.SkipSSLValidation,
	)
}

func newFormationName() string {
	return fmt.Sprintf("ias-adapter-e2e-tests-%d", time.Now().Unix())
}

func (t tester) assertApplicationConsumedAPIs(iasURL, clientID string, expected int) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	app, err := t.iasService.GetApplication(ctx, iasURL, clientID)
	require.NoError(t, err)
	require.Equal(t, len(app.Authentication.ConsumedAPIs), expected)
}

type testConfig struct {
	IASURL                         string `envconfig:"APP_TEST_IAS_URL"`
	ProviderAppID                  string `envconfig:"APP_TEST_PROVIDER_APP_ID"`
	ConsumerAppID                  string `envconfig:"APP_TEST_CONSUMER_APP_ID"`
	ConsumerAppClientID            string `envconfig:"APP_TEST_CONSUMER_APP_CLIENT_ID"`
	DirectorExternalCertSecuredURL string `envconfig:"default=http://compass-director-external-mtls.compass-system.svc.cluster.local:3000/graphql"`
	ExternalClientCertSecretName   string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	SkipSSLValidation              bool   `envconfig:"default=true"`
	CertLoaderConfig               certloader.Config
}

func (t tester) newTestConfig() testConfig {
	cfg := testConfig{}
	require.NoError(t, envconfig.InitWithPrefix(&cfg, "APP_TEST"))
	return cfg
}
