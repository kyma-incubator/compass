package main

import (
	"log"
	"net/http"
	"os"

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
	Database     storage.Config

	// feature flag indicates whether use Provisioner API which returns RuntimeID
	ProcessRuntimeID bool `envconfig:"default=false"`
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

	var provisionerClient provisioner.Client
	if cfg.ProcessRuntimeID {
		provisionerClient = provisioner.NewProvisionerClientV2(cfg.Provisioning.URL, true)
	} else {
		provisionerClient = provisioner.NewProvisionerClient(cfg.Provisioning.URL, true)
	}

	db, err := storage.New(cfg.Database.ConnectionURL())
	fatalOnError(err)

	dumper, err := broker.NewDumper()
	fatalOnError(err)

	kymaBrokerService := &broker.KymaEnvBroker{
		Dumper:            dumper,
		ProvisionerClient: provisionerClient,
		Storage:           db,

		Config: cfg.Provisioning,
	}

	brokerAPI := brokerapi.New(kymaBrokerService, logger, brokerCredentials)
	r := handlers.LoggingHandler(os.Stdout, brokerAPI)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
