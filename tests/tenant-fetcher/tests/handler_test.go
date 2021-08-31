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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/authentication"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/sjson"
)

const (
	tenantPathParamValue       = "tenant"
	regionPathParamValue       = "eu-1"
	defaultSubdomain           = "default-subdomain"
	defaultSubaccountSubdomain = "default-subdomain-eu1"
)

type Tenant struct {
	TenantID     string
	SubaccountID string
	CustomerID   string
	Subdomain    string
}

func TestOnboardingHandler(t *testing.T) {
	t.Run("Success with tenant and customerID", func(t *testing.T) {
		// GIVEN
		providedTenant := Tenant{
			TenantID:   uuid.New().String(),
			CustomerID: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}

		// WHEN
		request := createAccountTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenant, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, providedTenant.TenantID)
		require.NoError(t, err)

		parent, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, providedTenant.CustomerID)
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusOK, response.StatusCode)
		assertTenant(t, tenant, providedTenant.TenantID, providedTenant.Subdomain)
		assertTenant(t, parent, providedTenant.CustomerID, "")
	})

	t.Run("Success with only tenant", func(t *testing.T) {
		// GIVEN
		providedTenant := Tenant{
			TenantID:  uuid.New().String(),
			Subdomain: defaultSubdomain,
		}

		// WHEN
		request := createAccountTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenant, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, providedTenant.TenantID)
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusOK, response.StatusCode)
		assertTenant(t, tenant, providedTenant.TenantID, providedTenant.Subdomain)
	})

	t.Run("Should not fail when tenant already exists", func(t *testing.T) {
		//GIVEN
		providedTenant := Tenant{
			TenantID:   uuid.New().String(),
			CustomerID: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}

		//WHEN\
		for i := 0; i < 10; i++ {
			request := createAccountTenantRequest(t, providedTenant)

			httpClient := http.DefaultClient
			httpClient.Timeout = 15 * time.Second

			// THEN
			response, err := httpClient.Do(request)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
		}
	})

	t.Run("Should not add already existing tenants", func(t *testing.T) {
		//GIVEN
		providedTenant := Tenant{
			TenantID:   uuid.New().String(),
			CustomerID: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		//WHEN
		var response *http.Response
		for i := 0; i < 10; i++ {
			request := createAccountTenantRequest(t, providedTenant)

			httpClient := http.DefaultClient
			httpClient.Timeout = 15 * time.Second

			response, err = httpClient.Do(request)
			require.NoError(t, err)
		}

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState)+2, len(tenants))
		require.Equal(t, http.StatusOK, response.StatusCode)
		assertTenantExists(t, tenants, providedTenant.TenantID)
		assertTenantExists(t, tenants, providedTenant.CustomerID)
	})

	t.Run("Should fail when tenantID is not provided", func(t *testing.T) {
		providedTenant := Tenant{
			CustomerID: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		request := createAccountTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), fmt.Sprintf("mandatory property %q is missing from request body", config.TenantIdProperty))
	})

	t.Run("Should fail when subdomain is not provided", func(t *testing.T) {
		providedTenant := Tenant{
			TenantID:   uuid.New().String(),
			CustomerID: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		request := createAccountTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), fmt.Sprintf("mandatory property %q is missing from request body", config.SubdomainProperty))
	})
}

