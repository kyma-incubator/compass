package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	event_hub "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/http_client"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	gcli "github.com/machinebox/graphql"
	"github.com/pivotal-cf/brokerapi"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Config holds configuration for the whole application
type Config struct {
	Auth struct {
		Username string
		Password string
	}
	Host string `envconfig:"optional"`
	Port string `envconfig:"default=8080"`

	Provisioning       input.Config
	Director           director.Config
	Database           storage.Config
	ManagementPlaneURL string

	ServiceManager internal.ServiceManagerOverride

	KymaVersion                          string
	ManagedRuntimeComponentsYAMLFilePath string

	Broker broker.Config
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// create and fill config
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	fatalOnError(err)

	// create logger
	logger := lager.NewLogger("kyma-env-broker")
	logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.DEBUG))
	logger.RegisterSink(lager.NewWriterSink(os.Stderr, lager.ERROR))

	logger.Info("Starting Kyma Environment Broker")

	// create broker credentials
	brokerCredentials := brokerapi.BrokerCredentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	// create provisioner client
	provisionerClient := provisioner.NewProvisionerClient(cfg.Provisioning.URL, true)

	// create kubernetes client
	k8sCfg, err := config.GetConfig()
	fatalOnError(err)
	cli, err := client.New(k8sCfg, client.Options{})
	fatalOnError(err)

	// create director client on the base of graphQL client and OAuth client
	httpClient := http_client.NewHTTPClient(30, cfg.Director.SkipCertVerification)
	graphQLClient := gcli.NewClient(cfg.Director.URL, gcli.WithHTTPClient(httpClient))
	oauthClient := oauth.NewOauthClient(httpClient, cli, cfg.Director.OauthCredentialsSecretName, cfg.Director.Namespace)
	fatalOnError(oauthClient.WaitForCredentials())
	directorClient := director.NewDirectorClient(oauthClient, graphQLClient)

	// create storage
	db, err := storage.New(cfg.Database.ConnectionURL())
	fatalOnError(err)

	// Register disabler. Convention:
	// {component-name} : {component-disabler-service}
	//
	// Using map is intentional - we ensure that component name is not duplicated.
	optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Loki":   runtime.NewLokiDisabler(),
		"Kiali":  runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger": runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
	}

	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)

	runtimeProvider := runtime.NewComponentsListProvider(cfg.KymaVersion, cfg.ManagedRuntimeComponentsYAMLFilePath)
	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	fatalOnError(err)

	inputFactory := input.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, cfg.Provisioning, cfg.KymaVersion)

	// create log dumper
	dumper, err := broker.NewDumper()
	fatalOnError(err)

	// create and run queue, steps provisioning
	initialisation := provisioning.NewInitialisationStep(db.Operations(), db.Instances(), provisionerClient, directorClient, inputFactory, cfg.ManagementPlaneURL)

	secret := os.Getenv("AZURE_SECRET")
	azureConfig := azure.GetConfig("38c0ed1b-13d0-4936-8429-eccc80d2d8fb", secret, "42f7676c-f455-423c-82f6-dc2d99791af7", "35d42578-34d1-486d-a689-012a8d514c19")

	provisionAzureEventHub := event_hub.NewProvisionAzureEventHubStep(db.Operations(), azureConfig, ctx)
	runtimeStep := provisioning.NewCreateRuntimeStep(db.Operations(), db.Instances(), provisionerClient, cfg.ServiceManager)

	logs := logrus.New()
	stepManager := process.NewManager(db.Operations(), logs)
	stepManager.InitStep(initialisation)

	stepManager.AddStep(1, provisionAzureEventHub)
	stepManager.AddStep(2, runtimeStep)

	queue := process.NewQueue(stepManager)
	queue.Run(ctx.Done())

	plansValidator, err := broker.NewPlansSchemaValidator()
	fatalOnError(err)

	// create KymaEnvironmentBroker endpoints
	kymaEnvBroker := &broker.KymaEnvironmentBroker{
		broker.NewServices(cfg.Broker, optComponentsSvc, dumper),
		broker.NewProvision(cfg.Broker, db.Operations(), queue, inputFactory, plansValidator, dumper),
		broker.NewDeprovision(db.Instances(), provisionerClient, dumper),
		broker.NewUpdate(dumper),
		broker.NewGetInstance(db.Instances(), dumper),
		broker.NewLastOperation(db.Operations(), dumper),
		broker.NewBind(dumper),
		broker.NewUnbind(dumper),
		broker.NewGetBinding(dumper),
		broker.NewLastBindingOperation(dumper),
	}

	// create and run broker OSB API
	brokerAPI := brokerapi.New(kymaEnvBroker, logger, brokerCredentials)
	r := handlers.LoggingHandler(os.Stdout, brokerAPI)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
