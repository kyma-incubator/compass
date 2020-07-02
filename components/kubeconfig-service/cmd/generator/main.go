package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/authn"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/reload"
	"k8s.io/apiserver/pkg/authentication/authenticator"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/endpoints"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"

	log "github.com/sirupsen/logrus"
)

func main() {
	env.InitConfig()
	authnCfg := readAuthnConfig()
	log.Info("Starting kubeconfig-service sever")

	fileWatcherCtx, fileWatcherCtxCancel := context.WithCancel(context.Background())

	oidcAuthenticator, err := setupOIDCAuthReloader(fileWatcherCtx, authnCfg)

	if err != nil {
		log.Fatalf("Cannot create OIDC Authenticator, %v", err)
	}

	ec := endpoints.NewEndpointClient(env.Config.GraphqlURL)
	router := mux.NewRouter()
	router.Use(authn.AuthMiddleware(oidcAuthenticator))
	router.Methods("GET").Path("/kubeconfig/{tenantID}/{runtimeID}").HandlerFunc(ec.GetKubeConfig)

	healthRouter := mux.NewRouter()
	healthRouter.Methods("GET").Path("/health/ready").HandlerFunc(ec.GetHealthStatus)

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", env.Config.Port.Service), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Kubeconfig service started on port: %d", env.Config.Port.Service)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", env.Config.Port.Health), healthRouter)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Health endpoint started on port: %d", env.Config.Port.Health)

	log.Infof("Using GraphQL Service: %s", env.Config.GraphqlURL)
	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
		fileWatcherCtxCancel()
	}
}

func readAuthnConfig() *authn.OIDCConfig {
	return &authn.OIDCConfig{
		IssuerURL:            env.Config.OIDC.IssuerURL,
		ClientID:             env.Config.OIDC.ClientID,
		CAFilePath:           env.Config.OIDC.CA,
		UsernameClaim:        env.Config.OIDC.Claim.Username,
		UsernamePrefix:       env.Config.OIDC.Prefix.Username,
		GroupsClaim:          env.Config.OIDC.Claim.Groups,
		GroupsPrefix:         env.Config.OIDC.Prefix.Groups,
		SupportedSigningAlgs: env.Config.OIDC.SupportedSigningAlgs,
	}
}

func setupOIDCAuthReloader(fileWatcherCtx context.Context, cfg *authn.OIDCConfig) (authenticator.Request, error) {
	const eventBatchDelaySeconds = 10
	filesToWatch := []string{cfg.CAFilePath}

	cancelableAuthReqestConstructor := func() (authn.CancelableAuthRequest, error) {
		log.Infof("creating a new cancelable instance of authenticator.Request...")
		return authn.NewOIDCAuthenticator(cfg)
	}

	//Create reloader
	result, err := reload.NewCancelableAuthReqestReloader(cancelableAuthReqestConstructor)
	if err != nil {
		return nil, err
	}

	//Setup file watcher
	oidcCAFileWatcher := reload.NewWatcher("oidc-ca-dex-tls-cert", filesToWatch, eventBatchDelaySeconds, result.Reload)
	go oidcCAFileWatcher.Run(fileWatcherCtx)

	return result, nil
}
