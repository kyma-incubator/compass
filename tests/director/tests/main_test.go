package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
)

var (
	conf             = &config.DirectorConfig{}
	dexGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	config.ReadConfig(conf)

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	k8sClientSet, err := clients.NewK8SClientSet(context.Background(), time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}

	secret, err := k8sClientSet.CoreV1().Secrets(conf.CA.SecretNamespace).Get(context.Background(), conf.CA.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting k8s secret"))
	}

	conf.CA.Certificate = secret.Data[conf.CA.SecretCertificateKey]
	conf.CA.Key = secret.Data[conf.CA.SecretKeyKey]

	exitVal := m.Run()

	os.Exit(exitVal)
}
