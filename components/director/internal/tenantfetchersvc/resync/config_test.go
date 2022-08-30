package resync_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	JobName               = "testJob"
	TenantCreatedEndpoint = "https://tenantsregistry/v1/events/created"
	TenantUpdatedEndpoint = "https://tenantsregistry/v1/events/updated"
	TenantMovedEndpoint   = "https://tenantsregistry/v1/events/moved"
	TenantDeletedEndpoint = "https://tenantsregistry/v1/events/deleted"
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
		envVars := resync.ReadFromEnvironment(environ)

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
			jobNamesFromEnv := resync.GetJobNames(resync.ReadFromEnvironment(environ))

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

func TestJobConfig_ReadJobConfig(t *testing.T) {
	t.Run("Successfully reads events configuration from environment", func(t *testing.T) {
		environ := initEnvWithMandatory(t, nil, JobName, tenant.Subaccount, oauth.Standard)

		// WHEN
		jobConfig, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.NoError(t, err)
		eventsCfg := jobConfig.EventsConfig
		assert.Equal(t, JobName, jobConfig.JobName, fmt.Sprintf("Job name should be %s", JobName))
		assert.Equal(t, TenantCreatedEndpoint, eventsCfg.APIConfig.APIEndpointsConfig.EndpointSubaccountCreated, fmt.Sprintf("Tenant created endpint should be %s", TenantCreatedEndpoint))
	})
	t.Run("Successfully reads regional events configuration from environment", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		region := "EU-1"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_REGION_NAME"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_CONFIG_SECRET_KEY"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_MODE"] = "standard"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_CREATED"] = TenantCreatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_UPDATED"] = TenantUpdatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_MOVED"] = TenantMovedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_DELETED"] = TenantDeletedEndpoint

		environ := initEnv(t, envs, JobName)

		// WHEN
		jobConfig, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.NoError(t, err)
		eventsCfg := jobConfig.EventsConfig
		assert.Equal(t, JobName, jobConfig.JobName, fmt.Sprintf("Job name should be %s", JobName))
		assert.Equal(t, TenantCreatedEndpoint, eventsCfg.APIConfig.APIEndpointsConfig.EndpointSubaccountCreated, fmt.Sprintf("Tenant created endpint should be %s", TenantCreatedEndpoint))
		assert.Equal(t, TenantCreatedEndpoint, eventsCfg.RegionalAPIConfigs[strings.ToLower(region)].APIEndpointsConfig.EndpointSubaccountCreated, fmt.Sprintf("Regional tenant created endpint should be %s", TenantCreatedEndpoint))
	})
	t.Run("Returns an error when mandatory env var is missing", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		jobNameKey := "APP_%s_JOB_NAME"
		delete(envs, jobNameKey)
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "missing value")
		require.Contains(t, err.Error(), fmt.Sprintf(jobNameKey, strings.ToUpper(JobName)))
	})
	t.Run("Returns an error when mandatory env var is missing from regional config", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		region := "EU-1"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_REGION_NAME"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_CONFIG_SECRET_KEY"] = strings.ToLower(region)

		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_CREATED"] = TenantCreatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_UPDATED"] = TenantUpdatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_MOVED"] = TenantMovedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_DELETED"] = TenantDeletedEndpoint

		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Contains(t, err.Error(), "AUTH_MODE")
	})
	t.Run("Returns an error when mandatory API Config is from regional config", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		region := "EU-1"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_REGION_NAME"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_CONFIG_SECRET_KEY"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_MODE"] = "standard"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_CREATED"] = TenantCreatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_UPDATED"] = TenantUpdatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_MOVED"] = TenantMovedEndpoint

		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Contains(t, err.Error(), "missing API Client config properties: EndpointSubaccountDeleted")
	})
	t.Run("Returns an error when Auth Config cannot be found for region", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		region := "EU-1"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_REGION_NAME"] = strings.ToLower(region)
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_MODE"] = "standard"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_CREATED"] = TenantCreatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_UPDATED"] = TenantUpdatedEndpoint
		envs["APP_%s_REGIONAL_CONFIG_EU-1_ENDPOINT_SUBACCOUNT_MOVED"] = TenantMovedEndpoint

		secretKey := "missing_key"
		envs["APP_%s_REGIONAL_CONFIG_EU-1_AUTH_CONFIG_SECRET_KEY"] = secretKey
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Contains(t, err.Error(), fmt.Sprintf("secret file does not contain key %s", secretKey))
	})
	t.Run("Returns an error when Auth Config cannot be found for central API", func(t *testing.T) {
		secretKey := "missing_key"
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		envs["APP_%s_API_AUTH_CONFIG_SECRET_KEY"] = secretKey
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Contains(t, err.Error(), fmt.Sprintf("auth config not found for Events API: secret file does not contain key %s", secretKey))
	})
	t.Run("Returns an error when secrets file cannot be read", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		envs["APP_%s_SECRET_FILE_PATH"] = "testdata/notfound.json"
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to read job secret file")
	})
	t.Run("Returns an error when secrets file is not provided", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		envs["APP_%s_SECRET_FILE_PATH"] = ""
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "job secret path cannot be empty")
	})
	t.Run("Returns an error when secrets file is not a valid JSON", func(t *testing.T) {
		envs := mandatoryEnvVars(JobName, tenant.Subaccount, oauth.Standard)
		envs["APP_%s_SECRET_FILE_PATH"] = "testdata/invalid.json"
		environ := initEnv(t, envs, JobName)

		// WHEN
		_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to validate job auth configs")
	})

	var apiEndpointsCfgTestCase = []struct {
		name            string
		tenantType      tenant.Type
		missingProperty string
		propInErrorMsg  string
	}{
		{
			name:            "Fails when created endpoint is not provided for subaccount tenant type",
			tenantType:      tenant.Subaccount,
			missingProperty: "APP_%s_API_ENDPOINT_SUBACCOUNT_CREATED",
			propInErrorMsg:  "EndpointSubaccountCreated",
		},
		{
			name:            "Fails when updated endpoint is not provided for subaccount tenant type",
			tenantType:      tenant.Subaccount,
			missingProperty: "APP_%s_API_ENDPOINT_SUBACCOUNT_UPDATED",
			propInErrorMsg:  "EndpointSubaccountUpdated",
		},
		{
			name:            "Fails when moved endpoint is not provided for subaccount tenant type",
			tenantType:      tenant.Subaccount,
			missingProperty: "APP_%s_API_ENDPOINT_SUBACCOUNT_MOVED",
			propInErrorMsg:  "EndpointSubaccountMoved",
		},
		{
			name:            "Fails when deleted endpoint is not provided for subaccount tenant type",
			tenantType:      tenant.Subaccount,
			missingProperty: "APP_%s_API_ENDPOINT_SUBACCOUNT_DELETED",
			propInErrorMsg:  "EndpointSubaccountDeleted",
		},
		{
			name:            "Fails when created endpoint is not provided for account tenant type",
			tenantType:      tenant.Account,
			missingProperty: "APP_%s_API_ENDPOINT_TENANT_CREATED",
			propInErrorMsg:  "EndpointTenantCreated",
		},
		{
			name:            "Fails when updated endpoint is not provided for account tenant type",
			tenantType:      tenant.Account,
			missingProperty: "APP_%s_API_ENDPOINT_TENANT_UPDATED",
			propInErrorMsg:  "EndpointTenantUpdated",
		},
		{
			name:            "Fails when deleted endpoint is not provided for account tenant type",
			tenantType:      tenant.Account,
			missingProperty: "APP_%s_API_ENDPOINT_TENANT_DELETED",
			propInErrorMsg:  "EndpointTenantDeleted",
		},
	}
	for _, tc := range apiEndpointsCfgTestCase {
		t.Run(tc.name, func(t *testing.T) {
			envs := mandatoryEnvVars(JobName, tc.tenantType, oauth.Standard)
			envs["APP_%s_TENANT_TYPE"] = string(tc.tenantType)
			delete(envs, tc.missingProperty)
			environ := initEnv(t, envs, JobName)

			// WHEN
			_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()
			require.Error(t, err)
			require.Contains(t, err.Error(), fmt.Sprintf("missing API Client config properties: %s", tc.propInErrorMsg))
		})
	}

	var authCfgTestCase = []struct {
		name            string
		authMode        oauth.AuthMode
		missingProperty string
		propInErrorMsg  string
	}{
		{
			name:            "Fails when client ID is not provided",
			authMode:        oauth.Standard,
			missingProperty: "APP_%s_SECRET_CLIENT_ID_PATH",
			propInErrorMsg:  "ClientID",
		},
		{
			name:            "Fails when client secret is not provided",
			authMode:        oauth.Standard,
			missingProperty: "APP_%s_SECRET_CLIENT_SECRET_PATH",
			propInErrorMsg:  "ClientSecret",
		},
		{
			name:            "Fails when token endpoint is not provided",
			authMode:        oauth.Standard,
			missingProperty: "APP_%s_SECRET_TOKEN_ENDPOINT_PATH",
			propInErrorMsg:  "OAuthTokenEndpoint",
		},
		{
			name:            "Fails when client certificate is not provided",
			authMode:        oauth.Mtls,
			missingProperty: "APP_%s_SECRET_CERT_PATH",
			propInErrorMsg:  "Certificate",
		},
		{
			name:            "Fails when client certificate key is not provided",
			authMode:        oauth.Mtls,
			missingProperty: "APP_%s_SECRET_CERT_KEY_PATH",
			propInErrorMsg:  "CertificateKey",
		},
	}

	for _, tc := range authCfgTestCase {
		t.Run(tc.name, func(t *testing.T) {
			envs := mandatoryEnvVars(JobName, tenant.Subaccount, tc.authMode)
			envs[tc.missingProperty] = "unknown"
			environ := initEnv(t, envs, JobName)

			// WHEN
			_, err := resync.NewTenantFetcherJobEnvironment(context.TODO(), JobName, resync.ReadFromEnvironment(environ)).ReadJobConfig()

			// THEN
			require.Error(t, err)
			require.Contains(t, err.Error(), fmt.Sprintf(" missing API Client Auth config properties: %s", tc.propInErrorMsg))
		})
	}
}

