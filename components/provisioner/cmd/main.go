package main

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/steps"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"k8s.io/client-go/rest"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/healthz"

	"github.com/kyma-incubator/compass/components/provisioner/internal/api/middlewares"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime/clientbuilder"

	"github.com/kyma-incubator/compass/components/provisioner/internal/api"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/gardener"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const connStringFormat string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type config struct {
	Address                      string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint                  string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint        string `envconfig:"default=/graphql"`
	CredentialsNamespace         string `envconfig:"default=compass-system"`
	DirectorURL                  string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	SkipDirectorCertVerification bool   `envconfig:"default=false"`
	OauthCredentialsSecretName   string `envconfig:"default=compass-provisioner-credentials"`
	DownloadPreReleases          bool   `envconfig:"default=true"`

	Database struct {
		User     string `envconfig:"default=postgres"`
		Password string `envconfig:"default=password"`
		Host     string `envconfig:"default=localhost"`
		Port     string `envconfig:"default=5432"`
		Name     string `envconfig:"default=provisioner"`
		SSLMode  string `envconfig:"default=disable"`
	}

	Installation struct {
		Timeout                     time.Duration `envconfig:"default=40m"`
		ErrorsCountFailureThreshold int           `envconfig:"default=5"`
	}

	Gardener struct {
		Project                  string `envconfig:"default=gardenerProject"`
		KubeconfigPath           string `envconfig:"default=./dev/kubeconfig.yaml"`
		AuditLogsPolicyConfigMap string `envconfig:"optional"`
		AuditLogsTenant          string `envconfig:"optional"`
	}

	Provisioner             string `envconfig:"default=gardener"`
	SupportOnDemandReleases bool   `envconfig:"default=false"`

	LogLevel string `envconfig:"default=info"`
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s, CredentialsNamespace: %s, "+
		"DirectorURL: %s, SkipDirectorCertVerification: %v, OauthCredentialsSecretName: %s, DownloadPreReleases: %v, "+
		"DatabaseUser: %s, DatabaseHost: %s, DatabasePort: %s, "+
		"DatabaseName: %s, DatabaseSSLMode: %s, "+
		"GardenerProject: %s, GardenerKubeconfigPath: %s, GardenerAuditLogsPolicyConfigMap: %s, GardenerAuditLogsTenant: %s, "+
		"Provisioner: %s, SupportOnDemandReleases: %v, "+
		"LogLevel: %s",
		c.Address, c.APIEndpoint, c.CredentialsNamespace,
		c.DirectorURL, c.SkipDirectorCertVerification, c.OauthCredentialsSecretName, c.DownloadPreReleases,
		c.Database.User, c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.SSLMode,
		c.Gardener.Project, c.Gardener.KubeconfigPath, c.Gardener.AuditLogsPolicyConfigMap, c.Gardener.AuditLogsTenant,
		c.Provisioner, c.SupportOnDemandReleases,
		c.LogLevel)
}

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load application config")

	logLevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warnf("Invalid log level: '%s', defaulting to 'info'", cfg.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)

	log.Infof("Starting Provisioner")
	log.Infof("Config: %s", cfg.String())

	connString := fmt.Sprintf(connStringFormat, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	gardenerNamespace := fmt.Sprintf("garden-%s", cfg.Gardener.Project)

	gardenerClusterConfig, err := newGardenerClusterConfig(cfg)
	exitOnError(err, "Failed to initialize Gardener cluster client")

	gardenerClientSet, err := gardener.NewClient(gardenerClusterConfig)
	exitOnError(err, "Failed to create Gardener cluster clientset")

	shootClient := gardenerClientSet.Shoots(gardenerNamespace)

	connection, err := database.InitializeDatabaseConnection(connString, databaseConnectionRetries)
	exitOnError(err, "Failed to initialize persistence")

	// TODO: temporary solution
	installationHandlerConstructor := func(c *rest.Config, o ...installationSDK.InstallationOption) (installationSDK.Installer, error) {
		return installationSDK.NewKymaInstaller(c, o...)
	}

	dbsFactory := dbsession.NewFactory(connection)
	installationService := installation.NewInstallationService(cfg.Installation.Timeout, installationHandlerConstructor, cfg.Installation.ErrorsCountFailureThreshold)

	directorClient, err := newDirectorClient(cfg)
	exitOnError(err, "Failed to initialize Director client")

	runtimeConfigurator := runtime.NewRuntimeConfigurator(clientbuilder.NewConfigMapClientBuilder(), directorClient)

	installationQueue := createInstallationQueue(dbsFactory, installationService, runtimeConfigurator)
	upgradeQueue := createUpgradeQueue(dbsFactory, installationService)

	var provisioner provisioning.Provisioner
	switch strings.ToLower(cfg.Provisioner) {
	case "hydroform":
		hydroformSvc := hydroform.NewHydroformService(client.NewHydroformClient(), cfg.Gardener.KubeconfigPath)
		provisioner = hydroform.NewHydroformProvisioner(hydroformSvc, installationService, dbsFactory, directorClient, runtimeConfigurator)
	case "gardener":
		provisioner = gardener.NewProvisioner(gardenerNamespace, shootClient, cfg.Gardener.AuditLogsPolicyConfigMap, cfg.Gardener.AuditLogsTenant)
		shootController, err := newShootController(gardenerNamespace, gardenerClusterConfig, gardenerClientSet, dbsFactory, directorClient, installationQueue)
		exitOnError(err, "Failed to create Shoot controller.")
		go func() {
			err := shootController.StartShootController()
			exitOnError(err, "Failed to start Shoot Controller")
		}()
	default:
		log.Fatalf("Error: invalid provisioner provided: %s", cfg.Provisioner)
	}
	httpClient := newHTTPClient(false)

	releaseRepository := release.NewReleaseRepository(connection, uuid.NewUUIDGenerator())
	var releaseProvider release.Provider = releaseRepository
	if cfg.SupportOnDemandReleases {
		releaseProvider = release.NewOnDemandWrapper(httpClient, releaseRepository)
	}

	provisioningSVC := newProvisioningService(cfg.Gardener.Project, provisioner, dbsFactory, releaseProvider, directorClient)
	validator := api.NewValidator(dbsFactory.NewReadSession())

	resolver := api.NewResolver(provisioningSVC, validator)

	logger := log.WithField("Component", "Artifact Downloader")
	downloader := release.NewArtifactsDownloader(releaseRepository, 5, cfg.DownloadPreReleases, httpClient, logger)

	// Run release downloader
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go downloader.FetchPeriodically(ctx, release.ShortInterval, release.LongInterval)

	// Run installation queue
	installationQueue.Run(ctx.Done())
	// Run upgrade queue
	upgradeQueue.Run(ctx.Done())

	err = enqueueOperationsInProgress(dbsFactory, installationQueue, upgradeQueue)
	exitOnError(err, "Failed to enqueue in progress operations")

	gqlCfg := gqlschema.Config{
		Resolvers: resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	log.Infof("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.Use(middlewares.ExtractTenant)

	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))
	router.HandleFunc("/healthz", healthz.NewHTTPHandler(log.StandardLogger()))

	http.Handle("/", router)

	log.Infof("API listening on %s...", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		panic(err)
	}
}

