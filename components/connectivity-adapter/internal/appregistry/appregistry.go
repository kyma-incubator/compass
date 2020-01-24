package appregistry

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/validation"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
)

type Config struct {
	DirectorEndpoint string `envconfig:"default=http://127.0.0.1:3000/graphql"`
}

func RegisterHandler(router *mux.Router, cfg Config) {
	logger := logrus.New().WithField("component", "app-registry").Logger
	logger.SetReportCaller(true)

	gqlCliProvider := gqlcli.NewProvider(cfg.DirectorEndpoint)
	converter := service.NewConverter()
	validator := validation.NewServiceDetailsValidator()
	gqlRequestBuilder := service.NewGqlRequestBuilder()
	serviceHandler := service.NewHandler(gqlCliProvider, converter, validator, gqlRequestBuilder, logger)

	router.HandleFunc("/services", serviceHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/services", serviceHandler.Create).Methods(http.MethodPost)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Update).Methods(http.MethodPut)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Delete).Methods(http.MethodDelete)
}
