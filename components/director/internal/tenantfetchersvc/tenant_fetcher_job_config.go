package tenantfetchersvc

import (
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
)

type job struct {
	name            string
	environmentVars map[string]string
}

// NewTenantFetcherJob creates new tenant fetcher job config
func NewTenantFetcherJob(name string, environmentVars map[string]string) *job {
	return &job{
		name:            name,
		environmentVars: environmentVars,
	}
}

type JobConfig struct {
	EventsConfig  EventsConfig
	HandlerConfig HandlerConfig
}

func (j *job) NewJobConfig() *JobConfig {
	return &JobConfig{
		EventsConfig:  j.getEventsConfig(),
		HandlerConfig: j.getHandlerConfig(),
	}
}

type Handler2Config struct {
	DirectorGraphQLEndpoint     string        `envconfig:"APP_DIRECTOR_GRAPHQL_ENDPOINT"`
	ClientTimeout               time.Duration `envconfig:"default=60s"`
	HTTPClientSkipSslValidation bool          `envconfig:"APP_HTTP_CLIENT_SKIP_SSL_VALIDATION,default=false"`

	TenantProviderConfig
	features.Config

	Database persistence.DatabaseConfig

	Kubernetes          tenantfetcher.KubeConfig
	MetricsPushEndpoint string `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`

	TenantFetcherJobIntervalMins int           `envconfig:"default=5,APP_TENANT_FETCH_JOB_INTERVAL_MINS"`
	ShouldSyncSubaccounts        bool          `envconfig:"default=false,APP_SYNC_SUBACCOUNTS"`
	FullResyncInterval           time.Duration `envconfig:"default=12h"`
	TenantInsertChunkSize        int           `envconfig:"default=500,APP_TENANT_INSERT_CHUNK_SIZE"`
}

func (j *job) getEventsConfig() EventsConfig {
	return EventsConfig{
		AccountsRegion:              j.getEnvValueForKey("central", "APP_"+j.name+"_ACCOUNT_REGION"),
		SubaccountRegions:           strings.Split(j.getEnvValueForKey("central", "APP_"+j.name+"_SUBACCOUNT_REGIONS"), ","),
		MovedSubaccountFieldMapping: j.getMovedSubaccountsFieldMapping(),

		OAuthConfig:        j.getOAuth2Config(),
		APIConfig:          j.getAPIConfig(),
		AuthMode:           oauth.AuthMode(j.getEnvValueForKey("standard", "APP_"+j.name+"_OAUTH_AUTH_MODE")),
		QueryConfig:        j.getQueryConfig(),
		TenantFieldMapping: j.getTenantFieldMapping(),
	}
}

func (j *job) getHandlerConfig() HandlerConfig {
	return HandlerConfig{
		DirectorGraphQLEndpoint:     j.getEnvValueForKey("", "APP_DIRECTOR_GRAPHQL_ENDPOINT"),
		ClientTimeout:               getDuration(j.getEnvValueForKey("60s", ""), time.Duration(60)*time.Second),
		HTTPClientSkipSslValidation: getBoolVal(j.getEnvValueForKey("false", "APP_"+j.name+"_HTTP_CLIENT_SKIP_SSL_VALIDATION"), false),

		TenantProviderConfig: j.getTenantProviderConfig(),
		Features:             j.getFeaturesConfig(),

		Database: j.getDatabaseConfig(),

		Kubernetes:          j.getKubeConfig(),
		MetricsPushEndpoint: j.getEnvValueForKey("", "APP_METRICS_PUSH_ENDPOINT"),

		TenantFetcherJobIntervalMins: getIntVal(j.getEnvValueForKey("5", "APP_"+j.name+"_TENANT_FETCH_JOB_INTERVAL_MINS"), 5),
		ShouldSyncSubaccounts:        getBoolVal(j.getEnvValueForKey("false", "APP_SYNC_SUBACCOUNTS"), false),
		FullResyncInterval:           getDuration(j.getEnvValueForKey("12h", ""), 12*time.Hour),
		TenantInsertChunkSize:        getIntVal(j.getEnvValueForKey("500", "APP_TENANT_INSERT_CHUNK_SIZE"), 500),
	}
}

func (j *job) getMovedSubaccountsFieldMapping() tenantfetcher.MovedSubaccountsFieldMapping {
	return tenantfetcher.MovedSubaccountsFieldMapping{
		LabelValue:   j.getEnvValueForKey("", "APP_"+j.name+"_MAPPING_FIELD_ID"),
		SourceTenant: j.getEnvValueForKey("", "APP_"+j.name+"_MOVED_SUBACCOUNT_SOURCE_TENANT_FIELD"),
		TargetTenant: j.getEnvValueForKey("", "APP_"+j.name+"_MOVED_SUBACCOUNT_TARGET_TENANT_FIELD"),
	}
}

