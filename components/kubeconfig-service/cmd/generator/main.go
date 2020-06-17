package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/authn"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/reload"
	"k8s.io/apiserver/pkg/authentication/authenticator"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/endpoints"
	"github.com/kyma-incubator/compass/components/kubeconfig-service/pkg/env"
	//"github.com/kyma-project/kyma/components/iam-kubeconfig-service/cmd/generator/reload"
	log "github.com/sirupsen/logrus"
)

const (
	oidcIssuerURLFlag = "oidc-issuer-url"
	oidcClientIDFlag  = "oidc-client-id"
	oidcCAFileFlag    = "oidc-ca-file"
)

func main() {
	authnCfg := readAuthnConfig()
	env.InitConfig()

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
		err := http.ListenAndServe(fmt.Sprintf(":%d", env.Config.ServicePort), router)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Kubeconfig service started on port: %d", env.Config.ServicePort)

	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(9000), healthRouter)
		log.Errorf("Error serving HTTP: %v", err)
		term <- os.Interrupt
	}()

	log.Infof("Health endpoint started on port %d...", 9000)

	log.Infof("Using GraphQL Service: %s", env.Config.GraphqlURL)
	select {
	case <-term:
		log.Info("Received SIGTERM, exiting gracefully...")
		fileWatcherCtxCancel()
	}
}


func readAuthnConfig() *authn.OIDCConfig {
	oidcIssuerURLArg := flag.String(oidcIssuerURLFlag, "", "OIDC: The URL of the OpenID issuer. Used to verify the OIDC JSON Web Token (JWT)")
	oidcClientIDArg := flag.String(oidcClientIDFlag, "", "OIDC: The client ID for the OpenID Connect client")
	oidcUsernameClaimArg := flag.String("oidc-username-claim", "email", "OIDC: Identifier of the user in JWT claim")
	oidcGroupsClaimArg := flag.String("oidc-groups-claim", "groups", "OIDC: Identifier of groups in JWT claim")
	oidcUsernamePrefixArg := flag.String("oidc-username-prefix", "", "OIDC: If provided, all users will be prefixed with this value to prevent conflicts with other authentication strategies")
	oidcGroupsPrefixArg := flag.String("oidc-groups-prefix", "", "OIDC: If provided, all groups will be prefixed with this value to prevent conflicts with other authentication strategies")
	oidcCAFileArg := flag.String(oidcCAFileFlag, "", "File with Certificate Authority of the Kubernetes cluster, also used for OIDC authentication")

	var oidcSupportedSigningAlgsArg multiValFlag = []string{}
	flag.Var(&oidcSupportedSigningAlgsArg, "oidc-supported-signing-algs", "OIDC supported signing algorithms")

	flag.Parse()

	errors := false

	if *oidcIssuerURLArg == "" {
		log.Errorf("OIDC Issuer URL is required (-%s)", oidcIssuerURLFlag)
		errors = true
	}

	if *oidcClientIDArg == "" {
		log.Errorf("OIDC Client ID is required (-%s)", oidcClientIDFlag)
		errors = true
	}

	if errors {
		flag.Usage()
		os.Exit(1)
	}

	if len(oidcSupportedSigningAlgsArg) == 0 {
		oidcSupportedSigningAlgsArg = []string{"RS256"}
	}

	return &authn.OIDCConfig {
			IssuerURL:            *oidcIssuerURLArg,
			ClientID:             *oidcClientIDArg,
			CAFilePath:           *oidcCAFileArg,
			UsernameClaim:        *oidcUsernameClaimArg,
			UsernamePrefix:       *oidcUsernamePrefixArg,
			GroupsClaim:          *oidcGroupsClaimArg,
			GroupsPrefix:         *oidcGroupsPrefixArg,
			SupportedSigningAlgs: oidcSupportedSigningAlgsArg,
	}
}


//Support for multi-valued flag: -flagName=val1 -flagName=val2 etc.
type multiValFlag []string

func (vals *multiValFlag) String() string {
	res := "["

	if len(*vals) > 0 {
		res = res + (*vals)[0]
	}

	for _, v := range *vals {
		res = res + ", " + v
	}
	res = res + "]"
	return res
}

func (vals *multiValFlag) Set(value string) error {
	*vals = append(*vals, value)
	return nil
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
