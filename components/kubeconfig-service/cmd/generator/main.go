package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/endpoints"
	log "github.com/sirupsen/logrus"
)

type config struct {
	port       int
	address    string
	graphqlURL string
}

func main() {
	port := flag.Int("port", 8000, "Application port")
	address := flag.String("address", "", "Kubeconfig address")
	graphqlURL := flag.String("graphql-url", "", "URL to the GraphQL service")

	flag.Parse()

	cfg := config{
		port:       *port,
		address:    *address,
		graphqlURL: *graphqlURL,
	}

	log.Info("Starting kubeconfig-service sever")

	router := mux.NewRouter()

	router.Methods("GET").Path("/kubeconfig").HandlerFunc(endpoints.GetKubeConfig)
	router.Methods("GET").Path("/health/ready").HandlerFunc(endpoints.GetHealthStatus)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(cfg.port), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Kubeconfig service started on port: %d...", cfg.port)
	log.Infof("Using GraphQL Service: %s", cfg.graphqlURL)
	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
	}
}
