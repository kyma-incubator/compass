package tenantfetchersvc

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
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

	JobName        string      `envconfig:"JOB_NAME"`
	TenantProvider string      `envconfig:"TENANT_PROVIDER"`
	TenantType     tenant.Type `envconfig:"TENANT_TYPE"`
}

// EventsConfig contains configuration for Events API requests
type EventsConfig struct {
	QueryConfig
	TenantFieldMapping
	MovedSubaccountsFieldMapping

	RegionalAPIConfigs    map[string]EventAPIRegionalConfig `ignored:"true"`
	UniversalClientConfig APIClientConfig                   `envconfig:"UNIVERSAL_CLIENT"`
	TenantInsertChunkSize int                               `envconfig:"TENANT_INSERT_CHUNK_SIZE" default:"500"`
	RetryAttempts         uint                              `envconfig:"RETRY_ATTEMPTS" default:"7"`
}

type EventAPIRegionalConfig struct {
	APIClientConfig
	UniversalClientEnabled bool   `envconfig:"UNIVERSAL_CLIENT_ENABLED"` // weather the universal client should also fetch events for that region
	RegionName             string `envconfig:"REGION_NAME"`
	RegionPrefix           string `envconfig:"REGION_PREFIX"`
}

type APIClientConfig struct {
	AuthMode    oauth.AuthMode `envconfig:"AUTH_MODE"`
	OAuthConfig OAuth2Config   `envconfig:"OAUTH_CONFIG"`
	APIConfig   APIConfig      `envconfig:"API_CONFIG"`
}

type ResyncConfig struct {
	TenantFetcherJobIntervalMins time.Duration `default:"5m" envconfig:"TENANT_FETCHER_JOB_INTERVAL"`
	FullResyncInterval           time.Duration `default:"12h" envconfig:"FULL_RESYNC_INTERVAL"`
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

// ReadJobConfig reads job configuration from environment
func (j *job) ReadJobConfig() (*JobConfig, error) {
	if j.jobConfig != nil {
		return j.jobConfig, nil
	}

	jobConfigPrefix := fmt.Sprintf("APP_%s", j.name)
	jc := JobConfig{}
	if err := envconfig.Process(jobConfigPrefix, &jc); err != nil {
		return nil, errors.Wrapf(err, "while initializing job config with prefix %s", jobConfigPrefix)
	}

	jc.RegionalAPIConfigs = j.readRegionalEventsConfig()
	j.jobConfig = &jc
	return j.jobConfig, nil
}

func (j *job) readRegionalEventsConfig() map[string]EventAPIRegionalConfig {
	regEventsConfig := map[string]EventAPIRegionalConfig{}
	for _, region := range j.jobRegions() {
		regionalConfigEnvPrefix := fmt.Sprintf("APP_%s_API_CLIENT_CONFIG_%s", j.name, strings.ToUpper(region))
		newCfg := &EventAPIRegionalConfig{}
		if err := envconfig.Process(regionalConfigEnvPrefix, newCfg); err != nil {
			// TODO throw error
			log.D().Errorf("Failed to read config: %v", err)
			return nil
		}

		regEventsConfig[strings.ToLower(region)] = *newCfg
	}

	return regEventsConfig
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
