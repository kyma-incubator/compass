package main

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/open_discovery/puller"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"net/http"
	"time"
)

type config struct {
	Database persistence.DatabaseConfig
	Features features.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	stopCh := signal.SetupChannel()
	cfgProvider := createAndRunConfigProvider(stopCh, cfg)

	pullerSvc := createODPullerSvc(cfgProvider, cfg.Features, transact)
	ctx := context.Background()
	err = pullerSvc.SyncODDocuments(ctx)
	exitOnError(err, "Error while synchronizing open-discovery documents")

	log.Info("Successfully synchronized open-discovery documents")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func createODPullerSvc(cfgProvider *configprovider.Provider, featuresConfig features.Config, transact persistence.Transactioner) *puller.Service {
	httpClient := getHttpClient()

	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	labelConverter := label.NewConverter()
	labelDefConverter := labeldef.NewConverter()
	intSysConverter := integrationsystem.NewConverter()
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(frConverter, versionConverter)
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	specConverter := spec.NewConverter()

	webhookConverter := webhook.NewConverter(authConverter)
	bundleConverter := mp_bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	packageConverter := mp_package.NewConverter(bundleConverter)

	runtimeRepo := runtime.NewRepository()
	specRepo := spec.NewRepository(specConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	apiRepo := api.NewRepository(apiConverter, specRepo)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)

	applicationRepo := application.NewRepository(appConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	packageRepo := mp_package.NewRepository(packageConverter)
	bundleRepo := mp_bundle.NewRepository(bundleConverter)

	uidSvc := uid.NewService()
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewScenariosService(labelDefRepo, uidSvc, featuresConfig.DefaultScenarioEnabled)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, log.StandardLogger())
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	apiSvc := api.NewService(apiRepo, fetchRequestRepo, uidSvc, fetchRequestSvc, specSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, fetchRequestRepo, uidSvc)

	bundleSvc := mp_bundle.NewService(bundleRepo, apiSvc, eventAPISvc, docRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	appSvc := application.NewService(cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelUpsertSvc, scenariosSvc, bundleSvc, uidSvc)
	webhookSvc := webhook.NewService(webhookRepo, uidSvc)
	packageSvc := mp_package.NewService(packageRepo, bundleRepo, uidSvc, bundleSvc)

	return puller.NewService(transact, appSvc, webhookSvc, bundleSvc, packageSvc)
}

func getHttpClient() *http.Client {
	out := &http.Client{
		Timeout: time.Second * 3,
	}
	return out
}

func createAndRunConfigProvider(stopCh <-chan struct{}, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	err := provider.Load()
	exitOnError(err, "Error on loading configuration file")
	executor.NewPeriodic(cfg.ConfigurationFileReload, func(stopCh <-chan struct{}) {
		if err := provider.Load(); err != nil {
			exitOnError(err, "Error from Reloader watch")
		}
		log.Infof("Successfully reloaded configuration file")

	}).Run(stopCh)

	return provider
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}
