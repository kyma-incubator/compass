package tests

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type testConfig struct {
	TenantFetcherURL          string
	RootAPI                   string
	HandlerEndpoint           string
	HandlerRegionalEndpoint   string
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
	DirectorUrl               string
	SubscriptionCallbackScope string
	TenantProviderConfig
}

type TenantProviderConfig struct {
	TenantIdProperty           string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIdProperty string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIdProperty         string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty          string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
}

var config testConfig

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&config, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	exitVal := m.Run()
	os.Exit(exitVal)
}
