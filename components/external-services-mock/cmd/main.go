package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationcreator"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/formationnotification"

	"github.com/form3tech-oss/jwt-go"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/destinationfetcher"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/ias"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/provider"

	ord_global_registry "github.com/kyma-incubator/compass/components/external-services-mock/internal/ord-aggregator/globalregistry"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/subscription"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/selfreg"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/tenantfetcher"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/health"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/cert"

	modelJwt "github.com/kyma-incubator/compass/components/external-services-mock/internal/jwt"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/oauth"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/apispec"
	ord_aggregator "github.com/kyma-incubator/compass/components/external-services-mock/internal/ord-aggregator"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/systemfetcher"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configurationchange"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const healthzEndpoint = "/v1/healthz"

type config struct {
	Port        int `envconfig:"default=8080"`
	CertPort    int `envconfig:"default=8081"`
	ExternalURL string
	BaseURL     string `envconfig:"default=http://compass-external-services-mock.compass-system.svc.cluster.local"`
	JWKSPath    string `envconfig:"default=/jwks.json"`
	OAuthConfig
	BasicCredentialsConfig
	NotificationConfig       formationnotification.Configuration
	DestinationCreatorConfig *destinationcreator.Config
	DestinationServiceConfig DestinationServiceConfig
	ORDServers               ORDServers
	SelfRegConfig            selfreg.Config
	DefaultTenant            string `envconfig:"APP_DEFAULT_TENANT"`
	DefaultCustomerTenant    string `envconfig:"APP_DEFAULT_CUSTOMER_TENANT"`
	TrustedTenant            string `envconfig:"APP_TRUSTED_TENANT"`
	OnDemandTenant           string `envconfig:"APP_ON_DEMAND_TENANT"`

	KeyLoaderConfig credloader.KeysConfig

	TenantConfig         subscription.Config
	TenantProviderConfig subscription.ProviderConfig

	CACert string `envconfig:"APP_CA_CERT"`
	CAKey  string `envconfig:"APP_CA_KEY"`

	DirectDependencyXsappname string `envconfig:"APP_DIRECT_DEPENDENCY_XSAPPNAME"`
}

// DestinationServiceConfig configuration for destination service endpoints.
type DestinationServiceConfig struct {
	TenantDestinationsSubaccountLevelEndpoint           string `envconfig:"APP_DESTINATION_TENANT_SUBACCOUNT_LEVEL_ENDPOINT,default=/destination-configuration/v1/subaccountDestinations"`
	TenantDestinationCertificateSubaccountLevelEndpoint string `envconfig:"APP_DESTINATION_CERTIFICATE_TENANT_SUBACCOUNT_LEVEL_ENDPOINT,default=/destination-configuration/v1/subaccountCertificates"`
	TenantDestinationCertificateInstanceLevelEndpoint   string `envconfig:"APP_DESTINATION_CERTIFICATE_TENANT_INSTANCE_LEVEL_ENDPOINT,default=/destination-configuration/v1/instanceCertificates"`
	TenantDestinationFindAPIEndpoint                    string `envconfig:"APP_DESTINATION_SERVICE_FIND_API_ENDPOINT,default=/destination-configuration/local/v1/destinations"`
	SensitiveDataEndpoint                               string `envconfig:"APP_DESTINATION_SENSITIVE_DATA_ENDPOINT,default=/destination-configuration/v1/destinations"`
	SubaccountIDClaimKey                                string `envconfig:"APP_DESTINATION_SUBACCOUNT_CLAIM_KEY"`
	ServiceInstanceClaimKey                             string `envconfig:"APP_DESTINATION_SERVICE_INSTANCE_CLAIM_KEY"`
	TestDestinationInstanceID                           string `envconfig:"APP_TEST_DESTINATION_INSTANCE_ID"`
}

// ORDServers is a configuration for ORD e2e tests. Those tests are more complex and require a dedicated server per application involved.
// This is needed in order to ensure that every call in the context of an application happens in a single server isolated from others.
// Prior to this separation there were cases when tests succeeded (false positive) due to mistakenly configured baseURL resulting in different flow - different access strategy returned.
type ORDServers struct {
	CertPort                           int `envconfig:"default=8082"`
	UnsecuredPort                      int `envconfig:"default=8083"`
	BasicPort                          int `envconfig:"default=8084"`
	OauthPort                          int `envconfig:"default=8085"`
	GlobalRegistryCertPort             int `envconfig:"default=8086"`
	GlobalRegistryUnsecuredPort        int `envconfig:"default=8087"`
	UnsecuredWithAdditionalContentPort int `envconfig:"default=8088"`
	UnsecuredMultiTenantPort           int `envconfig:"default=8089"`
	ProxyPort                          int `envconfig:"default=8090"`
	CertSecuredBaseURL                 string
	CertSecuredGlobalBaseURL           string
}

