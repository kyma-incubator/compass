package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/endpoints"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"
	log "github.com/sirupsen/logrus"
)

func main() {
	env.InitConfig()

	log.Info("Starting kubeconfig-service sever")
	ec := endpoints.NewEndpointClient(env.Config.GraphqlURL)
	router := mux.NewRouter()

	router.Methods("GET").Path("/kubeconfig/{tenantID}/{runtimeID}").HandlerFunc(ec.GetKubeConfig)
	router.Methods("GET").Path("/health/ready").HandlerFunc(ec.GetHealthStatus)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", env.Config.ServicePort), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Kubeconfig service started on port: %d", env.Config.ServicePort)
	log.Infof("Using GraphQL Service: %s", env.Config.GraphqlURL)
	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
	}
}