func createInstallationQueue(factory dbsession.Factory, installationClient installation.Service, configurator runtime.Configurator) *installation.Queue {
	configureAgentStep := steps.NewConnectAgentStep(configurator, model.FinishedStep, 10*time.Minute)
	waitForInstallStep := steps.NewWaitForInstallationStep(installationClient, configureAgentStep.Name(), 50*time.Minute) // TODO: take form config
	installStep := steps.NewInstallKymaStep(installationClient, waitForInstallStep.Name(), 10*time.Minute)

	installSteps := map[model.OperationStage]installation.Step{
		model.ConnectRuntimeAgent:    configureAgentStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingInstallation:   installStep,
	}

	installationExecutor := steps.NewStepsExecutor(factory.NewReadWriteSession(), model.Provision, installSteps)

	return installation.NewQueue(installationExecutor)
}

func createUpgradeQueue(factory dbsession.Factory, installationClient installation.Service) *installation.Queue {

	// TODO: probably you will need some step for "committing" the changes to database

	waitForInstallStep := steps.NewWaitForInstallationStep(installationClient, model.FinishedStep, 50*time.Minute) // TODO: take form config
	installStep := steps.NewUpgradeKymaStep(installationClient, waitForInstallStep.Name(), 10*time.Minute)

	upgradeSteps := map[model.OperationStage]installation.Step{
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingInstallation:   installStep,
	}

	upgradeExecutor := steps.NewStepsExecutor(factory.NewReadWriteSession(), model.Upgrade, upgradeSteps)

	return installation.NewQueue(upgradeExecutor)
}

func enqueueOperationsInProgress(dbFactory dbsession.Factory, installationQueue, upgradeQueue installation.InstallationQueue) error {
	readSession := dbFactory.NewReadSession()

	inProgressOps, err := readSession.ListInProgressOperations()
	if err != nil {
		return err
	}

	for _, op := range inProgressOps {
		if op.Type == model.Provision && op.Stage != model.ShootProvisioning {
			installationQueue.Add(op.ID)
			continue
		}

		if op.Type == model.Upgrade {
			upgradeQueue.Add(op.ID)
		}
	}

	return nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
