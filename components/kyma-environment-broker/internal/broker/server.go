package broker

import (
	"net/http"

	"code.cloudfoundry.org/lager"
	"github.com/gorilla/mux"
	"github.com/pivotal-cf/brokerapi/v7/auth"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/handlers"
	"github.com/pivotal-cf/brokerapi/v7/middlewares"
)

type BrokerCredentials struct {
	Username string
	Password string
}

// copied from github.com/pivotal-cf/brokerapi/api.go
func New(serviceBroker domain.ServiceBroker, logger lager.Logger, brokerCredentials *BrokerCredentials) http.Handler {
	router := mux.NewRouter()

	AttachRoutes(router, serviceBroker, logger)

	if brokerCredentials != nil {
		authMiddleware := auth.NewWrapper(brokerCredentials.Username, brokerCredentials.Password).Wrap
		router.Use(authMiddleware)
	}
	apiVersionMiddleware := middlewares.APIVersionMiddleware{LoggerFactory: logger}

	router.Use(middlewares.AddCorrelationIDToContext)
	router.Use(middlewares.AddOriginatingIdentityToContext)
	router.Use(apiVersionMiddleware.ValidateAPIVersionHdr)
	router.Use(middlewares.AddInfoLocationToContext)

	return router
}

func AttachRoutes(router *mux.Router, serviceBroker domain.ServiceBroker, logger lager.Logger) {
	apiHandler := handlers.NewApiHandler(serviceBroker, logger)
	router.HandleFunc("/v2/catalog", apiHandler.Catalog).Methods("GET")

	router.HandleFunc("/v2/service_instances/{instance_id}", apiHandler.GetInstance).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", apiHandler.Provision).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}", apiHandler.Deprovision).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{instance_id}/last_operation", apiHandler.LastOperation).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}", apiHandler.Update).Methods("PATCH")

	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.GetBinding).Methods("GET")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.Bind).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}", apiHandler.Unbind).Methods("DELETE")

	router.HandleFunc("/v2/service_instances/{instance_id}/service_bindings/{binding_id}/last_operation", apiHandler.LastBindingOperation).Methods("GET")
}
