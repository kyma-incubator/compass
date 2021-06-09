/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/tests/pkg/authentication"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

const (
	tenantFetcherURL          = "TENANT_FETCHER_URL"
	rootAPI                   = "ROOT_API"
	handlerEndpoint           = "HANDLER_ENDPOINT"
	tenantPathParam           = "TENANT_PATH_PARAM"
	dbUser                    = "APP_DB_USER"
	dbPassword                = "APP_DB_PASSWORD"
	dbHost                    = "APP_DB_HOST"
	dbPort                    = "APP_DB_PORT"
	dbName                    = "APP_DB_NAME"
	dbSSL                     = "APP_DB_SSL"
	dbMaxOpenConnections      = "APP_DB_MAX_OPEN_CONNECTIONS"
	dbMaxIdleConnections      = "APP_DB_MAX_IDLE_CONNECTIONS"
	identityZone              = "APP_TENANT_IDENTITY_ZONE"
	defaultTenant             = "APP_TENANT"
	directorURL               = "APP_DIRECTOR_URL"
	subscriptionCallbackScope = "APP_SUBSCRIPTION_CALLBACK_SCOPE"
	tenantProvider            = "APP_TENANT_PROVIDER"
)

type config struct {
	TenantFetcherURL          string
	RootAPI                   string
	HandlerEndpoint           string
	TenantPathParam           string
	DbUser                    string
	DbPassword                string
	DbHost                    string
	DbPort                    string
	DbName                    string
	DbSSL                     string
	DbMaxIdleConnections      string
	DbMaxOpenConnections      string
	IdentityZone              string
	Tenant                    string
	DirectorUrl               string
	SubscriptionCallbackScope string
	TenantProvider            string
}

type Tenant struct {
	TenantId   string `json:"tenantId"`
	CustomerId string `json:"customerId"`
	Subdomain  string `json:"subdomain"`
}

func TestOnboardingHandler(t *testing.T) {
	config := loadConfig(t)

	t.Run("Success", func(t *testing.T) {
		// GIVEN

		providedTenant := &Tenant{
			TenantId:   "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72",
			CustomerId: "160269",
			Subdomain:  "subdomain",
		}

		cleanUp(t, providedTenant, config)

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// WHEN
		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Greater(t, len(tenants), len(oldTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Success with only tenant id provided", func(t *testing.T) {
		// GIVEN

		providedTenant := &Tenant{
			TenantId: "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72",
		}

		cleanUp(t, providedTenant, config)

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// WHEN
		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Greater(t, len(tenants), len(oldTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Success with only customer id provided", func(t *testing.T) {
		// GIVEN

		providedTenant := &Tenant{
			CustomerId: "160269",
		}

		cleanUp(t, providedTenant, config)

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// WHEN
		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Greater(t, len(tenants), len(oldTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should not fail when tenant already exists", func(t *testing.T) {
		providedTenant := &Tenant{
			TenantId:   config.Tenant,
			CustomerId: "160269",
			Subdomain:  "subdomain",
		}

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(tenants), len(oldTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should fail when both tenant id and customer id are missing", func(t *testing.T) {
		providedTenant := &Tenant{
			Subdomain: "subdomain",
		}

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(tenants), len(oldTenantState))
		require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	})
}

func TestDecommissioningHandler(t *testing.T) {
	config := loadConfig(t)

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		providedTenant := &Tenant{
			TenantId:   "cb0bb8f2-7b44-4dd2-bce1-fa0c19169b79",
			CustomerId: "160269",
			Subdomain:  "subdomain",
		}
		cleanUp(t, providedTenant, config)
		// WHEN
		tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantID, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		// Add test tenant
		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		// Initial state
		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		request, err = http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		response, err = httpClient.Do(request)
		require.NoError(t, err)

		newTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Greater(t, len(oldTenantState), len(newTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should not fail when tenant does not exists", func(t *testing.T) {
		providedTenant := &Tenant{
			TenantId:   "cb0bb8f2-7b44-4dd2-bce1-fa0c19169b79",
			CustomerId: "160269",
			Subdomain:  "subdomain",
		}
		cleanUp(t, providedTenant, config)

		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), providedTenant.TenantId, 1)
		url := config.TenantFetcherURL + config.RootAPI + endpoint

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		byteTenant, err := json.Marshal(providedTenant)
		require.NoError(t, err)
		request, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(byteTenant))
		require.NoError(t, err)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		newTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(newTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})
}

func loadConfig(t *testing.T) config {
	config := config{
		TenantFetcherURL:          os.Getenv(tenantFetcherURL),
		RootAPI:                   os.Getenv(rootAPI),
		HandlerEndpoint:           os.Getenv(handlerEndpoint),
		TenantPathParam:           os.Getenv(tenantPathParam),
		DbUser:                    os.Getenv(dbUser),
		DbPassword:                os.Getenv(dbPassword),
		DbHost:                    os.Getenv(dbHost),
		DbPort:                    os.Getenv(dbPort),
		DbName:                    os.Getenv(dbName),
		DbSSL:                     os.Getenv(dbSSL),
		DbMaxIdleConnections:      os.Getenv(dbMaxIdleConnections),
		DbMaxOpenConnections:      os.Getenv(dbMaxOpenConnections),
		IdentityZone:              os.Getenv(identityZone),
		Tenant:                    os.Getenv(defaultTenant),
		DirectorUrl:               os.Getenv(directorURL),
		SubscriptionCallbackScope: os.Getenv(subscriptionCallbackScope),
		TenantProvider:            os.Getenv(tenantProvider),
	}

	require.NotEmpty(t, config.TenantFetcherURL)
	require.NotEmpty(t, config.RootAPI)
	require.NotEmpty(t, config.HandlerEndpoint)
	require.NotEmpty(t, config.TenantPathParam)
	require.NotEmpty(t, config.DbUser)
	require.NotEmpty(t, config.DbPassword)
	require.NotEmpty(t, config.DbHost)
	require.NotEmpty(t, config.DbPort)
	require.NotEmpty(t, config.DbName)
	require.NotEmpty(t, config.DbSSL)
	require.NotEmpty(t, config.DbMaxIdleConnections)
	require.NotEmpty(t, config.DbMaxOpenConnections)
	require.NotEmpty(t, config.IdentityZone)
	require.NotEmpty(t, config.Tenant)
	require.NotEmpty(t, config.DirectorUrl)
	require.NotEmpty(t, config.SubscriptionCallbackScope)
	require.NotEmpty(t, config.TenantProvider)

	return config
}

func cleanUp(t *testing.T, tenant *Tenant, config config) {
	tenantID := "ad0bb8f2-7b44-4dd2-bce1-fa0c19169b72"
	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantID, 1)
	url := config.TenantFetcherURL + config.RootAPI + endpoint

	byteTenant, err := json.Marshal(tenant)
	require.NoError(t, err)

	request, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(byteTenant))
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))
	require.NoError(t, err)

	httpClient := http.DefaultClient
	httpClient.Timeout = 15 * time.Second

	_, err = httpClient.Do(request)
	require.NoError(t, err)
}
