package test

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director/oauth"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/runtime"
	gcli "github.com/machinebox/graphql"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/pkg/client/broker"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/vrischmann/envconfig"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	Broker   broker.Config
	Runtime  runtime.Config
	Director director.Config

	TenantID             string `default:"3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	SkipCertVerification bool   `envconfig:"default=true"`
}

// Suite provides set of clients able to provision and test Kyma runtime
type Suite struct {
	t *testing.T

	log              logrus.FieldLogger
	brokerClient     *broker.Client
	runtimeClient    *runtime.Client
	dashboardChecker *runtime.DashboardChecker
}

func (ts *Suite) TearDown() {
	ts.log.Info("Cleaning up...")
	err := ts.runtimeClient.EnsureUAAInstanceRemoved()
	assert.NoError(ts.t, err)
	operationID, err := ts.brokerClient.DeprovisionRuntime()
	require.NoError(ts.t, err)
	err = ts.brokerClient.AwaitOperationSucceeded(operationID)
	require.NoError(ts.t, err)
}

func newTestSuite(t *testing.T) *Suite {
	cfg := &Config{}
	err := envconfig.InitWithPrefix(cfg, "APP")
	require.NoError(t, err)

	httpClient := newHTTPClient(cfg.SkipCertVerification)

	log := logrus.New()
	instanceID := uuid.New().String()
	brokerClient := broker.NewClient(cfg.Broker, cfg.TenantID, instanceID, *httpClient, log.WithField("service", "broker_client"))

	// create director client on the base of graphQL client and OAuth client
	graphQLClient := gcli.NewClient(cfg.Director.URL, gcli.WithHTTPClient(httpClient))
	graphQLClient.Log = func(s string) { log.Println(s) }

	k8sConfig, err := config.GetConfig()
	if err != nil {
		panic(err)
	}
	cli, err := k8sClient.New(k8sConfig, k8sClient.Options{})

	oauthClient := oauth.NewOauthClient(httpClient, cli, cfg.Director.OauthCredentialsSecretName, cfg.Director.Namespace)
	err = oauthClient.WaitForCredentials()
	if err != nil {
		panic(err)
	}

	directorClient := director.NewDirectorClient(oauthClient, graphQLClient)

	runtimeClient := runtime.NewClient(cfg.Runtime, cfg.TenantID, instanceID, *httpClient, *directorClient, log.WithField("service", "runtime_client"))

	dashboardChecker := runtime.NewDashboardChecker(*httpClient, log.WithField("service", "dashboard_checker"))

	return &Suite{
		t:                t,
		log:              log,
		dashboardChecker: dashboardChecker,
		brokerClient:     brokerClient,
		runtimeClient:    runtimeClient,
	}
}

func newHTTPClient(insecureSkipVerify bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
		},
	}
}
