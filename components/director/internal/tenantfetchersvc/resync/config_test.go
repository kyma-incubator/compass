package resync

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	JobName               = "testJob"
	TenantCreatedEndpoint = "https://tenantsregistry/v1/events/created"
)

func TestFetcherJobConfig_ReadEnvVars(t *testing.T) {
	// GIVEN
	envValues := map[string]string{}
	envValues["k1"] = "v1"
	envValues["k2"] = "v2"
	envValues["k3"] = "v3"

	t.Run("Read environment variables", func(t *testing.T) {
		var environ []string
		for k, v := range envValues {
			environ = append(environ, k+"="+v)
		}

		// WHEN
		envVars := ReadFromEnvironment(environ)

		// THEN
		for k, v := range envValues {
			assert.NotNil(t, envVars[k], "Environment variables should contain: "+k)
			assert.Equal(t, v, envVars[k], fmt.Sprintf("Value of environment variable %s should be %s", k, v))
		}
	})
}

func TestFetcherJobConfig_GetJobsNames(t *testing.T) {
	// GIVEN
	jobNames := []string{"job1", "job2", "job3"}

	testCases := []struct {
		Name           string
		JobNames       []string
		JobNamePattern string
		ReadSuccess    bool
	}{
		{
			Name:           "Success getting tenant fetcher jobs names",
			JobNames:       jobNames,
			JobNamePattern: "APP_%s_JOB_NAME",
			ReadSuccess:    true,
		}, {
			Name:           "Failure getting tenant fetcher jobs names with wrong environment variable format",
			JobNames:       jobNames,
			JobNamePattern: "APP_WRONG_%s_JOB_NAME",
			ReadSuccess:    false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			var environ []string
			for _, name := range testCase.JobNames {
				varName := fmt.Sprintf(testCase.JobNamePattern, name)
				environ = append(environ, varName+"="+name)
			}

			// WHEN
			jobNamesFromEnv := GetJobNames(ReadFromEnvironment(environ))

			// THEN
			if testCase.ReadSuccess == true {
				for _, name := range testCase.JobNames {
					assert.Contains(t, jobNamesFromEnv, name, "Job names should contain: "+name)
				}
			} else {
				for _, name := range testCase.JobNames {
					assert.NotContains(t, jobNamesFromEnv, name, "Job names not expected to contain: "+name)
				}
			}
		})
	}
}

func TestFetcherJobConfig_ReadEventsConfig(t *testing.T) {
	// GIVEN
	envVars := map[string]string{"APP_%s_UNIVERSAL_CLIENT_API_CONFIG_ENDPOINT_TENANT_CREATED": TenantCreatedEndpoint}

	t.Run("Read events configuration from environment", func(t *testing.T) {
		environ := initEnvWithMandatory(t, envVars, JobName)

		// WHEN
		jobConfig, err := NewTenantFetcherJobEnvironment(context.TODO(), JobName, ReadFromEnvironment(environ)).ReadJobConfig()
		require.NoError(t, err)
		eventsCfg := jobConfig.EventsConfig

		// THEN
		assert.Equal(t, JobName, jobConfig.JobName, fmt.Sprintf("Job name should be %s", JobName))
		assert.Equal(t, TenantCreatedEndpoint, eventsCfg.APIConfig.APIEndpointsConfig.EndpointTenantCreated, fmt.Sprintf("Tenant created endpint should be %s", TenantCreatedEndpoint))
	})
}

func TestFetcherJobConfig_ReadDefaultEventsConfig(t *testing.T) {
	// GIVEN
	defaultCustomerIDField := "customerId"
	customerID := "123"
	defaultTenantCreatedEndpoint := ""
	tenantCreatedEndpoint := "created"
	testCases := []struct {
		Name    string
		JobName string
		EnvVars map[string]string
	}{
		{
			Name:    "Get default events configurations when no environment variables",
			JobName: JobName,
			EnvVars: nil,
		}, {
			Name:    "Get default events configurations when environment variables don't match configuration",
			JobName: JobName,
			EnvVars: map[string]string{"APP2_%s_MAPPING_FIELD_CUSTOMER_ID": customerID, "APP2_%s_UNIVERSAL_CLIENT_API_CONFIG_ENDPOINT_TENANT_CREATED": tenantCreatedEndpoint},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			environ := initEnvWithMandatory(t, testCase.EnvVars, testCase.JobName)

			// WHEN
			jobConfig, err := NewTenantFetcherJobEnvironment(context.TODO(), testCase.JobName, ReadFromEnvironment(environ)).ReadJobConfig()
			require.NoError(t, err)
			eventsCfg := jobConfig.EventsConfig

			// THEN
			assert.Equal(t, defaultCustomerIDField, eventsCfg.APIConfig.CustomerIDField, fmt.Sprintf("Default customer ID field should be %s", defaultCustomerIDField))
			assert.Equal(t, defaultTenantCreatedEndpoint, eventsCfg.APIConfig.APIEndpointsConfig.EndpointSubaccountCreated, fmt.Sprintf("Default tenant created endpint should be %s", defaultTenantCreatedEndpoint))
		})
	}
}

