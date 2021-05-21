package appregistry

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/appdetails"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/validation"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli"
)

type Config struct {
	DirectorEndpoint string        `envconfig:"default=http://127.0.0.1:3000/graphql"`
	ClientTimeout    time.Duration `envconfig:"default=115s"`
}

func RegisterHandler(router *mux.Router, cfg Config) {
	gqlCliProvider := gqlcli.NewProvider(cfg.DirectorEndpoint, cfg.ClientTimeout)
	reqContextProvider := service.NewRequestContextProvider()

	converter := service.NewConverter()
	validator := validation.NewServiceDetailsValidator()

	labeler := service.NewAppLabeler()

	serviceHandler := service.NewHandler(converter, validator, reqContextProvider, labeler)
	appMiddleware := appdetails.NewApplicationMiddleware(gqlCliProvider)

	router.Use(appMiddleware.Middleware)
	router.HandleFunc("/services", serviceHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/services", serviceHandler.Create).Methods(http.MethodPost)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Update).Methods(http.MethodPut)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Delete).Methods(http.MethodDelete)
}