func initEnv(t *testing.T, envVars map[string]string, jobName string) []string {
	os.Clearenv()
	return setEnvVars(t, envVars, strings.ToUpper(jobName))
}

func initEnvWithMandatory(t *testing.T, envVars map[string]string, jobName string, tenantType tenant.Type, authMode oauth.AuthMode) []string {
	os.Clearenv()
	environ := setEnvVars(t, mandatoryEnvVars(jobName, tenantType, authMode), strings.ToUpper(jobName))
	return append(environ, setEnvVars(t, envVars, strings.ToUpper(jobName))...)
}

func mandatoryEnvVars(jobName string, tenantType tenant.Type, authMode oauth.AuthMode) map[string]string {
	env := map[string]string{
		"APP_%s_JOB_NAME":                   jobName,
		"APP_%s_TENANT_PROVIDER":            "external-provider",
		"APP_%s_TENANT_TYPE":                string(tenantType),
		"APP_%s_API_REGION_NAME":            "eu-1",
		"APP_%s_API_AUTH_CONFIG_SECRET_KEY": "eu-1",
		"APP_%s_API_AUTH_MODE":              string(authMode),
		"APP_%s_SECRET_FILE_PATH":           "testdata/valid.json",
		"APP_%s_SECRET_CLIENT_ID_PATH":      "clientId",
		"APP_%s_SECRET_TOKEN_ENDPOINT_PATH": "tokenUrl",
		"APP_%s_SECRET_TOKEN_PATH":          "/oauth/token",
	}
	switch tenantType {
	case tenant.Account:
		env["APP_%s_API_ENDPOINT_TENANT_CREATED"] = TenantCreatedEndpoint
		env["APP_%s_API_ENDPOINT_TENANT_UPDATED"] = TenantUpdatedEndpoint
		env["APP_%s_API_ENDPOINT_TENANT_DELETED"] = TenantDeletedEndpoint
	case tenant.Subaccount:
		env["APP_%s_API_ENDPOINT_SUBACCOUNT_CREATED"] = TenantCreatedEndpoint
		env["APP_%s_API_ENDPOINT_SUBACCOUNT_UPDATED"] = TenantUpdatedEndpoint
		env["APP_%s_API_ENDPOINT_SUBACCOUNT_MOVED"] = TenantMovedEndpoint
		env["APP_%s_API_ENDPOINT_SUBACCOUNT_DELETED"] = TenantDeletedEndpoint
	}
	switch authMode {
	case oauth.Standard:
		env["APP_%s_SECRET_CLIENT_SECRET_PATH"] = "clientSecret"
	case oauth.Mtls:
		env["APP_%s_SECRET_CERT_PATH"] = "cert"
		env["APP_%s_SECRET_CERT_KEY_PATH"] = "key"
	}
	return env
}

func setEnvVars(t *testing.T, envVars map[string]string, jobName string) []string {
	environ := make([]string, 0, len(envVars))
	for nameFormat, value := range envVars {
		varName := fmt.Sprintf(nameFormat, jobName)
		environ = append(environ, varName+"="+value)
		err := os.Setenv(varName, value)
		require.NoError(t, err)
	}
	return environ
}
