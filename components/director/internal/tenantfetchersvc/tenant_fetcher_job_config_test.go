package tenantfetchersvc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	jobName := "testJob"
	accountRegion := "testRegion"
	tenantCreatedEndpoint := "create"
	envVars := map[string]string{"APP_%s_ACCOUNT_REGION": accountRegion, "APP_%s_ENDPOINT_TENANT_CREATED": tenantCreatedEndpoint}

	t.Run("Read events configuration from environment", func(t *testing.T) {
		environ := getEnvironment(envVars, jobName)

		// WHEN
		jobConfig := NewTenantFetcherJobEnvironment(context.TODO(), jobName, ReadFromEnvironment(environ)).ReadJobConfig()
		eventsCfg := jobConfig.GetEventsCgf()

		// THEN
		assert.Equal(t, accountRegion, eventsCfg.AccountsRegion, fmt.Sprintf("Account region should be %s", accountRegion))
		assert.Equal(t, tenantCreatedEndpoint, eventsCfg.APIConfig.EndpointTenantCreated, fmt.Sprintf("Tenant created endpint should be %s", tenantCreatedEndpoint))
	})
}

func TestFetcherJobConfig_ReadDefaultEventsConfig(t *testing.T) {
	// GIVEN
	jobName := "testJob"

	accountRegion := "testRegion"
	defaultAccountRegion := "central"

	tenantCreatedEndpoint := "create"
	defaultTenantCreatedEndpoint := ""

	testCases := []struct {
		Name    string
		JobName string
		EnvVars map[string]string
	}{
		{
			Name:    "Get default events configurations when no environment variables",
			JobName: jobName,
			EnvVars: nil,
		}, {
			Name:    "Get default events configurations when environment variables don't match configuration",
			JobName: jobName,
			EnvVars: map[string]string{"APP2_%s_ACCOUNT_REGION": accountRegion, "APP2_%s_ENDPOINT_TENANT_CREATED": tenantCreatedEndpoint},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			environ := getEnvironment(testCase.EnvVars, testCase.JobName)

			// WHEN
			jobConfig := NewTenantFetcherJobEnvironment(context.TODO(), testCase.JobName, ReadFromEnvironment(environ)).ReadJobConfig()
			eventsCfg := jobConfig.GetEventsCgf()

			// THEN
			assert.Equal(t, defaultAccountRegion, eventsCfg.AccountsRegion, fmt.Sprintf("Default account region should be %s", defaultAccountRegion))
			assert.Equal(t, defaultTenantCreatedEndpoint, eventsCfg.APIConfig.EndpointTenantCreated, fmt.Sprintf("Default tenant created endpint should be %s", defaultTenantCreatedEndpoint))
		})
	}
}

func TestFetcherJobConfig_ReadHandlerConfig(t *testing.T) {
	// GIVEN
	jobName := "testJob"
	jobIntervalMins := 3 * time.Minute
	tenantProvider := "testProvider"
	еnvVars := map[string]string{"APP_%s_TENANT_FETCHER_JOB_INTERVAL": jobIntervalMins.String(), "APP_%s_TENANT_PROVIDER": tenantProvider}

	t.Run("Read handler configuration from environment", func(t *testing.T) {
		environ := getEnvironment(еnvVars, jobName)

		// WHEN
		jobConfig := NewTenantFetcherJobEnvironment(context.TODO(), jobName, ReadFromEnvironment(environ)).ReadJobConfig()
		handlerCfg := jobConfig.GetHandlerCgf()

		// THEN
		assert.Equal(t, jobIntervalMins, handlerCfg.TenantFetcherJobIntervalMins, fmt.Sprintf("Job interval should be %s", jobIntervalMins))
		assert.Equal(t, tenantProvider, handlerCfg.TenantProviderConfig.TenantProvider, fmt.Sprintf("Tenant provider should be %s", tenantProvider))
	})
}

func TestFetcherJobConfig_ReadDefaultHandlerConfig(t *testing.T) {
	// GIVEN
	jobName := "testJob"

	jobIntervalMins := 3 * time.Minute
	defaultJobIntervalMins := 1 * time.Minute

	tenantProvider := "testProvider"
	defaultTenantProvider := "external-provider"

	testCases := []struct {
		Name    string
		JobName string
		EnvVars map[string]string
	}{
		{
			Name:    "Get default handler configurations when no environment variables",
			JobName: jobName,
			EnvVars: nil,
		}, {
			Name:    "Get default handler configurations when environment variables don't match configuration",
			JobName: jobName,
			EnvVars: map[string]string{"APP2_%s_TENANT_FETCHER_JOB_INTERVAL": jobIntervalMins.String(), "APP2_%s_TENANT_PROVIDER": tenantProvider},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			environ := getEnvironment(testCase.EnvVars, testCase.JobName)

			// WHEN
			jobConfig := NewTenantFetcherJobEnvironment(context.TODO(), testCase.JobName, ReadFromEnvironment(environ)).ReadJobConfig()
			handlerCfg := jobConfig.GetHandlerCgf()

			// THEN
			assert.Equal(t, defaultJobIntervalMins, handlerCfg.TenantFetcherJobIntervalMins, fmt.Sprintf("Default job interval should be %s", defaultJobIntervalMins))
			assert.Equal(t, defaultTenantProvider, handlerCfg.TenantProviderConfig.TenantProvider, fmt.Sprintf("Default tenant provider should be %s", defaultTenantProvider))
		})
	}
}

func getEnvironment(envVars map[string]string, jobName string) []string {
	var environ []string
	for nameFormat, value := range envVars {
		varName := fmt.Sprintf(nameFormat, jobName)
		environ = append(environ, varName+"="+value)
	}
	return environ
}