func (j *job) getOAuth2Config() tenantfetcher.OAuth2Config {
	return tenantfetcher.OAuth2Config{
		ClientID:           j.getEnvValueForKey("", "APP_"+j.name+"_CLIENT_ID"),
		ClientSecret:       j.getEnvValueForKey("", "APP_"+j.name+"_CLIENT_SECRET"),
		OAuthTokenEndpoint: j.getEnvValueForKey("", "APP_"+j.name+"_OAUTH_TOKEN_ENDPOINT"),
		TokenPath:          j.getEnvValueForKey("", "APP_"+j.name+"_OAUTH_TOKEN_PATH"),
		SkipSSLValidation:  getBoolVal(j.getEnvValueForKey("false", "APP_"+j.name+"_OAUTH_SKIP_SSL_VALIDATION"), false),
		X509Config: oauth.X509Config{
			Cert: j.getEnvValueForKey("", "APP_"+j.name+"_OAUTH_X509_CERT"),
			Key:  j.getEnvValueForKey("", "APP_"+j.name+"OAUTH_X509_KEY"),
		},
	}
}

func (j *job) getAPIConfig() tenantfetcher.APIConfig {
	return tenantfetcher.APIConfig{
		EndpointTenantCreated:     j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_TENANT_CREATED"),
		EndpointTenantDeleted:     j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_TENANT_DELETED"),
		EndpointTenantUpdated:     j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_TENANT_UPDATED"),
		EndpointSubaccountCreated: j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_SUBACCOUNT_CREATED"),
		EndpointSubaccountDeleted: j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_SUBACCOUNT_DELETED"),
		EndpointSubaccountUpdated: j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_SUBACCOUNT_UPDATED"),
		EndpointSubaccountMoved:   j.getEnvValueForKey("", "APP_"+j.name+"_ENDPOINT_SUBACCOUNT_MOVED"),
	}
}

func (j *job) getQueryConfig() tenantfetcher.QueryConfig {
	return tenantfetcher.QueryConfig{
		PageNumField:    j.getEnvValueForKey("pageNum", "APP_"+j.name+"_QUERY_PAGE_NUM_FIELD"),
		PageSizeField:   j.getEnvValueForKey("pageSize", "APP_"+j.name+"_QUERY_PAGE_SIZE_FIELD"),
		TimestampField:  j.getEnvValueForKey("timestamp", "APP_"+j.name+"_QUERY_TIMESTAMP_FIELD"),
		RegionField:     j.getEnvValueForKey("region", "APP_"+j.name+"_QUERY_REGION_FIELD"),
		PageStartValue:  j.getEnvValueForKey("0", "APP_"+j.name+"_QUERY_PAGE_START"),
		PageSizeValue:   j.getEnvValueForKey("150", "APP_"+j.name+"_QUERY_PAGE_SIZE"),
		SubaccountField: j.getEnvValueForKey("entityId", "APP_"+j.name+"_QUERY_ENTITY_FIELD"),
	}
}

func (j *job) getTenantFieldMapping() tenantfetcher.TenantFieldMapping {
	return tenantfetcher.TenantFieldMapping{
		TotalPagesField:   j.getEnvValueForKey("", "APP_"+j.name+"_TENANT_TOTAL_PAGES_FIELD"),
		TotalResultsField: j.getEnvValueForKey("", "APP_"+j.name+"_TENANT_TOTAL_RESULTS_FIELD"),
		EventsField:       j.getEnvValueForKey("", "APP_"+j.name+"_TENANT_EVENTS_FIELD"),

		NameField:              j.getEnvValueForKey("name", "APP_"+j.name+"_MAPPING_FIELD_NAME"),
		IDField:                j.getEnvValueForKey("id", "APP_"+j.name+"_MAPPING_FIELD_ID"),
		GlobalAccountGUIDField: j.getEnvValueForKey("globalAccountGUID", ""),
		SubaccountIDField:      j.getEnvValueForKey("subaccountId", ""),
		SubaccountGUIDField:    j.getEnvValueForKey("subaccountGuid", ""),
		CustomerIDField:        j.getEnvValueForKey("customerId", "APP_"+j.name+"_MAPPING_FIELD_CUSTOMER_ID"),
		SubdomainField:         j.getEnvValueForKey("subdomain", "APP_"+j.name+"_MAPPING_FIELD_SUBDOMAIN"),
		DetailsField:           j.getEnvValueForKey("details", "APP_"+j.name+"_MAPPING_FIELD_DETAILS"),
		DiscriminatorField:     j.getEnvValueForKey("", "APP_"+j.name+"_MAPPING_FIELD_DISCRIMINATOR"),
		DiscriminatorValue:     j.getEnvValueForKey("", "APP_"+j.name+"_MAPPING_VALUE_DISCRIMINATOR"),

		RegionField:     j.getEnvValueForKey("", "APP_"+j.name+"_MAPPING_FIELD_REGION"),
		EntityIDField:   j.getEnvValueForKey("entityId", "APP_"+j.name+"_MAPPING_FIELD_ENTITY_ID"),
		EntityTypeField: j.getEnvValueForKey("entityType", "APP_"+j.name+"_MAPPING_FIELD_ENTITY_TYPE"),

		GlobalAccountKey: j.getEnvValueForKey("gaID", "APP_"+j.name+"_GLOBAL_ACCOUNT_KEY"),
	}
}

