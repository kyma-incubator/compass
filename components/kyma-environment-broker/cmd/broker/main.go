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
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/deprovisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/handlers"
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
	EnableOnDemandVersion                bool `envconfig:"default=false"`
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
		"Kiali":     runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Jaeger":    runtime.NewGenericComponentDisabler("jaeger", "kyma-system"),
		"BackupInt": runtime.NewGenericComponentDisabler("backup-init", "kyma-system"),
		"Backup":    runtime.NewGenericComponentDisabler("backup", "kyma-system"),
		// TODO(workaround until #1049): following components should be always disabled and user should not be able to enable them in provisioning request. This implies following components cannot be specified under the plan schema definition.
		"KnativeProvisionerNatss": runtime.NewGenericComponentDisabler("knative-provisioner-natss", "knative-eventing"),
		"NatssStreaming":          runtime.NewGenericComponentDisabler("nats-streaming", "natss"),
	}
	optComponentsSvc := runtime.NewOptionalComponentsService(optionalComponentsDisablers)

	runtimeProvider := runtime.NewComponentsListProvider(cfg.ManagedRuntimeComponentsYAMLFilePath)

	gardenerClusterConfig, err := gardener.NewGardenerClusterConfig(cfg.Gardener.KubeconfigPath)
	fatalOnError(err)
	//
	gardenerSecrets, err := gardener.NewGardenerSecretsInterface(gardenerClusterConfig, cfg.Gardener.Project)
	fatalOnError(err)

	gardenerAccountPool := hyperscaler.NewAccountPool(gardenerSecrets)
	accountProvider := hyperscaler.NewAccountProvider(nil, gardenerAccountPool)

	inputFactory, err := input.NewInputBuilderFactory(optComponentsSvc, runtimeProvider, cfg.Provisioning, cfg.KymaVersion)
	fatalOnError(err)

	avsDel := avs.NewDelegator(cfg.Avs, db.Operations())
	externalEvalCreator := provisioning.NewExternalEvalCreator(cfg.Avs, avsDel, cfg.Avs.Disabled)
	// setup operation managers
	provisionManager := provisioning.NewManager(db.Operations(), logs)
	deprovisionManager := deprovisioning.NewManager(db.Operations(), logs)

	// define steps
	provisioningInit := provisioning.NewInitialisationStep(db.Operations(), db.Instances(), provisionerClient, directorClient, inputFactory, externalEvalCreator)
	provisionManager.InitStep(provisioningInit)

	provisioningSteps := []struct {
		disabled bool
		weight   int
		step     provisioning.Step
	}{
		{
			weight: 1,
			step:   provisioning.NewResolveCredentialsStep(db.Operations(), accountProvider),
		},
		{
			weight:   1,
			step:     provisioning.NewInternalEvaluationStep(cfg.Avs, avsDel),
			disabled: cfg.Avs.Disabled,
		},
		{
			weight:   1,
			step:     provisioning.NewProvideLmsTenantStep(lmsTenantManager, db.Operations()),
			disabled: cfg.LMS.Disabled,
		},
		{
			weight: 2,
			step:   provisioning.NewProvisionAzureEventHubStep(db.Operations(), azure.NewAzureProvider(), accountProvider, ctx),
		},
		{
			weight: 2,
			step:   provisioning.NewOverridesFromSecretsAndConfigStep(ctx, cli, db.Operations()),
		},
		{
			weight: 2,
			step:   provisioning.NewServiceManagerOverridesStep(db.Operations(), cfg.ServiceManager),
		},
		{
			weight:   4,
			step:     provisioning.NewLmsCertificatesStep(lmsClient, db.Operations()),
			disabled: cfg.LMS.Disabled,
		},
		{
			weight: 10,
			step:   provisioning.NewCreateRuntimeStep(db.Operations(), db.Instances(), provisionerClient),
		},
	}
	for _, step := range provisioningSteps {
		if !step.disabled {
			provisionManager.AddStep(step.weight, step.step)
		}
	}

	deprovisioningInit := deprovisioning.NewInitialisationStep(db.Operations(), db.Instances(), provisionerClient)
	deprovisionManager.InitStep(deprovisioningInit)
	deprovisioningSteps := []struct {
		weight int
		step   deprovisioning.Step
	}{
		{
			weight: 10,
			step:   deprovisioning.NewRemoveRuntimeStep(db.Operations(), db.Instances(), provisionerClient),
		},
	}
	for _, step := range deprovisioningSteps {
		deprovisionManager.AddStep(step.weight, step.step)
	}

	// run queues
	provisionQueue := process.NewQueue(provisionManager)
	provisionQueue.Run(ctx.Done())

	deprovisionQueue := process.NewQueue(deprovisionManager)
	deprovisionQueue.Run(ctx.Done())

	if !cfg.DisableProcessOperationsInProgress {
		err = processOperationsInProgressByType(dbmodel.OperationTypeProvision, db.Operations(), provisionQueue, logs)
		fatalOnError(err)
		err = processOperationsInProgressByType(dbmodel.OperationTypeDeprovision, db.Operations(), deprovisionQueue, logs)
		fatalOnError(err)
	} else {
		logger.Info("Skipping processing operation in progress on start")
	}

	plansValidator, err := broker.NewPlansSchemaValidator()
	fatalOnError(err)

	// create KymaEnvironmentBroker endpoints
	kymaEnvBroker := &broker.KymaEnvironmentBroker{
		broker.NewServices(cfg.Broker, optComponentsSvc, logs),
		broker.NewProvision(cfg.Broker, db.Operations(), db.Instances(), provisionQueue, inputFactory, plansValidator, cfg.EnableOnDemandVersion, logs),
		broker.NewDeprovision(db.Instances(), db.Operations(), deprovisionQueue, logs),
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
func processOperationsInProgressByType(opType dbmodel.OperationType, op storage.Operations, queue *process.Queue, log logrus.FieldLogger) error {
	operations, err := op.GetOperationsInProgressByType(opType)
	if err != nil {
		return errors.Wrap(err, "while getting in progress operations from storage")
	}
	for _, operation := range operations {
		queue.Add(operation.ID)
		log.Infof("Resuming the processing of %s operation ID: %s", opType, operation.ID)
	}
	return nil
}

func fatalOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
