package resync

import (
	"context"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const (
	jobNamePattern          = "^APP_(.*)_JOB_NAME$"
	regionNamePatternFormat = "^APP_%s_REGIONAL_CONFIG_(.*)_REGION_NAME$"
)

type job struct {
	ctx             context.Context
	name            string
	environmentVars map[string]string
	jobConfig       *JobConfig
}

// JobConfig contains tenant fetcher job configuration read from environment
type JobConfig struct {
	EventsConfig
	ResyncConfig
	KubeConfig

	JobName        string      `envconfig:"JOB_NAME" required:"true"`
	TenantProvider string      `envconfig:"TENANT_PROVIDER" required:"true"`
	TenantType     tenant.Type `envconfig:"TENANT_TYPE" required:"true"`
	RegionPrefix   string      `envconfig:"REGION_PREFIX"`
}

// EventsConfig contains configuration for Events API requests
type EventsConfig struct {
	QueryConfig
	PagingConfig

	RegionalAuthConfigSecret AuthProviderConfig            `envconfig:"SECRET"`
	RegionalAPIConfigs       map[string]*RegionalAPIConfig `ignored:"true"`
	APIConfig                EventsAPIConfig               `envconfig:"API"`
	TenantOperationChunkSize int                                 `envconfig:"TENANT_INSERT_CHUNK_SIZE" default:"500"`
	RetryAttempts            uint                                `envconfig:"RETRY_ATTEMPTS" default:"7"`
}

func (ec EventsConfig) Validate(tenantType tenant.Type) error {
	for region, config := range ec.RegionalAPIConfigs {
		if err := config.Validate(tenantType); err != nil {
			return errors.Wrapf(err, "while validating API configuration for region %s", region)
		}
	}
	return nil
}

type RegionalAPIConfig struct {
	EventsAPIConfig
	RegionName string `envconfig:"REGION_NAME" required:"true"`
}

type EventsAPIConfig struct {
	APIEndpointsConfig
	TenantFieldMapping
	MovedSubaccountsFieldMapping

	AuthConfigSecretKey string         `envconfig:"AUTH_CONFIG_SECRET_KEY"`
	AuthMode            oauth.AuthMode `envconfig:"AUTH_MODE" required:"true"`
	ClientTimeout       time.Duration  `envconfig:"TIMEOUT" default:"1m"`
	OAuthConfig         OAuth2Config   `ignored:"true"`
}

func (c EventsAPIConfig) Validate(tenantType tenant.Type) error {
	missingProperties := make([]string, 0)
	if tenantType == tenant.Subaccount {
		if len(c.APIEndpointsConfig.EndpointSubaccountCreated) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountCreated")
		}
		if len(c.APIEndpointsConfig.EndpointSubaccountUpdated) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountUpdated")
		}
		if len(c.APIEndpointsConfig.EndpointSubaccountMoved) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountMoved")
		}
		if len(c.APIEndpointsConfig.EndpointSubaccountDeleted) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountDeleted")
		}
	}
	if tenantType == tenant.Account {
		if len(c.APIEndpointsConfig.EndpointTenantCreated) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountCreated")
		}
		if len(c.APIEndpointsConfig.EndpointTenantUpdated) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountUpdated")
		}
		if len(c.APIEndpointsConfig.EndpointTenantDeleted) == 0 {
			missingProperties = append(missingProperties, "EndpointSubaccountMoved")
		}
	}
	if len(missingProperties) > 0 {
		return fmt.Errorf("missing API Client config properties: %s", strings.Join(missingProperties, ","))
	}

	return c.OAuthConfig.Validate(c.AuthMode)
}

type ResyncConfig struct {
	TenantFetcherJobIntervalMins time.Duration `envconfig:"TENANT_FETCHER_JOB_INTERVAL" default:"5m"`
	FullResyncInterval           time.Duration `envconfig:"FULL_RESYNC_INTERVAL" default:"12h"`
}

type PagingConfig struct {
	TotalPagesField   string `envconfig:"TENANT_TOTAL_PAGES_FIELD" default:"pages"`
	TotalResultsField string `envconfig:"TENANT_TOTAL_RESULTS_FIELD" default:"totalResults"`
}

// NewTenantFetcherJobEnvironment used for job configuration read from environment
func NewTenantFetcherJobEnvironment(ctx context.Context, name string, environmentVars map[string]string) *job {
	return &job{
		ctx:             ctx,
		name:            name,
		environmentVars: environmentVars,
		jobConfig:       nil,
	}
}

func (jc *JobConfig) Validate() error {
	switch {
	case len(jc.JobName) == 0:
		return errors.New("")
	}
	return nil
}

