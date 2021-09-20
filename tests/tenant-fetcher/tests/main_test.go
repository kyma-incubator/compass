package tests

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var dexGraphQLClient *graphql.Client
var httpClient *http.Client

type testConfig struct {
	TenantFetcherURL          string
	RootAPI                   string
	HandlerEndpoint           string
	RegionalHandlerEndpoint   string
	DependenciesEndpoint      string
	TenantPathParam           string
	RegionPathParam           string
	DbUser                    string
	DbPassword                string
	DbHost                    string
	DbPort                    string
	DbName                    string
	DbSSL                     string
	DbMaxIdleConnections      string
	DbMaxOpenConnections      string
	Tenant                    string
	SubscriptionCallbackScope string
	TenantProviderConfig
	ExternalServicesMockURL          string
	TenantFetcherFullURL             string `envconfig:"-"`
	TenantFetcherFullRegionalURL     string `envconfig:"-"`
	TenantFetcherFullDependenciesURL string `envconfig:"-"`
	SubscriptionProviderLabelKey     string `envconfig:"APP_SUBSCRIPTION_PROVIDER_LABEL_KEY"`
	ConsumerSubaccountIDsLabelKey    string `envconfig:"APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY"`
}

type TenantProviderConfig struct {
	TenantIDProperty               string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIDProperty     string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIDProperty             string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty              string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
	SubscriptionProviderIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY"`
}

var config testConfig

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&config, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	dexToken := server.Token()
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	httpClient = &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
	config.TenantFetcherFullURL = config.TenantFetcherURL + config.RootAPI + endpoint

	regionalEndpoint := strings.Replace(config.RegionalHandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", config.RegionPathParam), regionPathParamValue, 1)
	config.TenantFetcherFullRegionalURL = config.TenantFetcherURL + config.RootAPI + regionalEndpoint

	config.TenantFetcherFullDependenciesURL = config.TenantFetcherURL + config.RootAPI + config.DependenciesEndpoint

	exitVal := m.Run()
	os.Exit(exitVal)
}
