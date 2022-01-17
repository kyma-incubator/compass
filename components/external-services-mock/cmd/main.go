package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	ord_global_registry "github.com/kyma-incubator/compass/components/external-services-mock/internal/ord-aggregator/globalregistry"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/subscription"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/selfreg"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/tenantfetcher"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/health"

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
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Port       int `envconfig:"default=8080"`
	CertPort   int `envconfig:"default=8081"`
	ORDServers ORDServers
	BaseURL    string `envconfig:"default=http://compass-external-services-mock.compass-system.svc.cluster.local"`
	JWKSPath   string `envconfig:"default=/jwks.json"`
	OAuthConfig
	BasicCredentialsConfig
	SelfRegConfig selfreg.Config
	DefaultTenant string `envconfig:"APP_DEFAULT_TENANT"`

	TenantConfig         subscription.Config
	TenantProviderConfig subscription.ProviderConfig

	CACert string `envconfig:"APP_CA_CERT"`
	CAKey  string `envconfig:"APP_CA_KEY"`
}

// ORDServers is a configuration for ORD e2e tests. Those tests are more complex and require a dedicated server per application involved.
// This is needed in order to ensure that every call in the context of an application happens in a single server isolated from others.
// Prior to this separation there were cases when tests succeeded (false positive) due to mistakenly configured baseURL resulting in different flow - different access strategy returned.
type ORDServers struct {
	CertPort           int `envconfig:"default=8082"`
	UnsecuredPort      int `envconfig:"default=8083"`
	BasicPort          int `envconfig:"default=8084"`
	OauthPort          int `envconfig:"default=8085"`
	GlobalRegistryPort int `envconfig:"default=8086"`

	OrdCertSecuredBaseURL string
}

type OAuthConfig struct {
	ClientID     string `envconfig:"APP_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_CLIENT_SECRET"`
	TokenPath    string `envconfig:"APP_TOKEN_PATH"`

	Scopes       string `envconfig:"APP_OAUTH_SCOPES"`
	TenantHeader string `envconfig:"APP_OAUTH_TENANT_HEADER"`
}

type BasicCredentialsConfig struct {
	Username string `envconfig:"BASIC_USERNAME"`
	Password string `envconfig:"BASIC_PASSWORD"`
}

func claimsFunc(uniqueAttrKey, uniqueAttrValue, tenantID, identity, iss string, scopes []string) oauth.ClaimsGetterFunc {
	return func() map[string]interface{} {
		return map[string]interface{}{
			uniqueAttrKey: uniqueAttrValue,
			"scope":       scopes,
			"tenant":      tenantID,
			"identity":    identity,
			"iss":         iss,
			"exp":         time.Now().Unix() + int64(time.Minute.Seconds()),
		}
	}
}

