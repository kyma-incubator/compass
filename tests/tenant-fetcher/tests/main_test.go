package tests

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

var certSecuredGraphQLClient *graphql.Client
var httpClient *http.Client

type testConfig struct {
	DirectorExternalCertSecuredURL string
	TenantFetcherURL               string
	RootAPI                        string
	RegionalHandlerEndpoint        string
	DependenciesEndpoint           string
	TenantPathParam                string
	RegionPathParam                string
	SubscriptionCallbackScope      string
	TenantProviderConfig
	ExternalServicesMockURL          string
	ClientID                         string
	ClientSecret                     string
	TenantFetcherFullRegionalURL     string `envconfig:"-"`
	TenantFetcherFullDependenciesURL string `envconfig:"-"`
	SkipSSLValidation                bool   `envconfig:"default=false"`
	SelfRegDistinguishLabelKey       string
	SelfRegDistinguishLabelValue     string
	SelfRegRegion                    string
	CertLoaderConfig                 certloader.Config
}

type TenantProviderConfig struct {
	TenantIDProperty               string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIDProperty     string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIDProperty             string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty              string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
	SubscriptionProviderIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY"`
	ProviderSubaccountIDProperty   string `envconfig:"APP_TENANT_PROVIDER_PROVIDER_SUBACCOUNT_ID_PROPERTY"`
	SubscriptionAppNameProperty    string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_APP_NAME_PROPERTY"`
}

var config testConfig

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&config, "APP")
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()
	cc, err := certloader.StartCertLoader(ctx, config.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(config.DirectorExternalCertSecuredURL, cc.Get().PrivateKey, cc.Get().Certificate, config.SkipSSLValidation)

	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	config.TenantFetcherFullRegionalURL = tenantfetcher.BuildTenantFetcherRegionalURL(config.RegionalHandlerEndpoint, config.TenantPathParam, config.RegionPathParam, config.TenantFetcherURL, config.RootAPI)

	config.TenantFetcherFullDependenciesURL = config.TenantFetcherURL + config.RootAPI + config.DependenciesEndpoint

	exitVal := m.Run()
	os.Exit(exitVal)
}
