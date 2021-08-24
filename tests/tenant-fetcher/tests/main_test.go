package tests

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
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
	TenantPathParam           string
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
	TenantProvider            string

	TenantFetcherFullURL string `envconfig:"-"`
}

var config testConfig

func TestMain(m *testing.M) {
	err := envconfig.InitWithPrefix(&config, "APP")
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	dexToken := server.Token()
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	httpClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: httputil.NewServiceAccountTokenTransport(http.DefaultTransport),
	}

	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
	config.TenantFetcherFullURL = config.TenantFetcherURL + config.RootAPI + endpoint

	exitVal := m.Run()
	os.Exit(exitVal)
}