func main() {
	ctx := context.Background()

	cfg := config{}
	err := envconfig.InitWithOptions(&cfg, envconfig.Options{Prefix: "APP", AllOptional: true})
	exitOnError(err, "while loading configuration")

	extSvcMockURL := fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.Port)
	staticClaimsMapping := map[string]oauth.ClaimsGetterFunc{
		"tenantFetcherClaims": claimsFunc("test", "tenant-fetcher", "tenantID", "tenant-fetcher-test-identity", extSvcMockURL, []string{"prefix.Callback"}),
		"subscriptionClaims":  claimsFunc("subsc-key-test", "subscription-flow", cfg.TenantConfig.TestConsumerSubaccountID, "subscription-flow-identity", extSvcMockURL, []string{}),
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	exitOnError(err, "while generating rsa key")

	ordServers := initORDServers(cfg, key)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go startServer(ctx, initDefaultServer(cfg, key, staticClaimsMapping), wg)
	go startServer(ctx, initDefaultCertServer(cfg, key, staticClaimsMapping), wg)

	for _, server := range ordServers {
		wg.Add(1)
		go startServer(ctx, server, wg)
	}

	wg.Wait()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func initDefaultServer(cfg config, key *rsa.PrivateKey, staticMappingClaims map[string]oauth.ClaimsGetterFunc) *http.Server {
	logger := logrus.New()
	router := mux.NewRouter()

	router.HandleFunc("/v1/healtz", health.HandleFunc)

	// Oauth server handlers
	tokenHandler := oauth.NewHandlerWithSigningKey(cfg.ClientSecret, cfg.ClientID, cfg.TenantHeader, cfg.Username, cfg.Password, key, staticMappingClaims)
	router.HandleFunc("/secured/oauth/token", tokenHandler.Generate).Methods(http.MethodPost)
	router.HandleFunc("/oauth/token", tokenHandler.GenerateWithCredentialsFromReqBody).Methods(http.MethodPost)
	openIDConfigHandler := oauth.NewOpenIDConfigHandler(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.Port), cfg.JWKSPath)
	router.HandleFunc("/.well-known/openid-configuration", openIDConfigHandler.Handle)
	jwksHanlder := oauth.NewJWKSHandler(&key.PublicKey)
	router.HandleFunc(cfg.JWKSPath, jwksHanlder.Handle)

	// Subscription handlers
	subHandler := subscription.NewHandler(cfg.TenantConfig, cfg.TenantProviderConfig, fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.Port), cfg.TokenPath, cfg.ClientID, cfg.ClientSecret, staticMappingClaims)
	router.HandleFunc("/saas-manager/v1/application/tenants/{tenant_id}/subscriptions", subHandler.Subscription).Methods(http.MethodPost)
	router.HandleFunc("/saas-manager/v1/application/tenants/{tenant_id}/subscriptions", subHandler.Deprovisioning).Methods(http.MethodDelete)

	// OnSubscription callback handler
	router.HandleFunc("/tenants/v1/regional/{region}/callback/{tenantId}", subHandler.OnSubscription).Methods(http.MethodPut)

	// Get dependencies handler
	router.HandleFunc("/v1/dependencies/configure", subHandler.DependenciesConfigure).Methods(http.MethodPost)
	router.HandleFunc("/v1/dependencies", subHandler.Dependencies).Methods(http.MethodGet)

	// CA server handlers
	certHandler := cert.NewHandler(cfg.CACert, cfg.CAKey)
	router.HandleFunc("/cert", certHandler.Generate).Methods(http.MethodPost)

	// AL handlers
	configChangeSvc := configurationchange.NewService()
	configChangeHandler := configurationchange.NewConfigurationHandler(configChangeSvc, logger)
	configChangeRouter := router.PathPrefix("/audit-log/v2/configuration-changes").Subrouter()
	configChangeRouter.Use(oauthMiddleware(&key.PublicKey, noopClaimsValidator))
	configurationchange.InitConfigurationChangeHandler(configChangeRouter, configChangeHandler)

	// System fetcher handlers
	systemFetcherHandler := systemfetcher.NewSystemFetcherHandler(cfg.DefaultTenant)
	router.Methods(http.MethodPost).PathPrefix("/systemfetcher/configure").HandlerFunc(systemFetcherHandler.HandleConfigure)
	router.Methods(http.MethodDelete).PathPrefix("/systemfetcher/reset").HandlerFunc(systemFetcherHandler.HandleReset)
	systemsRouter := router.PathPrefix("/systemfetcher/systems").Subrouter()
	systemsRouter.Use(oauthMiddleware(&key.PublicKey, getClaimsValidator(cfg.DefaultTenant)))
	systemsRouter.HandleFunc("", systemFetcherHandler.HandleFunc)

	// Tenant fetcher handlers
	tenantFetcherHandler := tenantfetcher.NewHandler()

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-create/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("create"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-create/reset").HandlerFunc(tenantFetcherHandler.HandleReset("create"))
	router.HandleFunc("/tenant-fetcher/global-account-create", tenantFetcherHandler.HandleFunc("create"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-delete/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("delete"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-delete/reset").HandlerFunc(tenantFetcherHandler.HandleReset("delete"))
	router.HandleFunc("/tenant-fetcher/global-account-delete", tenantFetcherHandler.HandleFunc("delete"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-update/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("update"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-update/reset").HandlerFunc(tenantFetcherHandler.HandleReset("update"))
	router.HandleFunc("/tenant-fetcher/global-account-update", tenantFetcherHandler.HandleFunc("update"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-create/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("create_subaccount"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-create/reset").HandlerFunc(tenantFetcherHandler.HandleReset("create_subaccount"))
	router.HandleFunc("/tenant-fetcher/subaccount-create", tenantFetcherHandler.HandleFunc("create_subaccount"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-delete/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("delete_subaccount"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-delete/reset").HandlerFunc(tenantFetcherHandler.HandleReset("delete_subaccount"))
	router.HandleFunc("/tenant-fetcher/subaccount-delete", tenantFetcherHandler.HandleFunc("delete_subaccount"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-update/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("update_subaccount"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-update/reset").HandlerFunc(tenantFetcherHandler.HandleReset("update_subaccount"))
	router.HandleFunc("/tenant-fetcher/subaccount-update", tenantFetcherHandler.HandleFunc("update_subaccount"))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-move/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure("move_subaccount"))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-move/reset").HandlerFunc(tenantFetcherHandler.HandleReset("move_subaccount"))
	router.HandleFunc("/tenant-fetcher/subaccount-move", tenantFetcherHandler.HandleFunc("move_subaccount"))

	// Fetch request handlers
	router.HandleFunc("/external-api/spec", apispec.HandleFunc)

	oauthRouter := router.PathPrefix("/external-api/secured/oauth").Subrouter()
	oauthRouter.Use(oauthMiddleware(&key.PublicKey, noopClaimsValidator))
	oauthRouter.HandleFunc("/spec", apispec.HandleFunc)

	basicAuthRouter := router.PathPrefix("/external-api/secured/basic").Subrouter()
	basicAuthRouter.Use(basicAuthMiddleware(cfg.Username, cfg.Password))
	basicAuthRouter.HandleFunc("/spec", apispec.HandleFunc)

	// Operations controller handlers
	router.HandleFunc(webhook.DeletePath, webhook.NewDeleteHTTPHandler()).Methods(http.MethodDelete)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationGetHTTPHandler()).Methods(http.MethodGet)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationPostHTTPHandler()).Methods(http.MethodPost)

	// non-isolated and unsecured ORD handlers. NOTE: Do not host document endpoints on this default server in order to ensure tests separation.
	// Unsecured config pointing to cert secured document
	router.HandleFunc("/cert", ord_aggregator.HandleFuncOrdConfigWithDocPath(cfg.ORDServers.OrdCertSecuredBaseURL, "/open-resource-discovery/v1/documents/example2", "sap:cmp-mtls:v1"))

	selfRegisterHandler := selfreg.NewSelfRegisterHandler(cfg.SelfRegConfig)
	selfRegRouter := router.PathPrefix(cfg.SelfRegConfig.Path).Subrouter()
	selfRegRouter.Use(oauthMiddleware(&key.PublicKey, noopClaimsValidator))
	selfRegRouter.HandleFunc("", selfRegisterHandler.HandleSelfRegPrep).Methods(http.MethodPost)
	selfRegRouter.HandleFunc(fmt.Sprintf("/{%s}", selfreg.NamePath), selfRegisterHandler.HandleSelfRegCleanup).Methods(http.MethodDelete)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}
}

func initDefaultCertServer(cfg config, key *rsa.PrivateKey, staticMappingClaims map[string]oauth.ClaimsGetterFunc) *http.Server {
	router := mux.NewRouter()

	// Oauth server handlers
	tokenHandlerWithKey := oauth.NewHandlerWithSigningKey(cfg.ClientSecret, cfg.ClientID, cfg.TenantHeader, cfg.Username, cfg.Password, key, staticMappingClaims)
	// TODO The mtls_token_provider sends client id and scopes in url.values form. When the change for fetching xsuaa token
	// with certificate is merged GenerateWithCredentialsFromReqBody should be used for testing the flows that include fetching
	// xsuaa token with certificate. APP_SELF_REGISTER_OAUTH_TOKEN_PATH for local env should be adapted.
	router.HandleFunc("/oauth/token", tokenHandlerWithKey.GenerateWithCredentialsFromReqBody).Methods(http.MethodPost)

	tokenHandler := oauth.NewHandler(cfg.ClientSecret, cfg.ClientID)
	router.HandleFunc("/cert/token", tokenHandler.GenerateWithoutCredentials).Methods(http.MethodPost)

	router.HandleFunc(webhook.DeletePath, webhook.NewDeleteHTTPHandler()).Methods(http.MethodDelete)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationGetHTTPHandler()).Methods(http.MethodGet)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationPostHTTPHandler()).Methods(http.MethodPost)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.CertPort),
		Handler: router,
	}
}

func initORDServers(cfg config, key *rsa.PrivateKey) []*http.Server {
	servers := make([]*http.Server, 0, 0)
	servers = append(servers, initCertSecuredORDServer(cfg))
	servers = append(servers, initUnsecuredORDServer(cfg))
	servers = append(servers, initBasicSecuredORDServer(cfg))
	servers = append(servers, initOauthSecuredORDServer(cfg, key))
	servers = append(servers, initGlobalRegistryORDServer(cfg))
	return servers
}

func initCertSecuredORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "sap:cmp-mtls:v1"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(cfg.ORDServers.OrdCertSecuredBaseURL, "sap:cmp-mtls:v1"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocument(cfg.ORDServers.OrdCertSecuredBaseURL, "sap:cmp-mtls:v1"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.CertPort),
		Handler: router,
	}
}

func initUnsecuredORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "open"))
	router.HandleFunc("/test/fullPath", ord_aggregator.HandleFuncOrdConfigWithDocPath(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredPort), "/open-resource-discovery/v1/documents/example2", "open"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredPort), "open"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredPort), "open"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.UnsecuredPort),
		Handler: router,
	}
}

func initGlobalRegistryORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_global_registry.HandleFuncOrdConfig())

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_global_registry.HandleFuncOrdDocument())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.GlobalRegistryPort),
		Handler: router,
	}
}

func initBasicSecuredORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	configRouter := router.PathPrefix("/.well-known").Subrouter()
	configRouter.Use(basicAuthMiddleware(cfg.Username, cfg.Password))
	configRouter.HandleFunc("/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "open"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.BasicPort), "open"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.BasicPort),
		Handler: router,
	}
}

func initOauthSecuredORDServer(cfg config, key *rsa.PrivateKey) *http.Server {
	router := mux.NewRouter()

	configRouter := router.PathPrefix("/.well-known").Subrouter()
	configRouter.Use(oauthMiddleware(&key.PublicKey, noopClaimsValidator))
	configRouter.HandleFunc("/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "open"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.OauthPort), "open"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.OauthPort),
		Handler: router,
	}
}

func oauthMiddleware(key *rsa.PublicKey, validateClaims func(claims *oauth.Claims) bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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

			token := strings.TrimPrefix(authHeader, "Bearer ")
			parsed := &oauth.Claims{}

			if _, err := jwt.ParseWithClaims(token, parsed, func(_ *jwt.Token) (interface{}, error) {
				return key, nil
			}); err != nil {
				httphelpers.WriteError(w, errors.Wrap(err, "Invalid Bearer token"), http.StatusUnauthorized)
				return
			}
			if !validateClaims(parsed) {
				httphelpers.WriteError(w, errors.New("Could not validate claims"), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func basicAuthMiddleware(username, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()

			if !ok {
				httphelpers.WriteError(w, errors.New("No Basic credentials"), http.StatusUnauthorized)
				return
			}
			if username != u || password != p {
				httphelpers.WriteError(w, errors.New("Bad credentials"), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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

func noopClaimsValidator(_ *oauth.Claims) bool {
	return true
}
func getClaimsValidator(expectedTenant string) func(*oauth.Claims) bool {
	return func(claims *oauth.Claims) bool {
		return claims.Tenant == expectedTenant
	}
}
