package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/appinfo"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/director/oauth"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/edp"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/gardener"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/health"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/httputil"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/lms"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/middleware"
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
	"github.com/gorilla/mux"
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

	// DevelopmentMode if set to true then errors are returned in http
	// responses, otherwise errors are only logged and generic message
	// is returned to client.
	// Currently works only with /info endpoints.
	DevelopmentMode bool `envconfig:"default=false"`

	// DumpProvisionerRequests enables dumping Provisioner requests. Must be disabled on Production environments
	// because some data must not be visible in the log file.
	DumpProvisionerRequests bool `envconfig:"default=false"`

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
	DefaultRequestRegion                 string `envconfig:"default=cf-eu10"`

	Broker broker.Config

	Avs avs.Config
	LMS lms.Config
	IAS ias.Config
	EDP edp.Config
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
	provisionerClient := provisioner.NewProvisionerClient(cfg.Provisioning.URL, cfg.DumpProvisionerRequests)

	// create kubernetes client
	k8sCfg, err := config.GetConfig()
	fatalOnError(err)
	cli, err := client.New(k8sCfg, client.Options{})
	fatalOnError(err)

	// create director client on the base of graphQL client and OAuth client
	httpClient := httputil.NewClient(30, cfg.Director.SkipCertVerification)
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
		"Kiali":   runtime.NewGenericComponentDisabler("kiali", "kyma-system"),
		"Tracing": runtime.NewGenericComponentDisabler("tracing", "kyma-system"),
		// TODO(workaround until #1049): following components should be always disabled and user should not be able to enable them in provisioning request. This implies following components cannot be specified under the plan schema definition.
		"BackupInt":               runtime.NewGenericComponentDisabler("backup-init", "kyma-system"),
		"Backup":                  runtime.NewGenericComponentDisabler("backup", "kyma-system"),
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

	edpClient := edp.NewClient(cfg.EDP, logs)

	avsDel := avs.NewDelegator(cfg.Avs, db.Operations())
	externalEvalAssistant := avs.NewExternalEvalAssistant(cfg.Avs)
	internalEvalAssistant := avs.NewInternalEvalAssistant(cfg.Avs)
	externalEvalCreator := provisioning.NewExternalEvalCreator(cfg.Avs, avsDel, cfg.Avs.Disabled, externalEvalAssistant)

	bundleBuilder := ias.NewBundleBuilder(httpClient, cfg.IAS)
	iasTypeSetter := provisioning.NewIASType(bundleBuilder, cfg.IAS.Disabled)

	// setup operation managers
	provisionManager := provisioning.NewManager(db.Operations(), logs)
	deprovisionManager := deprovisioning.NewManager(db.Operations(), logs)

	// define steps
	provisioningInit := provisioning.NewInitialisationStep(db.Operations(), db.Instances(),
		provisionerClient, directorClient, inputFactory, externalEvalCreator, iasTypeSetter, cfg.Provisioning.Timeout)
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
			step:     provisioning.NewInternalEvaluationStep(cfg.Avs, avsDel, internalEvalAssistant),
			disabled: cfg.Avs.Disabled,
		},
		{
			weight:   1,
			step:     provisioning.NewProvideLmsTenantStep(lmsTenantManager, db.Operations(), cfg.LMS.Region),
			disabled: cfg.LMS.Disabled,
		},
		{
			weight:   1,
			step:     provisioning.NewEDPRegistrationStep(db.Operations(), edpClient, cfg.EDP),
			disabled: cfg.EDP.Disabled,
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
			weight:   5,
			step:     provisioning.NewIASRegistrationStep(db.Operations(), bundleBuilder),
			disabled: cfg.IAS.Disabled,
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
		disabled bool
		weight   int
		step     deprovisioning.Step
	}{
		{
			weight: 1,
			step:   deprovisioning.NewAvsEvaluationsRemovalStep(avsDel, db.Operations(), externalEvalAssistant, internalEvalAssistant),
		},
		{
			weight:   1,
			step:     deprovisioning.NewEDPDeregistrationStep(edpClient, cfg.EDP),
			disabled: cfg.EDP.Disabled,
		},
		{
			weight:   1,
			step:     deprovisioning.NewIASDeregistrationStep(db.Operations(), bundleBuilder),
			disabled: cfg.IAS.Disabled,
		},
		{
			weight: 10,
			step:   deprovisioning.NewRemoveRuntimeStep(db.Operations(), db.Instances(), provisionerClient),
		},
	}
	for _, step := range deprovisioningSteps {
		if !step.disabled {
			deprovisionManager.AddStep(step.weight, step.step)
		}
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

	// create server
	router := mux.NewRouter()

	// create info endpoints
	respWriter := httputil.NewResponseWriter(logs, cfg.DevelopmentMode)
	runtimesInfoHandler := appinfo.NewRuntimeInfoHandler(db.Instances(), respWriter)
	router.Handle("/info/runtimes", runtimesInfoHandler)

	// create OSB API endpoints
	basicAuth := &broker.Credentials{
		Username: cfg.Auth.Username,
		Password: cfg.Auth.Password,
	}
	router.Use(middleware.AddRegionToContext(cfg.DefaultRequestRegion))
	for prefix, creds := range map[string]*broker.Credentials{
		"/":                basicAuth, // legacy basic auth
		"/oauth/":          nil,       // oauth2 handled by Ory
		"/oauth/{region}/": nil,       // oauth2 handled by Ory with region
	} {
		route := router.PathPrefix(prefix).Subrouter()
		broker.AttachRoutes(route, kymaEnvBroker, logger, creds)
	}

	svr := handlers.LoggingHandler(os.Stdout, router)
	fatalOnError(http.ListenAndServe(cfg.Host+":"+cfg.Port, svr))
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
