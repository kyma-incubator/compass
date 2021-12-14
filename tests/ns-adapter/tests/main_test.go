package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/vrischmann/envconfig"
)


type ConnectorCAConfig struct {
	Certificate          []byte `envconfig:"-"`
	Key                  []byte `envconfig:"-"`
	SecretName           string
	SecretNamespace      string
	SecretCertificateKey string
	SecretKeyKey         string
}

type config struct {
	CA ConnectorCAConfig

	DefaultTestTenant string
	Domain            string
	DirectorURL       string
}

var (
	testConfig       config
	dexGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}
	testConfig.DirectorURL = fmt.Sprintf("https://compass-gateway-auth-oauth.%s/director/graphql", testConfig.Domain)

	tenant.TestTenants.Init()

	k8sClientSet, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}

	secret, err := k8sClientSet.CoreV1().Secrets(testConfig.CA.SecretNamespace).Get(ctx, testConfig.CA.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting k8s secret"))
	}

	testConfig.CA.Certificate = secret.Data[testConfig.CA.SecretCertificateKey]
	testConfig.CA.Key = secret.Data[testConfig.CA.SecretKeyKey]

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}