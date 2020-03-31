package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/gardener"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/health"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/http_client"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
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

	DbInMemory bool `envconfig:"default=false"`

	// DisableProcessOperationsInProgress allows to disable processing operations
	// which are in progress on starting application. Set to true if you are
	// running in a separate testing deployment but with the production DB.
	DisableProcessOperationsInProgress bool `envconfig:"default=false"`

	Host       string `envconfig:"optional"`
	Port       string `envconfig:"default=8080"`
	StatusPort string `envconfig:"default=8071"`

	Provisioning input.Config
	Director     director.Config
	Database     storage.Config
	Gardener     gardener.Config

	ServiceManager provisioning.ServiceManagerOverrideConfig

	KymaVersion                          string
	ManagedRuntimeComponentsYAMLFilePath string

	Broker broker.Config

	Avs avs.Config

	LMS lms.Config
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

	logs := logrus.New()

	logger.Info("Registering healthz endpoint for health probes")
	health.NewServer(cfg.Host, cfg.StatusPort, logs).ServeAsync()

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
	var db storage.BrokerStorage
	if cfg.DbInMemory {
		db = storage.NewMemoryStorage()
	} else {
		db, err = storage.NewFromConfig(cfg.Database, logs)
		fatalOnError(err)
	}

	// LMS
	fatalOnError(cfg.LMS.Validate())
	lmsClient := lms.NewClient(cfg.LMS, logs)
	lmsTenantManager := lms.NewTenantManager(db.LMSTenants(), lmsClient, logs.WithField("service", "lmsTenantManager"))

	// Register disabler. Convention:
	// {component-name} : {component-disabler-service}
	//
	// Using map is intentional - we ensure that component name is not duplicated.
	optionalComponentsDisablers := runtime.ComponentsDisablers{
		"Kiali":  runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger": runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		// TODO(workaround until #1049): following components should be always disabled and user should not be able to enable them in provisioning request. This implies following components cannot be specified under the plan schema definition.
		"KnativeProvisionerNatss": runtime.NewGenericComponentDisabler("knative-provisioner-natss", "knative-eventing"),
		"NatssStreaming":          runtime.NewGenericComponentDisabler("nats-streaming", "natss"),
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

	// create and run queue, steps provisioning
	initialisation := provisioning.NewInitialisationStep(db.Operations(), db.Instances(), provisionerClient, directorClient, inputFactory)

	resolveCredentialsStep := provisioning.NewResolveCredentialsStep(db.Operations(), accountProvider)
	evaluationStep := provisioning.NewInternalEvaluationStep(cfg.Avs, db.Operations())
	lmsProvideTenantStep := provisioning.NewProvideLmsTenantStep(lmsTenantManager, db.Operations())
	lmsCertStep := provisioning.NewLmsCertificatesStep(lmsClient, db.Operations())
	provisionAzureEventHub := provisioning.NewProvisionAzureEventHubStep(db.Operations(), azure.NewAzureProvider(), accountProvider, ctx)
	runtimeStep := provisioning.NewCreateRuntimeStep(db.Operations(), db.Instances(), provisionerClient)
	overridesStep := provisioning.NewOverridesFromSecretsAndConfigStep(ctx, cli, db.Operations())
	smOverrideStep := provisioning.NewServiceManagerOverridesStep(db.Operations(), cfg.ServiceManager)
	backupSetupStep := provisioning.NewSetupBackupStep(db.Operations())

	stepManager := process.NewManager(db.Operations(), logs)
	stepManager.InitStep(initialisation)

	stepManager.AddStep(1, resolveCredentialsStep)
	if !cfg.Avs.Disabled {
		stepManager.AddStep(1, evaluationStep)
	}
	if !cfg.LMS.Disabled {
		stepManager.AddStep(1, lmsProvideTenantStep)
		stepManager.AddStep(4, lmsCertStep) // must be just before runtimeStep and after lmsProvideTenantStep
	}
	stepManager.AddStep(2, provisionAzureEventHub)
	stepManager.AddStep(2, overridesStep)
	stepManager.AddStep(2, smOverrideStep)
	stepManager.AddStep(3, backupSetupStep)
	stepManager.AddStep(10, runtimeStep)

	queue := process.NewQueue(stepManager)
	queue.Run(ctx.Done())

	if !cfg.DisableProcessOperationsInProgress {
		err = processOperationsInProgress(db.Operations(), queue, logs)
		fatalOnError(err)
	} else {
		logger.Info("Skipping processing operation in progress on start")
	}

	plansValidator, err := broker.NewPlansSchemaValidator()
	fatalOnError(err)

	// create KymaEnvironmentBroker endpoints
	kymaEnvBroker := &broker.KymaEnvironmentBroker{
		broker.NewServices(cfg.Broker, optComponentsSvc, logs),
		broker.NewProvision(cfg.Broker, db.Operations(), queue, inputFactory, plansValidator, logs),
		broker.NewDeprovision(db.Instances(), provisionerClient, logs),
		broker.NewUpdate(logs),
		broker.NewGetInstance(db.Instances(), logs),
		broker.NewLastOperation(db.Operations(), logs),
		broker.NewBind(logs),
		broker.NewUnbind(logs),
		broker.NewGetBinding(logs),
		broker.NewLastBindingOperation(logs),
	}

	// create broker credentials
	brokerCredentials := broker.BrokerCredentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}

	// create and run broker OSB API in 2 modes:
	// with basic auth
	// with oauth
	brokerAPI := broker.New(kymaEnvBroker, logger, nil)
	brokerBasicAPI := broker.New(kymaEnvBroker, logger, &brokerCredentials)

	sm := http.NewServeMux()
	sm.Handle("/", brokerBasicAPI)
	sm.Handle("/oauth/", http.StripPrefix("/oauth", brokerAPI))

	r := handlers.LoggingHandler(os.Stdout, sm)

	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, r))
}

// queues all in progress provision operations existing in the database
func processOperationsInProgress(op storage.Operations, queue *process.Queue, log logrus.FieldLogger) error {
	operations, err := op.GetOperationsInProgressByType(dbsession.OperationTypeProvision)
	if err != nil {
		return errors.Wrap(err, "while getting in progress operations from storage")
	}
	for _, operation := range operations {
		queue.Add(operation.ID)
		log.Infof("Resuming the processing of operation ID: %s", operation.ID)
	}
	return nil
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
