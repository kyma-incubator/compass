package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/clients"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	ctx := context.Background()
	k8sClientSet, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}

	extCrtSecret, err := k8sClientSet.CoreV1().Secrets(conf.ExternalCA.SecretNamespace).Get(ctx, conf.ExternalCA.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting k8s secret"))
	}

	conf.ExternalCA.Key = extCrtSecret.Data[conf.ExternalCA.SecretKeyKey]
	conf.ExternalCA.Certificate = extCrtSecret.Data[conf.ExternalCA.SecretCertificateKey]

	exitVal := m.Run()

	os.Exit(exitVal)
}