type OAuthConfig struct {
	ClientID     string `envconfig:"APP_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_CLIENT_SECRET"`
	Scopes       string `envconfig:"APP_OAUTH_SCOPES"`
	TenantHeader string `envconfig:"APP_OAUTH_TENANT_HEADER"`
}

type BasicCredentialsConfig struct {
	Username string `envconfig:"BASIC_USERNAME"`
	Password string `envconfig:"BASIC_PASSWORD"`
}

func claimsFunc(uniqueAttrKey, uniqueAttrValue, clientID, tenantID, identity, userNameClaim, iss string, scopes []string, extAttributes map[string]interface{}) oauth.ClaimsGetterFunc {
	return func() map[string]interface{} {
		return map[string]interface{}{
			uniqueAttrKey: uniqueAttrValue,
			"ext_attr":    extAttributes,
			"scope":       scopes,
			"client_id":   clientID,
			"tenant":      tenantID,
			"identity":    identity,
			"user_name":   userNameClaim,
			"iss":         iss,
			"exp":         time.Now().Unix() + int64(time.Minute.Seconds()*10),
		}
	}
}

func main() {
	ctx := context.Background()

	cfg := config{}
	err := envconfig.InitWithOptions(&cfg, envconfig.Options{Prefix: "APP", AllOptional: true})
	exitOnError(err, "while loading configuration")

	keyCache, err := credloader.StartKeyLoader(ctx, cfg.KeyLoaderConfig)
	exitOnError(err, "failed to initialize key loader")

	err = credloader.WaitForKeyCache(keyCache)
	exitOnError(err, "failed to wait key loader")

	extSvcMockURL := fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.Port)
	staticClaimsMapping := map[string]oauth.ClaimsGetterFunc{
		claims.TenantFetcherClaimKey:                   claimsFunc("test", "tenant-fetcher", "client_id", cfg.TenantConfig.TestConsumerSubaccountID, "tenant-fetcher-test-identity", "", extSvcMockURL, []string{"prefix.Callback"}, map[string]interface{}{}),
		claims.SubscriptionClaimKey:                    claimsFunc("subsc-key-test", "subscription-flow", cfg.TenantConfig.SubscriptionProviderID, cfg.TenantConfig.TestConsumerSubaccountID, "subscription-flow-identity", "user-name@test.com", extSvcMockURL, []string{}, map[string]interface{}{cfg.TenantConfig.ConsumerClaimsTenantIDKey: cfg.TenantConfig.TestConsumerSubaccountID, cfg.TenantConfig.ConsumerClaimsSubdomainKey: "consumerSubdomain"}),
		claims.NotificationServiceAdapterClaimKey:      claimsFunc("ns-adapter-test", "ns-adapter-flow", "test_prefix", cfg.DefaultTenant, "nsadapter-flow-identity", "", extSvcMockURL, []string{}, map[string]interface{}{"subaccountid": "08b6da37-e911-48fb-a0cb-fa635a6c4321"}),
		claims.TenantFetcherTenantHierarchyClaimKey:    claimsFunc("test", "tenant-fetcher", "client_id", cfg.TenantConfig.TestConsumerSubaccountIDTenantHierarchy, "tenant-fetcher-test-identity", "", extSvcMockURL, []string{"prefix.Callback"}, map[string]interface{}{}),
		claims.DestinationProviderClaimKey:             claimsFunc("dest-provider-key-test", "destination-flow", "client_id", cfg.TenantConfig.TestProviderSubaccountID, "destination-provider-flow-identity", "destination-user-name@test.com", extSvcMockURL, []string{}, map[string]interface{}{cfg.DestinationServiceConfig.SubaccountIDClaimKey: cfg.TenantConfig.TestProviderSubaccountID}),
		claims.DestinationProviderWithInstanceClaimKey: claimsFunc("dest-provider-with-instance-key-test", "destination-flow", "client_id", cfg.TenantConfig.TestProviderSubaccountID, "destination-provider-with-instance-flow-identity", "destination-user-name@test.com", extSvcMockURL, []string{}, map[string]interface{}{cfg.DestinationServiceConfig.SubaccountIDClaimKey: cfg.TenantConfig.TestProviderSubaccountID, cfg.DestinationServiceConfig.ServiceInstanceClaimKey: cfg.DestinationServiceConfig.TestDestinationInstanceID}),
		claims.DestinationConsumerClaimKey:             claimsFunc("dest-consumer-key-test", "destination-flow", "client_id", cfg.TenantConfig.TestConsumerSubaccountID, "destination-consumer-flow-identity", "destination-user-name@test.com", extSvcMockURL, []string{}, map[string]interface{}{cfg.DestinationServiceConfig.SubaccountIDClaimKey: cfg.TenantConfig.TestConsumerSubaccountID}),
		claims.DestinationConsumerWithInstanceClaimKey: claimsFunc("dest-consumer-with-instance-key-test", "destination-flow", "client_id", cfg.TenantConfig.TestConsumerSubaccountID, "destination-consumer-with-instance-flow-identity", "destination-user-name@test.com", extSvcMockURL, []string{}, map[string]interface{}{cfg.DestinationServiceConfig.SubaccountIDClaimKey: cfg.TenantConfig.TestConsumerSubaccountID, cfg.DestinationServiceConfig.ServiceInstanceClaimKey: cfg.DestinationServiceConfig.TestDestinationInstanceID}),
		claims.AccountAuthenticatorClaimKey:            claimsFunc("unique-attr-authenticator-key", "unique-attr-authenticator-value", "client_id", "", "", "user-name-account-authenticator@test.com", extSvcMockURL, []string{"prefix.application:read", "prefix2.application:write"}, map[string]interface{}{"globalaccountid": "5984a414-1eed-4972-af2c-b2b6a415c7d7"}), // the '5984a414-1eed-4972-af2c-b2b6a415c7d7' is the external ID of 'ApplicationsForRuntimeTenantName' test tenant
		claims.SubaccountAuthenticatorClaimKey:         claimsFunc("unique-attr-authenticator-key", "unique-attr-authenticator-value", "client_id", "", "", "user-name-subaccount-authenticator@test.com", extSvcMockURL, []string{"prefix.application:read", "prefix2.application:write"}, map[string]interface{}{"subaccountid": "e1e2f861-2b2e-42a9-ba9f-404d292e5471"}), // the 'e1e2f861-2b2e-42a9-ba9f-404d292e5471' is the external ID of 'TestTenantSubstitutionSubaccount2' test tenant
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	exitOnError(err, "while generating rsa key")

	ordServers := initORDServers(cfg, key)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	httpClient := &http.Client{
		Timeout: 2 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	destinationCreatorHandler := destinationcreator.NewHandler(cfg.DestinationCreatorConfig)

	go startServer(ctx, initDefaultServer(cfg, keyCache, key, staticClaimsMapping, httpClient, destinationCreatorHandler), wg)
	go startServer(ctx, initDefaultCertServer(cfg, key, staticClaimsMapping, destinationCreatorHandler), wg)

	for _, server := range ordServers {
		wg.Add(1)
		go startServer(ctx, server, wg)
	}

	wg.Wait()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func initDefaultServer(cfg config, keyCache credloader.KeysCache, key *rsa.PrivateKey, staticMappingClaims map[string]oauth.ClaimsGetterFunc, httpClient *http.Client, destinationCreatorHandler *destinationcreator.Handler) *http.Server {
	logger := logrus.New()
	router := mux.NewRouter()
	router.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthzEndpoint))

	router.HandleFunc(healthzEndpoint, health.HandleFunc)

	// Oauth server handlers
	tokenHandler := oauth.NewHandlerWithSigningKey(cfg.ClientSecret, cfg.ClientID, cfg.Username, cfg.Password, cfg.TenantHeader, cfg.ExternalURL, key, staticMappingClaims)
	router.HandleFunc("/secured/oauth/token", tokenHandler.Generate).Methods(http.MethodPost)
	openIDConfigHandler := oauth.NewOpenIDConfigHandler(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.Port), cfg.JWKSPath)
	router.HandleFunc("/.well-known/openid-configuration", openIDConfigHandler.Handle)
	jwksHanlder := oauth.NewJWKSHandler(&key.PublicKey)
	router.HandleFunc(cfg.JWKSPath, jwksHanlder.Handle)

	// Subscription handlers that mock subscription manager API's. On real environment we use the same path but with different(real) host
	jobID := "818cbe72-8dea-4e01-850d-bc1b54b00e78" // randomly chosen UUID
	subHandler := subscription.NewHandler(httpClient, cfg.TenantConfig, cfg.TenantProviderConfig, jobID)
	router.HandleFunc("/saas-manager/v1/applications/{app_name}/subscription", subHandler.Subscribe).Methods(http.MethodPost)
	router.HandleFunc("/saas-manager/v1/applications/{app_name}/subscription", subHandler.Unsubscribe).Methods(http.MethodDelete)
	router.HandleFunc(fmt.Sprintf("/api/v1/jobs/%s", jobID), subHandler.JobStatus).Methods(http.MethodGet)

	// Both handlers below are part of the provider setup. On real environment when someone is subscribed to provider tenant we want to mock OnSubscription and GetDependency callbacks
	// and return expected results. CMP will be returned as dependency and will execute its subscription logic.
	// On local setup, subscription request will be directly to tenant fetcher component with preconfigured data, without a need of these mocks.

	providerHandler := provider.NewHandler(cfg.DirectDependencyXsappname)
	// OnSubscription callback handler. It handles subscription manager API callback request executed on real environment when someone is subscribed to a given tenant
	router.HandleFunc("/tenants/v1/regional/{region}/callback/{tenantId}", providerHandler.OnSubscription).Methods(http.MethodPut, http.MethodDelete)

	// Get dependencies handler. It handles subscription manager API dependency callback request executed on real environment when someone is subscribed to a given tenant
	router.HandleFunc("/v1/dependencies/configure", providerHandler.DependenciesConfigure).Methods(http.MethodPost)
	router.HandleFunc("/v1/dependencies", providerHandler.Dependencies).Methods(http.MethodGet)
	router.HandleFunc("/v1/dependencies/indirect", providerHandler.DependenciesIndirect).Methods(http.MethodGet)

	// CA server handlers
	certHandler := cert.NewHandler(cfg.CACert, cfg.CAKey)
	router.HandleFunc("/cert", certHandler.Generate).Methods(http.MethodPost)

	// AL handlers
	configChangeSvc := configurationchange.NewService()
	configChangeHandler := configurationchange.NewConfigurationHandler(configChangeSvc, logger)
	configChangeRouter := router.PathPrefix("/audit-log/v2/configuration-changes").Subrouter()
	configChangeRouter.Use(oauthMiddleware(&key.PublicKey, noopClaimsValidator))
	configurationchange.InitConfigurationChangeHandler(configChangeRouter, configChangeHandler)

	// Destination Service handler
	destinationHandler := destinationfetcher.NewHandler()
	tenantDestinationEndpoint := cfg.DestinationServiceConfig.TenantDestinationsSubaccountLevelEndpoint
	sensitiveDataEndpoint := cfg.DestinationServiceConfig.SensitiveDataEndpoint + "/{name}"
	router.HandleFunc(tenantDestinationEndpoint,
		destinationHandler.GetSubaccountDestinationsPage).Methods(http.MethodGet)
	router.HandleFunc(tenantDestinationEndpoint, destinationHandler.PostDestination).Methods(http.MethodPost)
	router.HandleFunc(tenantDestinationEndpoint+"/{name}", destinationHandler.DeleteDestination).Methods(http.MethodDelete)
	router.HandleFunc(sensitiveDataEndpoint, destinationHandler.GetSensitiveData).Methods(http.MethodGet)

	// destination service handlers but the destination creator handler is used due to shared mappings
	router.HandleFunc(cfg.DestinationServiceConfig.TenantDestinationFindAPIEndpoint+"/{name}", destinationCreatorHandler.FindDestinationByNameFromDestinationSvc).Methods(http.MethodGet)

	router.HandleFunc(cfg.DestinationServiceConfig.TenantDestinationCertificateSubaccountLevelEndpoint+"/{name}", destinationCreatorHandler.GetDestinationCertificateByNameFromDestinationSvc).Methods(http.MethodGet)
	router.HandleFunc(cfg.DestinationServiceConfig.TenantDestinationCertificateInstanceLevelEndpoint+"/{name}", destinationCreatorHandler.GetDestinationCertificateByNameFromDestinationSvc).Methods(http.MethodGet)

	var iasConfig ias.Config
	err := envconfig.Init(&iasConfig)
	exitOnError(err, "while loading IAS adapter config")
	iasHandler := ias.NewHandler(iasConfig)
	router.HandleFunc("/ias/Applications/v1", iasHandler.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/ias/Applications/v1/{appID}", iasHandler.Patch).Methods(http.MethodPatch)

	// System fetcher handlers
	systemFetcherHandler := systemfetcher.NewSystemFetcherHandler(cfg.DefaultTenant)
	router.Methods(http.MethodPost).PathPrefix("/systemfetcher/configure").HandlerFunc(systemFetcherHandler.HandleConfigure)
	router.Methods(http.MethodDelete).PathPrefix("/systemfetcher/reset").HandlerFunc(systemFetcherHandler.HandleReset)
	systemsRouter := router.PathPrefix("/systemfetcher/systems").Subrouter()
	systemsRouter.Use(oauthMiddlewareMultiple([]MiddlewareArgs{
		{
			key:            &key.PublicKey,
			validateClaims: getClaimsValidator([]string{cfg.DefaultTenant, cfg.TrustedTenant}),
			ClaimGetter:    func() jwt.Claims { return &oauth.Claims{} },
		},
		{
			key:            keyCache.Get()[cfg.KeyLoaderConfig.KeysSecretName].PublicKey,
			validateClaims: getClaimsValidator([]string{cfg.DefaultCustomerTenant, cfg.TrustedTenant}),
			ClaimGetter:    func() jwt.Claims { return &modelJwt.Claims{} },
		},
	}))
	systemsRouter.HandleFunc("", systemFetcherHandler.HandleFunc)

	// Tenant fetcher handlers
	allowedSubaccounts := []string{cfg.OnDemandTenant, cfg.TenantConfig.TestTenantOnDemandID}
	tenantFetcherHandler := tenantfetcher.NewHandler(allowedSubaccounts, cfg.DefaultTenant, cfg.DefaultCustomerTenant)

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-create/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.AccountCreationEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-create/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.AccountCreationEventType))
	router.HandleFunc("/tenant-fetcher/global-account-create", tenantFetcherHandler.HandleFunc(tenantfetcher.AccountCreationEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-delete/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.AccountDeletionEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-delete/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.AccountDeletionEventType))
	router.HandleFunc("/tenant-fetcher/global-account-delete", tenantFetcherHandler.HandleFunc(tenantfetcher.AccountDeletionEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/global-account-update/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.AccountUpdateEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/global-account-update/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.AccountUpdateEventType))
	router.HandleFunc("/tenant-fetcher/global-account-update", tenantFetcherHandler.HandleFunc(tenantfetcher.AccountUpdateEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-create/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.SubaccountCreationEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-create/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.SubaccountCreationEventType))
	router.HandleFunc("/tenant-fetcher/subaccount-create", tenantFetcherHandler.HandleFunc(tenantfetcher.SubaccountCreationEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-delete/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.SubaccountDeletionEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-delete/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.SubaccountDeletionEventType))
	router.HandleFunc("/tenant-fetcher/subaccount-delete", tenantFetcherHandler.HandleFunc(tenantfetcher.SubaccountDeletionEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-update/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.SubaccountUpdateEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-update/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.SubaccountUpdateEventType))
	router.HandleFunc("/tenant-fetcher/subaccount-update", tenantFetcherHandler.HandleFunc(tenantfetcher.SubaccountUpdateEventType))

	router.Methods(http.MethodPost).PathPrefix("/tenant-fetcher/subaccount-move/configure").HandlerFunc(tenantFetcherHandler.HandleConfigure(tenantfetcher.SubaccountMoveEventType))
	router.Methods(http.MethodDelete).PathPrefix("/tenant-fetcher/subaccount-move/reset").HandlerFunc(tenantFetcherHandler.HandleReset(tenantfetcher.SubaccountMoveEventType))
	router.HandleFunc("/tenant-fetcher/subaccount-move", tenantFetcherHandler.HandleFunc(tenantfetcher.SubaccountMoveEventType))

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
	router.HandleFunc("/cert", ord_aggregator.HandleFuncOrdConfigWithDocPath(cfg.ORDServers.CertSecuredBaseURL, "/open-resource-discovery/v1/documents/example2", "sap:cmp-mtls:v1"))

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

