package configuration

import (
	"net/http"

	"github.com/gorilla/mux"
)

func InitConfigurationChangeHandler(router *mux.Router, handler *ConfigChangeHandler) {
	router.HandleFunc("/search", handler.SearchByString).Methods(http.MethodGet)
	router.HandleFunc("", handler.Save).Methods(http.MethodPost)
	router.Path("/").HandlerFunc(handler.List).Methods(http.MethodGet)
	router.HandleFunc("/{id}", handler.Get).Methods(http.MethodGet)
	router.HandleFunc("/{id}", handler.Delete).Methods(http.MethodDelete)
	router.Path("/search").Queries("query", "*").HandlerFunc(handler.SearchByString).Methods("GET")
}
