package main

import (
	"context"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/adapter"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/handler"
	"github.com/kyma-incubator/compass/components/director/internal/nsadapter/httputil"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	directorHandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"net/http"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	conf := adapter.Configuration{}
	err := envconfig.Init(&conf)
	exitOnError(err, "while reading Pairing Adapter configuration")



	//
	//
	//cfgProvider := createAndRunConfigProvider(ctx, conf)
	//
	//
	//uidSvc := uid.NewService()
	//
	//tenantConv := tenant.NewConverter()
	//tenantRepo := tenant.NewRepository(tenantConv)
	//
	//authConverter := auth.NewConverter()
	//frConverter := fetchrequest.NewConverter(authConverter)
	//versionConverter := version.NewConverter()
	//docConverter := document.NewConverter(frConverter)
	//webhookConverter := webhook.NewConverter(authConverter)
	//specConverter := spec.NewConverter(frConverter)
	//apiConverter := api.NewConverter(versionConverter, specConverter)
	//eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	//labelDefConverter := labeldef.NewConverter()
	//labelConverter := label.NewConverter()
	//intSysConverter := integrationsystem.NewConverter()
	//bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	//appConverter := application.NewConverter(webhookConverter, bundleConverter)
	//runtimeConverter := runtime.NewConverter()
	//bundleReferenceConv := bundlereferences.NewConverter()
	//
	//runtimeRepo := runtime.NewRepository(runtimeConverter)
	//applicationRepo := application.NewRepository(appConverter)
	//labelRepo := label.NewRepository(labelConverter)
	//labelDefRepo := labeldef.NewRepository(labelDefConverter)
	//webhookRepo := webhook.NewRepository(webhookConverter)
	//apiRepo := api.NewRepository(apiConverter)
	//eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	//specRepo := spec.NewRepository(specConverter)
	//docRepo := document.NewRepository(docConverter)
	//fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	//intSysRepo := integrationsystem.NewRepository(intSysConverter)
	//bundleRepo := bundleutil.NewRepository(bundleConverter)
	//bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	//
	//labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	//assignmentConv := scenarioassignment.NewConverter()
	//scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	//scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, conf.Features.DefaultScenarioEnabled)
	//fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, &http.Client{Timeout: conf.ClientTimeout}, accessstrategy.NewDefaultExecutorProvider())
	//specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	//bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	//apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	//eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	//docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	//bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	//appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, scenariosSvc, bundleSvc, uidSvc)




	//h := handler.NewHandler(appSvc)
	h := handler.NewChunkedHandler()

	handlerWithTimeout, err := directorHandler.WithTimeoutWithErrorMessage(h, conf.ServerTimeout, httputil.GetTimeoutMessage())
	exitOnError(err, "Failed configuring timeout on handler")

	router := mux.NewRouter()

	router.Use(correlation.AttachCorrelationIDToContext())
	router.NewRoute().
		Methods(http.MethodPost). //TODO make me Put
		Path("/api/v1/notifications").
		Handler(handlerWithTimeout)
	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required")

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", conf.Port),
		Handler:           router,
		ReadHeaderTimeout: conf.ServerTimeout,
	}
	ctx, err = log.Configure(ctx, conf.Log)
	exitOnError(err, "while configuring logger")

	log.C(ctx).Infof("API listening on %s", conf.Address)
	exitOnError(server.ListenAndServe(), "on starting HTTP server")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func createAndRunConfigProvider(ctx context.Context, cfg adapter.Configuration) *configprovider.Provider {
	//TODO
	return nil
}