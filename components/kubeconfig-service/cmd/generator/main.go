package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/endpoints"
	log "github.com/sirupsen/logrus"
)

type config struct {
	port             int
	oidcIssuerURL    string
	oidcClientID     string
	oidcClientSecret string
	graphqlURL       string
}

func getConfig() *config {
	port := flag.Int("port", 8000, "Application port")
	oidcIssuerURL := flag.String("oidc-issuer-url", "", "URL of the OIDC provider")
	oidcClientID := flag.String("oidc-client-id", "", "A client id that token is issued for")
	oidcClientSecret := flag.String("oidc-client-secret", "", "A client's secret")
	graphqlURL := flag.String("graphql-url", "", "URL to the GraphQL service")

	flag.Parse()

	return &config{
		port:             *port,
		oidcIssuerURL:    *oidcIssuerURL,
		oidcClientID:     *oidcClientID,
		oidcClientSecret: *oidcClientSecret,
		graphqlURL:       *graphqlURL,
	}
}

func main() {
	cfg := getConfig()

	log.Info("Starting kubeconfig-service sever")

	ec := endpoints.NewEndpointClient(cfg.graphqlURL, cfg.oidcIssuerURL, cfg.oidcClientID, cfg.oidcClientSecret)
	router := mux.NewRouter()

	router.Methods("GET").Path("/kubeconfig/{tenantID}/{runtimeID}").HandlerFunc(ec.GetKubeConfig)
	router.Methods("GET").Path("/health/ready").HandlerFunc(ec.GetHealthStatus)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Kubeconfig service started on port: %d", cfg.port)
	log.Infof("Using GraphQL Service: %s", cfg.graphqlURL)
	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
	}
}