func initDefaultCertServer(cfg config, key *rsa.PrivateKey, staticMappingClaims map[string]oauth.ClaimsGetterFunc, destinationCreatorHandler *destinationcreator.Handler) *http.Server {
	router := mux.NewRouter()
	router.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthzEndpoint))

	// Healthz handler
	router.HandleFunc(healthzEndpoint, health.HandleFunc)

	// Oauth server handlers
	tokenHandlerWithKey := oauth.NewHandlerWithSigningKey(cfg.ClientSecret, cfg.ClientID, cfg.Username, cfg.Password, cfg.TenantHeader, cfg.ExternalURL, key, staticMappingClaims)
	// TODO The mtls_token_provider sends client id and scopes in url.values form. When the change for fetching xsuaa token
	// with certificate is merged GenerateWithCredentialsFromReqBody should be used for testing the flows that include fetching
	// xsuaa token with certificate. APP_SELF_REGISTER_OAUTH_TOKEN_PATH for local env should be adapted.
	router.HandleFunc("/cert/token", tokenHandlerWithKey.Generate).Methods(http.MethodPost)

	router.HandleFunc(webhook.DeletePath, webhook.NewDeleteHTTPHandler()).Methods(http.MethodDelete)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationGetHTTPHandler()).Methods(http.MethodGet)
	router.HandleFunc(webhook.OperationPath, webhook.NewWebHookOperationPostHTTPHandler()).Methods(http.MethodPost)

	notificationHandler := formationnotification.NewHandler(cfg.NotificationConfig)
	// formation assignment notifications sync handlers
	router.HandleFunc("/formation-callback/{tenantId}", notificationHandler.Patch).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/{tenantId}/{applicationId}", notificationHandler.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/no-configuration/{tenantId}", notificationHandler.RespondWithNoConfig).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/no-configuration/{tenantId}/{applicationId}", notificationHandler.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/configuration/{tenantId}", notificationHandler.RespondWithIncomplete).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/configuration/{tenantId}/{applicationId}", notificationHandler.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/with-state/{tenantId}", notificationHandler.PatchWithState).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/with-state/{tenantId}/{applicationId}", notificationHandler.DeleteWithState).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/configuration/redirect-notification/{tenantId}", notificationHandler.RespondWithIncompleteAndRedirectDetails).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/configuration/redirect-notification/{tenantId}/{applicationId}", notificationHandler.RespondWithIncompleteAndRedirectDetails).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/redirect-notification/{tenantId}", notificationHandler.RedirectNotificationHandler).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/fail-once/{tenantId}", notificationHandler.FailOnceResponse).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/fail-once/{tenantId}/{applicationId}", notificationHandler.FailOnceResponse).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/fail/{tenantId}", notificationHandler.FailResponse).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/fail/{tenantId}/{applicationId}", notificationHandler.FailResponse).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/reset-should-fail", notificationHandler.ResetShouldFail).Methods(http.MethodDelete)
	// formation assignment notifications handlers for kyma integration
	router.HandleFunc("/v1/tenants/emptyCredentials", notificationHandler.KymaEmptyCredentials).Methods(http.MethodPatch, http.MethodDelete)
	router.HandleFunc("/v1/tenants/basicCredentials", notificationHandler.KymaBasicCredentials).Methods(http.MethodPatch, http.MethodDelete)
	router.HandleFunc("/v1/tenants/oauthCredentials", notificationHandler.KymaOauthCredentials).Methods(http.MethodPatch, http.MethodDelete)
	// formation assignment notifications async handlers
	router.HandleFunc("/formation-callback/async-old/{tenantId}", notificationHandler.AsyncOld).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async-old/{tenantId}/{applicationId}", notificationHandler.AsyncDelete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/async/{tenantId}", notificationHandler.Async).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async/{tenantId}/{applicationId}", notificationHandler.AsyncDelete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/async-no-config/{tenantId}", notificationHandler.AsyncNoConfig).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async-no-response/{tenantId}", notificationHandler.AsyncNoResponseAssign).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async-no-response/{tenantId}/{applicationId}", notificationHandler.AsyncNoResponseUnassign).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/async-fail-once/{tenantId}", notificationHandler.AsyncFailOnce).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async-fail-once/{tenantId}/{applicationId}", notificationHandler.AsyncFailOnce).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/async-fail/{tenantId}", notificationHandler.AsyncFail).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async-fail/{tenantId}/{applicationId}", notificationHandler.AsyncFail).Methods(http.MethodDelete)
	// formation assignment notifications handler for the destination creation/deletion
	router.HandleFunc("/formation-callback/destinations/configuration/{tenantId}", notificationHandler.RespondWithIncompleteAndDestinationDetails).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/destinations/configuration/{tenantId}/{applicationId}", notificationHandler.DestinationDelete).Methods(http.MethodDelete)
	router.HandleFunc("/formation-callback/async/destinations/{tenantId}", notificationHandler.AsyncDestinationPatch).Methods(http.MethodPatch)
	router.HandleFunc("/formation-callback/async/destinations/{tenantId}/{applicationId}", notificationHandler.AsyncDestinationDelete).Methods(http.MethodDelete)
	// formation(lifecycle) notifications sync handlers
	router.HandleFunc("/v1/businessIntegration/{uclFormationId}", notificationHandler.PostFormation).Methods(http.MethodPost)
	router.HandleFunc("/v1/businessIntegration/{uclFormationId}", notificationHandler.DeleteFormation).Methods(http.MethodDelete)
	router.HandleFunc("/v1/businessIntegration/fail-once/{uclFormationId}", notificationHandler.FailOnceFormation).Methods(http.MethodPost, http.MethodDelete)
	// formation(lifecycle) notifications async handlers
	router.HandleFunc("/v1/businessIntegration/async/{uclFormationId}", notificationHandler.AsyncPostFormation).Methods(http.MethodPost)
	router.HandleFunc("/v1/businessIntegration/async/{uclFormationId}", notificationHandler.AsyncDeleteFormation).Methods(http.MethodDelete)
	router.HandleFunc("/v1/businessIntegration/async-no-response/{uclFormationId}", notificationHandler.AsyncNoResponse).Methods(http.MethodPost, http.MethodDelete)
	router.HandleFunc("/v1/businessIntegration/async-fail-once/{uclFormationId}", notificationHandler.AsyncFormationFailOnce).Methods(http.MethodPost, http.MethodDelete)
	// "technical" handlers for getting/deleting the FA notifications
	router.HandleFunc("/formation-callback", notificationHandler.GetResponses).Methods(http.MethodGet)
	router.HandleFunc("/formation-callback/cleanup", notificationHandler.Cleanup).Methods(http.MethodDelete)

	// destination creator handlers
	destinationCreatorSubaccountLevelPath := cfg.DestinationCreatorConfig.DestinationAPIConfig.SubaccountLevelPath
	deleteDestinationCreatorSubaccountLevelPathSuffix := fmt.Sprintf("/{%s}", cfg.DestinationCreatorConfig.DestinationAPIConfig.DestinationNameParam)
	router.HandleFunc(destinationCreatorSubaccountLevelPath, destinationCreatorHandler.CreateDestinations).Methods(http.MethodPost)
	router.HandleFunc(destinationCreatorSubaccountLevelPath+deleteDestinationCreatorSubaccountLevelPathSuffix, destinationCreatorHandler.DeleteDestinations).Methods(http.MethodDelete)

	destinationCreatorInstanceLevelPath := cfg.DestinationCreatorConfig.DestinationAPIConfig.InstanceLevelPath
	deleteDestinationCreatorInstanceLevelPathSuffix := fmt.Sprintf("/{%s}", cfg.DestinationCreatorConfig.DestinationAPIConfig.DestinationNameParam)
	router.HandleFunc(destinationCreatorInstanceLevelPath, destinationCreatorHandler.CreateDestinations).Methods(http.MethodPost)
	router.HandleFunc(destinationCreatorInstanceLevelPath+deleteDestinationCreatorInstanceLevelPathSuffix, destinationCreatorHandler.DeleteDestinations).Methods(http.MethodDelete)

	certificateSubaccountLevelPath := cfg.DestinationCreatorConfig.CertificateAPIConfig.SubaccountLevelPath
	deleteCertificateSubaccountLevelPathSuffix := fmt.Sprintf("/{%s}", cfg.DestinationCreatorConfig.CertificateAPIConfig.CertificateNameParam)
	router.HandleFunc(certificateSubaccountLevelPath, destinationCreatorHandler.CreateCertificate).Methods(http.MethodPost)
	router.HandleFunc(certificateSubaccountLevelPath+deleteCertificateSubaccountLevelPathSuffix, destinationCreatorHandler.DeleteCertificate).Methods(http.MethodDelete)

	certificateInstanceLevelPath := cfg.DestinationCreatorConfig.CertificateAPIConfig.InstanceLevelPath
	deleteCertificateInstanceLevelPathSuffix := fmt.Sprintf("/{%s}", cfg.DestinationCreatorConfig.CertificateAPIConfig.CertificateNameParam)
	router.HandleFunc(certificateInstanceLevelPath, destinationCreatorHandler.CreateCertificate).Methods(http.MethodPost)
	router.HandleFunc(certificateInstanceLevelPath+deleteCertificateInstanceLevelPathSuffix, destinationCreatorHandler.DeleteCertificate).Methods(http.MethodDelete)

	// "internal technical" handlers for deleting in-memory destinations and destination certificates mappings
	router.HandleFunc("/destinations/cleanup", destinationCreatorHandler.CleanupDestinations).Methods(http.MethodDelete)
	router.HandleFunc("/destination-certificates/cleanup", destinationCreatorHandler.CleanupDestinationCertificates).Methods(http.MethodDelete)

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
	servers = append(servers, initMultiTenantORDServer(cfg))
	servers = append(servers, initOauthSecuredORDServer(cfg, key))
	servers = append(servers, initUnsecuredORDServerWithAdditionalContent(cfg))
	servers = append(servers, initSecuredGlobalRegistryORDServer(cfg))
	servers = append(servers, initUnsecuredGlobalRegistryORDServer(cfg))
	servers = append(servers, initCertSecuredProxyORDServer(cfg))
	return servers
}

func initCertSecuredORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "sap:cmp-mtls:v1"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(cfg.ORDServers.CertSecuredBaseURL, "sap:cmp-mtls:v1"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocument(cfg.ORDServers.CertSecuredBaseURL, "sap:cmp-mtls:v1"))

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

func initMultiTenantORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "open"))
	router.HandleFunc("/test/fullPath", ord_aggregator.HandleFuncOrdConfigWithDocPath(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredMultiTenantPort), "/open-resource-discovery/v1/documents/example2", "open"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredMultiTenantPort), "open"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocument(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredMultiTenantPort), "open"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.UnsecuredMultiTenantPort),
		Handler: router,
	}
}

func initUnsecuredORDServerWithAdditionalContent(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_aggregator.HandleFuncOrdConfig("", "open"))
	router.HandleFunc("/test/fullPath", ord_aggregator.HandleFuncOrdConfigWithDocPath(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredWithAdditionalContentPort), "/open-resource-discovery/v1/documents/example2", "open"))

	testProperties := `"testProperty1": "testValue1", "testProperty2": "testValue2", "testProperty3": "testValue3"`
	additionalTestEntity := fmt.Sprintf(`,"testEntity": { %s }`, testProperties)
	additionalTestProperties := fmt.Sprintf(`,%s`, testProperties)

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocumentWithAdditionalContent(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredWithAdditionalContentPort), "open", additionalTestEntity, additionalTestProperties))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocumentWithAdditionalContent(fmt.Sprintf("%s:%d", cfg.BaseURL, cfg.ORDServers.UnsecuredWithAdditionalContentPort), "open", additionalTestEntity, additionalTestProperties))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.UnsecuredWithAdditionalContentPort),
		Handler: router,
	}
}