func TestFetcherJobConfig_ReadJobConfig(t *testing.T) {
	// GIVEN
	jobIntervalMins := 3 * time.Minute
	tenantProvider := "testProvider"
	envVars := map[string]string{"APP_%s_TENANT_FETCHER_JOB_INTERVAL": jobIntervalMins.String(), "APP_%s_TENANT_PROVIDER": tenantProvider}

	t.Run("Read handler configuration from environment", func(t *testing.T) {
		environ := initEnvWithMandatory(t, envVars, JobName)

		// WHEN
		jobConfig, err := NewTenantFetcherJobEnvironment(context.TODO(), JobName, ReadFromEnvironment(environ)).ReadJobConfig()
		require.NoError(t, err)

		// THEN
		assert.Equal(t, jobIntervalMins, jobConfig.TenantFetcherJobIntervalMins, fmt.Sprintf("Job interval should be %s", jobIntervalMins))
		assert.Equal(t, tenantProvider, jobConfig.TenantProvider, fmt.Sprintf("Tenant provider should be %s", tenantProvider))
	})
}

func TestFetcherJobConfig_ReadDefaultJobConfig(t *testing.T) {
	// GIVEN
	jobIntervalMins := 3 * time.Minute
	defaultJobIntervalMins := 5 * time.Minute

	tenantProvider := "testProvider"
	defaultTenantProvider := "external-provider"

	testCases := []struct {
		Name    string
		JobName string
		EnvVars map[string]string
	}{
		{
			Name:    "Get default handler configurations when no environment variables",
			JobName: JobName,
			EnvVars: nil,
		}, {
			Name:    "Get default handler configurations when environment variables don't match configuration",
			JobName: JobName,
			EnvVars: map[string]string{"APP2_%s_TENANT_FETCHER_JOB_INTERVAL": jobIntervalMins.String(), "APP2_%s_TENANT_PROVIDER": tenantProvider},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			environ := initEnvWithMandatory(t, testCase.EnvVars, testCase.JobName)

			// WHEN
			jobConfig, err := NewTenantFetcherJobEnvironment(context.TODO(), testCase.JobName, ReadFromEnvironment(environ)).ReadJobConfig()
			require.NoError(t, err)

			// THEN
			assert.Equal(t, defaultJobIntervalMins, jobConfig.TenantFetcherJobIntervalMins, fmt.Sprintf("Default job interval should be %s", defaultJobIntervalMins))
			assert.Equal(t, defaultTenantProvider, jobConfig.TenantProvider, fmt.Sprintf("Default tenant provider should be %s", defaultTenantProvider))
		})
	}
}

func initEnvWithMandatory(t *testing.T, envVars map[string]string, jobName string) []string {
	os.Clearenv()
	mandatoryEnvVars := map[string]string{
		"APP_%s_JOB_NAME":                   jobName,
		"APP_%s_TENANT_PROVIDER":            "external-provider",
		"APP_%s_TENANT_TYPE":                "subaccount",
		"APP_%s_UNIVERSAL_CLIENT_AUTH_MODE": "standard",
	}
	environ := initEnv(t, mandatoryEnvVars, strings.ToUpper(jobName))
	return append(environ, initEnv(t, envVars, strings.ToUpper(jobName))...)
}

func initEnv(t *testing.T, envVars map[string]string, jobName string) []string {
	environ := make([]string, 0, len(envVars))
	for nameFormat, value := range envVars {
		varName := fmt.Sprintf(nameFormat, jobName)
		environ = append(environ, varName+"="+value)
		err := os.Setenv(varName, value)
		require.NoError(t, err)
	}
	return environ
}
