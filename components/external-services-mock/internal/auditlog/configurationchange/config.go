package configurationchange

import (
	"net/http"

	"github.com/gorilla/mux"
)

func InitConfigurationChangeHandler(router *mux.Router, handler *ConfigChangeHandler) {
	router.HandleFunc("/search", handler.SearchByTimestamp).Methods(http.MethodGet)
	router.HandleFunc("", handler.Save).Methods(http.MethodPost)
	router.HandleFunc("", handler.List).Methods(http.MethodGet)
	router.HandleFunc("/{id}", handler.Get).Methods(http.MethodGet)
	router.HandleFunc("/{id}", handler.Delete).Methods(http.MethodDelete)
	router.Path("/search").Queries("timeFrom", "*", "timeTo", "*").HandlerFunc(handler.SearchByTimestamp).Methods("GET")
}