func initCertSecuredProxyORDServer(cfg config) *http.Server {
	router := mux.NewRouter()
	proxyPath := "/proxy"
	proxyHeader := "target_host"
	baseURL := fmt.Sprintf("%s:%d%s", cfg.BaseURL, cfg.ORDServers.ProxyPort, proxyPath)

	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exists := r.Header.Get(proxyHeader); exists == "" {
				httphelpers.WriteError(w, errors.Errorf("required header %s is missing in the request", proxyHeader), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	router.HandleFunc(proxyPath, ord_aggregator.HandleFuncOrdConfig(fmt.Sprintf("%s:%d%s", cfg.BaseURL, cfg.ORDServers.ProxyPort, proxyPath), "sap:cmp-mtls:v1"))

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_aggregator.HandleFuncOrdDocument(baseURL, "sap:cmp-mtls:v1"))
	router.HandleFunc("/open-resource-discovery/v1/documents/example2", ord_aggregator.HandleFuncOrdDocument(baseURL, "sap:cmp-mtls:v1"))

	router.HandleFunc("/external-api/spec", apispec.HandleFunc)
	router.HandleFunc("/external-api/spec/flapping", apispec.FlappingHandleFunc())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.ProxyPort),
		Handler: router,
	}
}

func initUnsecuredGlobalRegistryORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/open-resource-discovery", ord_global_registry.HandleFuncOrdConfig(cfg.ORDServers.CertSecuredGlobalBaseURL))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.GlobalRegistryUnsecuredPort),
		Handler: router,
	}
}