// ReadJobConfig reads job configuration from environment
func (j *job) ReadJobConfig() (*JobConfig, error) {
	if j.jobConfig != nil {
		return j.jobConfig, nil
	}

	jobConfigPrefix := fmt.Sprintf("APP_%s", strings.ToUpper(j.name))
	jc := JobConfig{}
	if err := envconfig.Process(jobConfigPrefix, &jc); err != nil {
		return nil, errors.Wrapf(err, "while initializing job config with prefix %s", jobConfigPrefix)
	}

	regionalCfg, err := j.readRegionalEventsConfig()
	if err != nil {
		return nil, err
	}

	authConfigs, err := j.mapClientsAuthConfigs(jc)
	if err != nil {
		return nil, err
	}

	clientCfg, ok := authConfigs[jc.APIConfig.AuthConfigSecretKey]
	if !ok {
		return nil, errors.New("auth config not found for Events API")
	}

	jc.APIConfig.OAuthConfig = clientCfg

	for key, authCfg := range authConfigs {
		// assuming that config keys will match the regions of additional API clients
		if cfg, ok := regionalCfg[key]; ok {
			cfg.OAuthConfig = authCfg
		}
	}

	jc.RegionalAPIConfigs = regionalCfg

	if err := jc.Validate(); err != nil {
		return nil, errors.Wrapf(err, "while reading job configuration for job %s", j.name)
	}

	j.jobConfig = &jc
	return j.jobConfig, nil
}

func (j *job) readRegionalEventsConfig() (map[string]*RegionalAPIConfig, error) {
	regEventsConfig := map[string]*RegionalAPIConfig{}
	for _, region := range j.jobRegions() {
		regionalConfigEnvPrefix := fmt.Sprintf("APP_%s_REGIONAL_CONFIG_%s", j.name, strings.ToUpper(region))
		newCfg := &RegionalAPIConfig{}
		if err := envconfig.Process(regionalConfigEnvPrefix, newCfg); err != nil {
			return nil, errors.Wrapf(err, "while reading config for region %s", region)
		}

		regEventsConfig[strings.ToLower(region)] = newCfg
	}

	return regEventsConfig, nil
}

// mapClientsAuthConfigs Parses the InstanceConfigs json string to map with key: region name and value: InstanceConfig for the instance in the region
func (j *job) mapClientsAuthConfigs(jc JobConfig) (map[string]OAuth2Config, error) {
	secretData, err := j.getJobSecret(jc)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting job secret")
	}

	if ok := gjson.Valid(secretData); !ok {
		return nil, errors.New("failed to validate instance configs")
	}

	authConfig := jc.EventsConfig.RegionalAuthConfigSecret
	bindingsResult := gjson.Parse(secretData)
	bindingsMap := bindingsResult.Map()
	clientsAuthConfig := make(map[string]OAuth2Config)
	for secretKey, config := range bindingsMap {
		i := OAuth2Config{
			ClientID:           gjson.Get(config.String(), authConfig.ClientIDPath).String(),
			ClientSecret:       gjson.Get(config.String(), authConfig.ClientSecretPath).String(),
			OAuthTokenEndpoint: gjson.Get(config.String(), authConfig.TokenEndpointPath).String(),
			TokenPath:          authConfig.TokenPath,
			X509Config: X509Config{
				Cert: gjson.Get(config.String(), authConfig.CertPath).String(),
				Key:  gjson.Get(config.String(), authConfig.KeyPath).String(),
			},
		}

		clientsAuthConfig[secretKey] = i
	}

	return clientsAuthConfig, nil
}

func (j *job) getJobSecret(jc JobConfig) (string, error) {
	path := jc.RegionalAuthConfigSecret.SecretFilePath
	if path == "" {
		return "", errors.New("job secret path cannot be empty")
	}
	secret, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read job secret file")
	}

	return string(secret), nil
}

// ReadFromEnvironment returns a key-value map of environment variables
func ReadFromEnvironment(environ []string) map[string]string {
	vars := make(map[string]string)
	for _, env := range environ {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]
		vars[key] = value
	}
	return vars
}

// GetJobNames retrieves the names of tenant fetchers jobs
func GetJobNames(envVars map[string]string) []string {
	searchPattern := regexp.MustCompile(jobNamePattern)
	var jobNames []string

	for key := range envVars {
		matches := searchPattern.FindStringSubmatch(key)
		if len(matches) > 0 {
			jobName := matches[1]
			jobNames = append(jobNames, jobName)
		}
	}

	return jobNames
}

// jobRegions retrieves the names of the tenant fetchers job regions
func (j *job) jobRegions() []string {
	searchPattern := regexp.MustCompile(fmt.Sprintf(regionNamePatternFormat, j.name))

	var regionNames []string
	for key := range j.environmentVars {
		matches := searchPattern.FindStringSubmatch(key)
		if len(matches) > 0 {
			regionName := matches[1]
			regionNames = append(regionNames, regionName)
		}
	}

	return regionNames
}