func (j *job) getTenantProviderConfig() TenantProviderConfig {
	return TenantProviderConfig{
		TenantIDProperty:               j.getEnvValueForKey("tenantId", "APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"),
		SubaccountTenantIDProperty:     j.getEnvValueForKey("subaccountTenantId", "APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"),
		CustomerIDProperty:             j.getEnvValueForKey("customerId", "APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"),
		SubdomainProperty:              j.getEnvValueForKey("subdomain", "APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"),
		TenantProvider:                 j.getEnvValueForKey("external-provider", "APP_"+j.name+"_TENANT_PROVIDER"),
		SubscriptionProviderIDProperty: j.getEnvValueForKey("subscriptionProviderId", "APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY"),
	}
}

func (j *job) getFeaturesConfig() features.Config {
	return features.Config{
		DefaultScenarioEnabled:        getBoolVal(j.getEnvValueForKey("true", "APP_DEFAULT_SCENARIO_ENABLED"), true),
		ProtectedLabelPattern:         j.getEnvValueForKey(".*_defaultEventing|^consumer_subaccount_ids$", ""),
		ImmutableLabelPattern:         j.getEnvValueForKey("^xsappnameCMPClone$", "APP_SELF_REGISTER_LABEL_KEY_PATTERN"),
		SubscriptionProviderLabelKey:  j.getEnvValueForKey("subscriptionProviderId", "APP_SUBSCRIPTION_PROVIDER_LABEL_KEY"),
		ConsumerSubaccountIDsLabelKey: j.getEnvValueForKey("consumer_subaccount_ids", "APP_CONSUMER_SUBACCOUNT_IDS_LABEL_KEY"),
	}
}

func (j *job) getDatabaseConfig() persistence.DatabaseConfig {
	return persistence.DatabaseConfig{
		User:               j.getEnvValueForKey("postgres", "APP_DB_USER"),
		Password:           j.getEnvValueForKey("pgsql@12345", "APP_DB_PASSWORD"),
		Host:               j.getEnvValueForKey("localhost", "APP_DB_HOST"),
		Port:               j.getEnvValueForKey("5432", "APP_DB_PORT"),
		Name:               j.getEnvValueForKey("compass", "APP_DB_NAME"),
		SSLMode:            j.getEnvValueForKey("disable", "APP_DB_SSL"),
		MaxOpenConnections: getIntVal(j.getEnvValueForKey("2", "APP_"+j.name+"_DB_MAX_OPEN_CONNECTIONS"), 2),
		MaxIdleConnections: getIntVal(j.getEnvValueForKey("2", "APP_"+j.name+"_DB_MAX_IDLE_CONNECTIONS"), 2),
		ConnMaxLifetime:    getDuration(j.getEnvValueForKey("30m", "APP_DB_CONNECTION_MAX_LIFETIME"), 30*time.Minute),
	}
}

func (j *job) getKubeConfig() tenantfetcher.KubeConfig {
	return tenantfetcher.KubeConfig{
		UseKubernetes:                 j.getEnvValueForKey("true", "APP_"+j.name+"_USE_KUBERNETES"),
		ConfigMapNamespace:            j.getEnvValueForKey("compass-system", "APP_"+j.name+"_CONFIGMAP_NAMESPACE"),
		ConfigMapName:                 j.getEnvValueForKey("tenant-fetcher-config", "APP_"+j.name+"_LAST_EXECUTION_TIME_CONFIG_MAP_NAME"),
		ConfigMapTimestampField:       j.getEnvValueForKey("lastConsumedTenantTimestamp", "APP_"+j.name+"_CONFIGMAP_TIMESTAMP_FIELD"),
		ConfigMapResyncTimestampField: j.getEnvValueForKey("lastFullResyncTimestamp", "APP_"+j.name+"_CONFIGMAP_RESYNC_TIMESTAMP_FIELD"),
	}
}

func (j *job) getEnvValueForKey(defaultVal string, key string) string {
	if val, ok := j.environmentVars[key]; ok {
		return val
	}
	return defaultVal
}

func getBoolVal(value string, defaultVal bool) bool {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return defaultVal
	}
	return v
}

func getIntVal(value string, defaultVal int) int {
	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal
	}
	return v
}

func getDuration(val string, defaultVal time.Duration) time.Duration {
	v, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return v
}

func ReadEnvironmentVars() map[string]string {
	vars := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]
		vars[key] = value
	}
	return vars
}
