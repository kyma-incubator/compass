package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/gardener"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
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

	Provisioning input.Config
	Director     director.Config
	Database     storage.Config
	Gardener     gardener.Config

	ManagementPlaneURL string

	ServiceManager provisioning.ServiceManagerOverrideConfig

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
		"Kiali":  runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger": runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
	}

	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)

	runtimeProvider := runtime.NewComponentsListProvider(cfg.KymaVersion, cfg.ManagedRuntimeComponentsYAMLFilePath)
	fullRuntimeComponentList, err := runtimeProvider.AllComponents()
	fatalOnError(err)

	gardenerClusterConfig, err := gardener.NewGardenerClusterConfig(cfg.Gardener.KubeconfigPath)
	fatalOnError(err)
	//
	gardenerSecrets, err := gardener.NewGardenerSecretsInterface(gardenerClusterConfig, cfg.Gardener.Project)
	fatalOnError(err)

	gardenerAccountPool := hyperscaler.NewAccountPool(gardenerSecrets)
	accountProvider := hyperscaler.NewAccountProvider(nil, gardenerAccountPool)

	inputFactory := input.NewInputBuilderFactory(optComponentsSvc, fullRuntimeComponentList, cfg.Provisioning, cfg.KymaVersion)

	// create log dumper
	dumper, err := broker.NewDumper()
	fatalOnError(err)

	// create and run queue, steps provisioning
	initialisation := provisioning.NewInitialisationStep(db.Operations(), db.Instances(), provisionerClient, directorClient, inputFactory, cfg.ManagementPlaneURL)
	resolveCredentialsStep := provisioning.NewResolveCredentialsStep(db.Operations(), accountProvider)
	runtimeStep := provisioning.NewCreateRuntimeStep(db.Operations(), db.Instances(), provisionerClient)
	smOverrideStep := provisioning.NewServiceManagerOverridesStep(db.Operations(), cfg.ServiceManager)
	backupSetupStep := provisioning.NewSetupBackupStep(db.Operations(), db.Instances(), provisionerClient, cfg.ServiceManager, accountProvider)

	runtimeStep := provisioning.NewCreateRuntimeStep(db.Operations(), db.Instances(), provisionerClient, cfg.ServiceManager)

	logs := logrus.New()
	stepManager := process.NewManager(db.Operations(), logs)
	stepManager.InitStep(initialisation)

	stepManager.AddStep(10, runtimeStep)
	stepManager.AddStep(2, smOverrideStep)
	stepManager.AddStep(1, resolveCredentialsStep)
	stepManager.AddStep(2, backupSetupStep)

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
