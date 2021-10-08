package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/cert"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/apispec"
	ord_aggregator "github.com/kyma-incubator/compass/components/external-services-mock/internal/ord-aggregator"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/systemfetcher"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configurationchange"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address                  string `envconfig:"default=127.0.0.1:8080"`
	CertSecuredServerAddress string `envconfig:"default=127.0.0.1:8081"`
	BaseURL                  string `envconfig:"default=http://compass-external-services-mock.compass-system.svc.cluster.local:8080"`
	JWKSPath                 string `envconfig:"default=/jwks.json"`
	OAuthConfig
	BasicCredentialsConfig
	DefaultTenant string `envconfig:"APP_DEFAULT_TENANT"`

	CACert        string `envconfig:"APP_CA_CERT"`
	CAKey         string `envconfig:"APP_CA_KEY"`
	CertSvcRootCA string `envconfig:"APP_CERT_SVC_ROOT_CA"`
}

type OAuthConfig struct {
	ClientID     string `envconfig:"APP_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_CLIENT_SECRET"`
}

type BasicCredentialsConfig struct {
	Username string `envconfig:"BASIC_USERNAME"`
	Password string `envconfig:"BASIC_PASSWORD"`
}

func main() {
	ctx := context.Background()

	cfg := config{}
	err := envconfig.InitWithOptions(&cfg, envconfig.Options{Prefix: "APP", AllOptional: true})
	exitOnError(err, "while loading configuration")

	handler, err := initHTTP(cfg)
	if err != nil {
		log.Fatal(err)
	}

	certSecuredHandler, err := initCertSecuredAPIs(cfg)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: handler,
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(cfg.CACert))
	if len(cfg.CertSvcRootCA) > 0 {
		rootCA, err := base64.StdEncoding.DecodeString(cfg.CertSvcRootCA)
		exitOnError(err, "while base64 decoding Cert Svc ROOT CA")
		caCertPool.AppendCertsFromPEM(rootCA)
	}

	certSecuredServer := &http.Server{
		Addr:    cfg.CertSecuredServerAddress,
		Handler: certSecuredHandler,
		TLSConfig: &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go startServer(ctx, server, wg)
	go startServer(ctx, certSecuredServer, wg)

	wg.Wait()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func initHTTP(cfg config) (http.Handler, error) {
	logger := logrus.New()
	router := mux.NewRouter()
	configChangeSvc := configurationchange.NewService()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, errors.Wrap(err, "while generating rsa key")
	}

	tokenHandler := oauth.NewHandlerWithSigningKey(cfg.ClientSecret, cfg.ClientID, key)
	router.HandleFunc("/oauth/token", tokenHandler.GenerateWithoutCredentials).Methods(http.MethodPost)
	router.HandleFunc("/secured/oauth/token", tokenHandler.Generate).Methods(http.MethodPost)

	openIDConfigHandler := oauth.NewOpenIDConfigHandler(cfg.BaseURL, cfg.JWKSPath)
	router.HandleFunc("/.well-known/openid-configuration", openIDConfigHandler.Handle)

	jwksHanlder := oauth.NewJWKSHandler(&key.PublicKey)
	router.HandleFunc(cfg.JWKSPath, jwksHanlder.Handle)

	configChangeHandler := configurationchange.NewConfigurationHandler(configChangeSvc, logger)

	certHandler := cert.NewHandler(cfg.CACert, cfg.CAKey)
	router.HandleFunc("/cert", certHandler.Generate).Methods(http.MethodPost)

	unsignedTokenHandler := oauth.NewHandler(cfg.ClientSecret, cfg.ClientID)

	router.HandleFunc("/v1/healtz", health.HandleFunc)

	configChangeRouter := router.PathPrefix("/audit-log/v2/configuration-changes").Subrouter()
	configChangeRouter.Use(oauthMiddleware)
	configurationchange.InitConfigurationChangeHandler(configChangeRouter, configChangeHandler)

	router.HandleFunc("/audit-log/v2/oauth/token", unsignedTokenHandler.Generate).Methods(http.MethodPost)

	router.HandleFunc("/external-api/unsecured/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/unsecured/spec/flapping", apispec.FlappingHandleFunc())

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("open"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(8080))

	router.HandleFunc("/test/fullPath", ord_aggregator.HandleFuncOrdConfig("open"))

	systemFetcherHandler := systemfetcher.NewSystemFetcherHandler(cfg.DefaultTenant)
	router.Methods(http.MethodPost).PathPrefix("/systemfetcher/configure").HandlerFunc(systemFetcherHandler.HandleConfigure)
	router.Methods(http.MethodDelete).PathPrefix("/systemfetcher/reset").HandlerFunc(systemFetcherHandler.HandleReset)
	router.HandleFunc("/systemfetcher/systems", systemFetcherHandler.HandleFunc)
	router.HandleFunc("/systemfetcher/oauth/token", unsignedTokenHandler.GenerateWithoutCredentials)

	oauthRouter := router.PathPrefix("/external-api/secured/oauth").Subrouter()
	oauthRouter.Use(oauthMiddleware)
	oauthRouter.HandleFunc("/spec", apispec.HandleFunc)

	basicAuthRouter := router.PathPrefix("/external-api/secured/basic").Subrouter()

	h := &handler{
		Username: cfg.Username,
		Password: cfg.Password,
	}
	basicAuthRouter.Use(h.basicAuthMiddleware)
	basicAuthRouter.HandleFunc("/spec", apispec.HandleFunc)

	router.HandleFunc(webhook.DeletePath, webhook.NewDeleteHTTPHandler()).Methods(http.MethodDelete)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationGetHTTPHandler()).Methods(http.MethodGet)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationPostHTTPHandler()).Methods(http.MethodPost)

	return router, nil
}

func initCertSecuredAPIs(cfg config) (http.Handler, error) {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("sap:cmp-mtls:v1"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(8081))

	return router, nil
}

func oauthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			httphelpers.WriteError(w, errors.New("No Authorization header"), http.StatusUnauthorized)
			return
		}
		if !strings.Contains(authHeader, "Bearer") {
			httphelpers.WriteError(w, errors.New("No Bearer token"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type handler struct {
	Username string `envconfig:"BASIC_USERNAME"`
	Password string `envconfig:"BASIC_PASSWORD"`
}

func (h *handler) basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok {
			httphelpers.WriteError(w, errors.New("No Basic credentials"), http.StatusUnauthorized)
			return
		}
		if username != h.Username || password != h.Password {
			httphelpers.WriteError(w, errors.New("Bad credentials"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func startServer(parentCtx context.Context, server *http.Server, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	go func() {
		defer wg.Done()
		<-ctx.Done()
		stopServer(server)
	}()

	log.Printf("Starting and listening on %s://%s", "http", server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s://%s: %v\n", "http", server.Addr, err)
	}
}

func stopServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()

		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Fatal("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	server.SetKeepAlivesEnabled(false)

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}
