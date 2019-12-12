package main

import (
	"log"
	"net/http"
	"os"

	"github.tools.sap/gophers-team/kyma-environment-service-broker/internal/broker"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	"github.com/pivotal-cf/brokerapi"
	"github.com/vrischmann/envconfig"
	"github.tools.sap/gophers-team/kyma-environment-service-broker/internal/provisioner"
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

	dumper, err := broker.NewDumper()
	fatalOnError(err)

	provisionerClient := provisioner.NewProvisionerClient(cfg.Provisioning.URL, true)

	kymaBrokerService := &broker.KymaEnvBroker{
		Dumper:                      dumper,
		ProvisionerClient:           provisionerClient,

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