func initSecuredGlobalRegistryORDServer(cfg config) *http.Server {
	router := mux.NewRouter()

	router.HandleFunc("/open-resource-discovery/v1/documents/example1", ord_global_registry.HandleFuncOrdDocument())

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ORDServers.GlobalRegistryCertPort),
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

type MiddlewareArgs struct {
	key            *rsa.PublicKey
	validateClaims func(claims jwt.Claims) bool
	ClaimGetter    func() jwt.Claims
}

func oauthMiddlewareMultiple(args []MiddlewareArgs) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(httphelpers.AuthorizationHeaderKey)
			if len(authHeader) == 0 {
				httphelpers.WriteError(w, errors.New("No Authorization header"), http.StatusUnauthorized)
				return
			}
			if !strings.Contains(authHeader, "Bearer") {
				httphelpers.WriteError(w, errors.New("No Bearer token"), http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			for _, middlewareArgs := range args {
				claims := middlewareArgs.ClaimGetter()
				if _, err := jwt.ParseWithClaims(token, claims, func(_ *jwt.Token) (interface{}, error) {
					return middlewareArgs.key, nil
				}); err != nil {
					continue
				}

				if !middlewareArgs.validateClaims(claims) {
					continue
				}

				tenant := getClaimsTenant(claims)

				log.C(r.Context()).Infof("Middleware authenticated successfully. Continue with request.")
				r.Header.Set("tenant", tenant)
				log.C(r.Context()).Infof("Setting tenant %s for tenant header", tenant)

				next.ServeHTTP(w, r)
				return
			}

			httphelpers.WriteError(w, errors.New("Could not validate token"), http.StatusUnauthorized)
			return
		})
	}
}

func oauthMiddleware(key *rsa.PublicKey, validateClaims func(claims Claims) bool) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get(httphelpers.AuthorizationHeaderKey)
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
			r.Header.Set("tenant", parsed.Tenant)
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

	log.C(ctx).Infof("Starting and listening on %s://%s", "http", server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.C(ctx).Fatalf("Could not listen on %s://%s: %v\n", "http", server.Addr, err)
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
			log.C(ctx).Fatal("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	server.SetKeepAlivesEnabled(false)

	if err := server.Shutdown(ctx); err != nil {
		log.C(ctx).Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func noopClaimsValidator(_ Claims) bool {
	return true
}

type Claims interface {
	GetTenant() string
}

func getClaimsValidator(trustedTenants []string) func(jwt.Claims) bool {
	return func(claims jwt.Claims) bool {
		for _, tenant := range trustedTenants {
			claimsTenant := getClaimsTenant(claims)
			if claimsTenant == tenant {
				return true
			}
		}
		return false
	}
}

func getClaimsTenant(claims jwt.Claims) string {
	switch c := claims.(type) {
	case *oauth.Claims:
		return c.GetTenant()
	case *modelJwt.Claims:
		return c.GetTenant()
	default:
		return ""
	}
}
