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

	"github.com/tidwall/gjson"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tenantPathParamValue = "tenant"
	defaultSubdomain     = "default-subdomain"
)

type Tenant struct {
	TenantId   string `json:"tenantId"`
	CustomerId string `json:"customerId"`
	Subdomain  string `json:"subdomain"`
}

func TestOnboardingHandler(t *testing.T) {
	t.Run("Success with tenant and customerID", func(t *testing.T) {
		tenantWithCustomer := Tenant{
			TenantId:   uuid.New().String(),
			CustomerId: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}
		// WHEN
		addTenantExpectStatusCode(t, tenantWithCustomer, http.StatusOK)

		tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenantWithCustomer.TenantId)
		require.NoError(t, err)

		parent, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenantWithCustomer.CustomerId)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tenant, tenantWithCustomer.TenantId, tenantWithCustomer.Subdomain)
		assertTenant(t, parent, tenantWithCustomer.CustomerId, "")
	})

	t.Run("Success with only tenant", func(t *testing.T) {
		tenant := Tenant{
			TenantId: uuid.New().String(),
			Subdomain: defaultSubdomain,
		}

		addTenantExpectStatusCode(t, tenant, http.StatusOK)

		tnt, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenant.TenantId)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tnt, tenant.TenantId, tenant.Subdomain)
	})

	t.Run("Should not add already existing tenants", func(t *testing.T) {
		tenantWithCustomer := Tenant{
			TenantId:   uuid.New().String(),
			CustomerId: uuid.New().String(),
			Subdomain:  defaultSubdomain,
		}
		//GIVEN
		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		//WHEN
		for i := 0; i < 10; i++ {
			addTenantExpectStatusCode(t, tenantWithCustomer, http.StatusOK)
		}

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState)+2, len(tenants))
		assertTenantExists(t, tenants, tenantWithCustomer.TenantId)
		assertTenantExists(t, tenants, tenantWithCustomer.CustomerId)
	})

	t.Run("Should fail when no tenantID is provided", func(t *testing.T) {
		providedTenant := Tenant{
			CustomerId: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})

	t.Run("Should fail when no subdomain is provided", func(t *testing.T) {
		providedTenant := Tenant{
			TenantId:   uuid.New().String(),
			CustomerId: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})
}

func TestDecommissioningHandler(t *testing.T) {
	t.Run("Success noop", func(t *testing.T) {
		providedTenant := Tenant{
			TenantId:  uuid.New().String(),
			Subdomain: defaultSubdomain,
		}

		addTenantExpectStatusCode(t, providedTenant, http.StatusOK)

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		removeTenantExpectStatusCode(t, providedTenant, http.StatusOK)

		newTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(newTenantState))
	})
}

func addTenantExpectStatusCode(t *testing.T, providedTenant Tenant, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenant, http.MethodPut, expectedStatusCode)
}

func removeTenantExpectStatusCode(t *testing.T, providedTenant Tenant, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenant, http.MethodDelete, expectedStatusCode)
}

func makeTenantRequestExpectStatusCode(t *testing.T, providedTenant Tenant, httpMethod string, expectedStatusCode int) {
	byteTenant, err := json.Marshal(providedTenant)
	require.NoError(t, err)

	request, err := http.NewRequest(httpMethod, config.TenantFetcherFullURL, bytes.NewBuffer(byteTenant))
	require.NoError(t, err)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", fetchToken(t)))

	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
}

func fetchToken(t *testing.T) string {
	claims := map[string]interface{}{
		"test": "tenant-fetcher",
		"scope": []string{
			"prefix.Callback",
		},
		"tenant":   "tenant",
		"identity": "tenant-fetcher-tests",
		"iss":      config.ExternalServicesMockURL,
		"exp":      time.Now().Unix() + int64(time.Minute.Seconds()),
	}

	data, err := json.Marshal(claims)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, config.ExternalServicesMockURL+"/oauth/token", bytes.NewBuffer(data))
	require.NoError(t, err)

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	token := gjson.GetBytes(body, "access_token")
	require.True(t, token.Exists())

	return token.String()
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
