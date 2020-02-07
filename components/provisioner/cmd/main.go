package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

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
		Project        string `envconfig:"default=gardenerProject"`
		KubeconfigPath string `envconfig:"default=./dev/kubeconfig.yaml"`
	}

	Provisioner string `envconfig:"default=gardener"`
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s, CredentialsNamespace: %s, "+
		"DirectorURL: %s, SkipDirectorCertVerification: %v, OauthCredentialsSecretName: %s"+
		"DatabaseUser: %s, DatabaseHost: %s, DatabasePort: %s, "+
		"DatabaseName: %s, DatabaseSSLMode: %s, "+
		"GardenerProject: %s, GardenerKubeconfigPath: %s, Provisioner: %s",
		c.Address, c.APIEndpoint, c.CredentialsNamespace,
		c.DirectorURL, c.SkipDirectorCertVerification, c.OauthCredentialsSecretName,
		c.Database.User, c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.SSLMode,
		c.Gardener.Project, c.Gardener.KubeconfigPath, c.Provisioner)
}

func main() {
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load application config")

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

	dbsFactory := dbsession.NewFactory(connection)
	releaseRepository := release.NewReleaseRepository(connection, uuid.NewUUIDGenerator())

	installationService := installation.NewInstallationService(cfg.Installation.Timeout, installationSDK.NewKymaInstaller, cfg.Installation.ErrorsCountFailureThreshold)

	directorClient, err := newDirectorClient(cfg)
	exitOnError(err, "Failed to initialize Director client")

	runtimeConfigurator := runtime.NewRuntimeConfigurator(clientbuilder.NewConfigMapClientBuilder(), directorClient)

	var provisioner provisioning.Provisioner
	switch strings.ToLower(cfg.Provisioner) {
	case "hydroform":
		hydroformSvc := hydroform.NewHydroformService(client.NewHydroformClient(), cfg.Gardener.KubeconfigPath)
		provisioner = hydroform.NewHydroformProvisioner(hydroformSvc, installationService, dbsFactory, directorClient, runtimeConfigurator)
	case "gardener":
		provisioner = gardener.NewProvisioner(gardenerNamespace, shootClient)
		shootController, err := newShootController(cfg, gardenerNamespace, gardenerClusterConfig, gardenerClientSet, dbsFactory, installationService, directorClient, runtimeConfigurator)
		exitOnError(err, "Failed to create Shoot controller.")
		go func() {
			err := shootController.StartShootController()
			exitOnError(err, "Failed to start Shoot Controller")
		}()
	default:
		log.Fatalf("Error: invalid provisioner provided: %s", cfg.Provisioner)
	}

	provisioningSVC := newProvisioningService(cfg.Gardener.Project, provisioner, dbsFactory, releaseRepository, directorClient)
	validator := api.NewValidator(dbsFactory.NewReadSession())

	resolver := api.NewResolver(provisioningSVC, validator)

	httpClient := newHTTPClient(false)
	logger := log.WithField("Component", "Artifact Downloader")
	downloader := release.NewArtifactsDownloader(releaseRepository, 5, false, httpClient, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go downloader.FetchPeriodically(ctx, release.ShortInterval, release.LongInterval)

	gqlCfg := gqlschema.Config{
		Resolvers: resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	log.Infof("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.Use(middlewares.ExtractTenant)

	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	http.Handle("/", router)

	log.Infof("API listening on %s...", cfg.Address)

	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		panic(err)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
