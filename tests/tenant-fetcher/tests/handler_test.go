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
)

type Tenant struct {
	TenantId   string `json:"tenantId"`
	CustomerId string `json:"customerId"`
}

func TestOnboardingHandler(t *testing.T) {
	t.Run("Success with tenant and customerID", func(t *testing.T) {
		tenantWithCustomer := Tenant{
			TenantId:   uuid.New().String(),
			CustomerId: uuid.New().String(),
		}
		// GIVEN
		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// WHEN
		addTenantExpectStatusCode(t, tenantWithCustomer, http.StatusOK)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		require.Equal(t, len(oldTenantState)+2, len(tenants))
		containsTenantWithTenantID(tenantWithCustomer.TenantId, tenants)
		containsTenantWithTenantID(tenantWithCustomer.CustomerId, tenants)
	})

	t.Run("Success with only tenant", func(t *testing.T) {
		tenant := Tenant{
			TenantId: uuid.New().String(),
		}
		// GIVEN
		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// WHEN
		addTenantExpectStatusCode(t, tenant, http.StatusOK)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		require.Equal(t, len(oldTenantState)+1, len(tenants))
		containsTenantWithTenantID(tenant.TenantId, tenants)
	})

	t.Run("Should not add already existing tenants", func(t *testing.T) {
		tenantWithCustomer := Tenant{
			TenantId:   uuid.New().String(),
			CustomerId: uuid.New().String(),
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
		containsTenantWithTenantID(tenantWithCustomer.TenantId, tenants)
		containsTenantWithTenantID(tenantWithCustomer.CustomerId, tenants)
	})

	t.Run("Should fail when no tenantID is provided", func(t *testing.T) {
		providedTenant := Tenant{
			CustomerId: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenant, http.StatusInternalServerError)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})
}

func TestDecommissioningHandler(t *testing.T) {
	t.Run("Success noop", func(t *testing.T) {
		providedTenant := Tenant{
			TenantId: uuid.New().String(),
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

func containsTenantWithTenantID(tenantID string, tenants []*graphql.Tenant) bool {
	for _, t := range tenants {
		if t.ID == tenantID {
			return true
		}
	}
	return false
}
