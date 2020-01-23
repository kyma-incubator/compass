package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/pivotal-cf/brokerapi"
	"github.com/vrischmann/envconfig"
)

// Config holds configuration for the whole application
type Config struct {
	Auth struct {
		Username string
		Password string
	}
	Host string `envconfig:"optional"`
	Port string `envconfig:"default=8080"`

	Provisioning broker.ProvisioningConfig
	Director     director.Config
	Database     storage.Config

	ServiceManager struct {
		URL      string
		Password string
		Username string
	}
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err)

	logger := lager.NewLogger("kyma-env-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Info("Starting Kyma Environment Broker")

	brokerCredentials := brokerapi.BrokerCredentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	provisionerClient := provisioner.NewProvisionerClientV2(cfg.Provisioning.URL, cfg.ServiceManager.URL, cfg.ServiceManager.Username, cfg.ServiceManager.Password, true)

	k8sCfg, err := config.GetConfig()
	fatalOnError(err)
	cli, err := client.New(k8sCfg, client.Options{})
	fatalOnError(err)

	oauthClient := oauth.NewOauthClient(newHTTPClient(cfg.Director.SkipCertVerification), cli, cfg.Director.OauthCredentialsSecretName, cfg.Director.Namespace)
	fatalOnError(oauthClient.WaitForCredentials())

	director.NewDirectorClient(oauthClient)

	db, err := storage.New(cfg.Database.ConnectionURL())
	fatalOnError(err)

	kymaEnvBroker, err := broker.NewBroker(provisionerClient, cfg.Provisioning, db.Instances())
	fatalOnError(err)

	brokerAPI := brokerapi.New(kymaEnvBroker, logger, brokerCredentials)
	r := handlers.LoggingHandler(os.Stdout, brokerAPI)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

func newHTTPClient(skipCertVeryfication bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVeryfication},
		},
		Timeout: 30 * time.Second,
	}
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