func TestSubaccountCreation(t *testing.T) {
	t.Run("Success with subaccount tenant and parent account tenant", func(t *testing.T) {
		// GIVEN
		parentTenant := Tenant{
			TenantID:  uuid.New().String(),
			Subdomain: defaultSubdomain,
		}
		childTenant := Tenant{
			SubaccountID: uuid.New().String(),
			TenantID:     parentTenant.TenantID,
			Subdomain:    defaultSubaccountSubdomain,
		}

		parentRequest := createAccountTenantRequest(t, parentTenant)
		tenantRequest := createRegionalTenantRequest(t, childTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(parentRequest)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		parent, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, parentTenant.TenantID)
		require.NoError(t, err)
		assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)

		// WHEN
		response, err = httpClient.Do(tenantRequest)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		// THEN
		tenant, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, childTenant.SubaccountID)
		require.NoError(t, err)
		assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
	})

	t.Run("Should not fail when tenant already exists", func(t *testing.T) {
		// GIVEN
		parentTenantId := uuid.New().String()
		parentTenant := Tenant{
			TenantID:  parentTenantId,
			Subdomain: defaultSubaccountSubdomain,
		}
		childTenant := Tenant{
			TenantID:     parentTenantId,
			CustomerID:   uuid.New().String(),
			SubaccountID: uuid.New().String(),
			Subdomain:    defaultSubaccountSubdomain,
		}
		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// WHEN
		parentRequest := createAccountTenantRequest(t, parentTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(parentRequest)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		parent, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, parentTenant.TenantID)
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			tenantRequest := createRegionalTenantRequest(t, childTenant)
			response, err = httpClient.Do(tenantRequest)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
		}
		tenant, err := fixtures.GetTenantByExternalID(config.DirectorUrl, config.Tenant, childTenant.SubaccountID)
		require.NoError(t, err)

		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		// THEN
		// THEN
		assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)
		assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
		assert.Equal(t, len(oldTenantState)+2, len(tenants))
	})

	t.Run("Should fail when parent tenant does not exist", func(t *testing.T) {
		// GIVEN
		providedTenant := Tenant{
			TenantID:     uuid.New().String(),
			CustomerID:   uuid.New().String(),
			SubaccountID: uuid.New().String(),
			Subdomain:    defaultSubaccountSubdomain,
		}

		// WHEN
		request := createRegionalTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		response, err := httpClient.Do(request)
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusInternalServerError, response.StatusCode)
	})

	t.Run("Should fail when parent tenantID is not provided", func(t *testing.T) {
		// GIVEN
		providedTenant := Tenant{
			CustomerID:   uuid.New().String(),
			SubaccountID: uuid.New().String(),
			Subdomain:    defaultSubaccountSubdomain,
		}
		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		request := createRegionalTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		// WHEN
		response, err := httpClient.Do(request)
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)
		assert.Equal(t, len(oldTenantState), len(tenants))
	})

	t.Run("Should fail when subdomain is not provided", func(t *testing.T) {
		// GIVEN
		providedTenant := Tenant{
			TenantID:     uuid.New().String(),
			SubaccountID: uuid.New().String(),
			CustomerID:   uuid.New().String(),
		}
		oldTenantState, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)

		request := createRegionalTenantRequest(t, providedTenant)

		httpClient := http.DefaultClient
		httpClient.Timeout = 15 * time.Second

		// WHEN
		response, err := httpClient.Do(request)
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
		tenants, err := fixtures.GetTenants(config.DirectorUrl, config.Tenant)
		require.NoError(t, err)
		assert.Equal(t, len(oldTenantState), len(tenants))
	})
}

func TestDecommissioningHandler(t *testing.T) {
	t.Run("Success noop", func(t *testing.T) {
		// GIVEN
		providedTenant := map[string]string{
			config.TenantIdProperty:  uuid.New().String(),
			config.SubdomainProperty: defaultSubdomain,
		}

		// WHEN
		endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
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
		assert.Equal(t, len(oldTenantState), len(newTenantState))
		require.Equal(t, http.StatusOK, response.StatusCode)
	})
}

func createAccountTenantRequest(t *testing.T, tenant Tenant) *http.Request {
	endpoint := strings.Replace(config.HandlerEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
	url := config.TenantFetcherURL + config.RootAPI + endpoint

	return createTenantRequest(t, tenant, url)
}

func createRegionalTenantRequest(t *testing.T, tenant Tenant) *http.Request {
	endpoint := strings.Replace(config.HandlerRegionalEndpoint, fmt.Sprintf("{%s}", config.TenantPathParam), tenantPathParamValue, 1)
	endpoint = strings.Replace(config.HandlerRegionalEndpoint, fmt.Sprintf("{%s}", config.RegionPathParam), regionPathParamValue, 1)
	url := config.TenantFetcherURL + config.RootAPI + endpoint

	return createTenantRequest(t, tenant, url)
}

func createTenantRequest(t *testing.T, tenant Tenant, url string) *http.Request {
	var (
		body = "{}"
		err  error
	)

	if len(tenant.TenantID) > 0 {
		body, err = sjson.Set(body, config.TenantIdProperty, tenant.TenantID)
		require.NoError(t, err)
	}
	if len(tenant.SubaccountID) > 0 {
		body, err = sjson.Set(body, config.SubaccountTenantIdProperty, tenant.SubaccountID)
		require.NoError(t, err)
	}
	if len(tenant.CustomerID) > 0 {
		body, err = sjson.Set(body, config.CustomerIdProperty, tenant.CustomerID)
		require.NoError(t, err)
	}
	if len(tenant.Subdomain) > 0 {
		body, err = sjson.Set(body, config.SubdomainProperty, tenant.Subdomain)
		require.NoError(t, err)
	}

	t.Log(fmt.Sprintf("SENDING BODY %s", body))
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer([]byte(body)))
	require.NoError(t, err)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authentication.CreateNotSingedToken(t)))

	return request
}

func assertTenant(t *testing.T, tenant *graphql.Tenant, tenantID, subdomain string) {
	require.Equal(t, tenantID, tenant.ID)
	if len(subdomain) > 0 {
		require.Equal(t, subdomain, tenant.Labels["subdomain"])
	}
}

func assertTenantExists(t *testing.T, tenants []*graphql.Tenant, tenantID string) {
	for _, tenant := range tenants {
		if tenant.ID == tenantID {
			return
		}
	}

	require.Fail(t, fmt.Sprintf("Tenant with ID %q not found in %v", tenantID, tenants))
}
