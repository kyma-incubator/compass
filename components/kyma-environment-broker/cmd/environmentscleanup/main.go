package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/environmentscleanup"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/environmentscleanup/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/gardener"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/httputil"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8sClientCfg "sigs.k8s.io/controller-runtime/pkg/client/config"

	log "github.com/sirupsen/logrus"
)

type config struct {
	MaxAgeHours time.Duration `envconfig:"APP_MAX_AGE_HOURS"`
	Gardener    gardener.Config
	Director    director.Config
	Broker      broker.Config
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(errors.Wrap(err, "while loading cleanup config"))

	clusterCfg, err := gardener.NewGardenerClusterConfig(cfg.Gardener.KubeconfigPath)
	fatalOnError(errors.Wrap(err, "while creating Gardener cluster config"))
	cli, err := gardener.NewClient(clusterCfg)
	fatalOnError(errors.Wrap(err, "while creating Gardener client"))
	gardenerNamespace := fmt.Sprintf("garden-%s", cfg.Gardener.Project)
	shootClient := cli.Shoots(gardenerNamespace)

	k8sCfg, err := k8sClientCfg.GetConfig()
	fatalOnError(err)
	k8sCli, err := client.New(k8sCfg, client.Options{})
	fatalOnError(err)

	httpClient := httputil.NewClient(30, cfg.Director.SkipCertVerification)
	graphQLClient := gcli.NewClient(cfg.Director.URL, gcli.WithHTTPClient(httpClient))
	oauthClient := oauth.NewOauthClient(httpClient, k8sCli, cfg.Director.OauthCredentialsSecretName, cfg.Director.Namespace)
	err = oauthClient.WaitForCredentials()
	fatalOnError(err)
	directorClient := director.NewDirectorClient(oauthClient, graphQLClient)

	ctx := context.Background()
	brokerClient := broker.NewClient(ctx, cfg.Broker)

	svc := environmentscleanup.NewService(shootClient, directorClient, brokerClient, cfg.MaxAgeHours)
	svc.PerformCleanup()
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
