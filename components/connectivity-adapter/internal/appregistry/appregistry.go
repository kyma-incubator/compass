package appregistry

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
)

type Config struct{
	DirectorURL string `envconfig:"default=http://127.0.0.1:3000/graphql"`
}

func RegisterHandler(router *mux.Router, cfg Config) {
	serviceHandler := service.NewHandler(cfg.DirectorURL)

	router.HandleFunc("/services", serviceHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/services", serviceHandler.Create).Methods(http.MethodPost)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Update).Methods(http.MethodPut)
	router.HandleFunc("/services/{serviceId}", serviceHandler.Delete).Methods(http.MethodDelete)
}
