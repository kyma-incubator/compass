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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	gcli "github.com/machinebox/graphql"

	"github.com/google/uuid"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"
)

func TestRegionalOnboardingHandler(t *testing.T) {
	t.Run("Regional account tenant creation", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, tenant, providedTenant.TenantID, providedTenant.Subdomain)
			require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])
		})
	})

	t.Run("Regional subaccount tenant creation", func(t *testing.T) {
		t.Run("Success when parent account tenant is pre-existing", func(t *testing.T) {
			// GIVEN
			parentTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			childTenant := tenantfetcher.Tenant{
				SubaccountID:                uuid.New().String(),
				TenantID:                    parentTenant.TenantID,
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			addRegionalTenantExpectStatusCode(t, parentTenant, http.StatusOK)

			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)
			require.Equal(t, tenantfetcher.RegionPathParamValue, parent.Labels[tenantfetcher.RegionKey])

			// WHEN
			addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
			require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])

			parentTenantAfterInsert, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenantAfterInsert, parentTenant.TenantID, parentTenant.Subdomain)
			require.Equal(t, tenantfetcher.RegionPathParamValue, parentTenantAfterInsert.Labels[tenantfetcher.RegionKey])
		})

		t.Run("Success when parent account tenant does not exist", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				CustomerID:                  uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			// THEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusOK)

			// THEN
			childTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.SubaccountID)
			require.NoError(t, err)
			assertTenant(t, childTenant, providedTenant.SubaccountID, providedTenant.Subdomain)
			require.Equal(t, tenantfetcher.RegionPathParamValue, childTenant.Labels[tenantfetcher.RegionKey])

			parentTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenant, providedTenant.TenantID, "")
			require.Empty(t, parentTenant.Labels)

			customerTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.CustomerID)
			require.NoError(t, err)
			assertTenant(t, customerTenant, providedTenant.CustomerID, "")
			require.Empty(t, customerTenant.Labels)
		})

		t.Run("Should not fail when tenant already exists", func(t *testing.T) {
			// GIVEN
			parentTenantId := uuid.New().String()
			parentTenant := tenantfetcher.Tenant{
				TenantID:                    parentTenantId,
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			childTenant := tenantfetcher.Tenant{
				TenantID:                    parentTenantId,
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			addRegionalTenantExpectStatusCode(t, parentTenant, http.StatusOK)
			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)

			// WHEN
			for i := 0; i < 10; i++ {
				addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)
			}

			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)

			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// THEN
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
			assert.Equal(t, oldTenantState.TotalCount+2, tenants.TotalCount)
		})

		t.Run("Should fail when parent tenantID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				CustomerID:                  uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

		t.Run("Should fail when subdomain is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				CustomerID:                  uuid.New().String(),
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

		t.Run("Should fail when SubscriptionProviderID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				CustomerID:                  uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

		t.Run("Should fail when providerSubaccountID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				CustomerID:                  uuid.New().String(),
				SubscriptionProviderID:      uuid.New().String(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

		t.Run("Should fail when consumerTenantID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				CustomerID:                  uuid.New().String(),
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

		t.Run("Should fail when subscriptionProviderAppName is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:               uuid.New().String(),
				SubaccountID:           uuid.New().String(),
				Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
				CustomerID:             uuid.New().String(),
				SubscriptionProviderID: uuid.New().String(),
				ConsumerTenantID:       uuid.New().String(),
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
		})

	})
}

func TestGetDependenciesHandler(t *testing.T) {
	t.Run("Returns empty body", func(t *testing.T) {
		// GIVEN
		request, err := http.NewRequest(http.MethodGet, config.TenantFetcherFullDependenciesURL, nil)
		require.NoError(t, err)

		tkn := token.GetClientCredentialsToken(t, context.Background(), config.ExternalServicesMockURL+"/secured/oauth/token", config.ClientID, config.ClientSecret, "tenantFetcherClaims")
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

		// WHEN
		response, err := httpClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		responseBody, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		responseBodyJson := make(map[string]interface{}, 0)

		// THEN
		err = json.Unmarshal(responseBody, &responseBodyJson)
		require.NoError(t, err)
		require.Empty(t, responseBodyJson)
	})
}

func addRegionalTenantExpectStatusCode(t *testing.T, providedTenant tenantfetcher.Tenant, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenant, http.MethodPut, config.TenantFetcherFullRegionalURL, expectedStatusCode)
}

func makeTenantRequestExpectStatusCode(t *testing.T, providedTenant tenantfetcher.Tenant, httpMethod, url string, expectedStatusCode int) {
	tenantProperties := tenantfetcher.TenantIDProperties{
		TenantIDProperty:                    config.TenantIDProperty,
		SubaccountTenantIDProperty:          config.SubaccountTenantIDProperty,
		CustomerIDProperty:                  config.CustomerIDProperty,
		SubdomainProperty:                   config.SubdomainProperty,
		SubscriptionProviderIDProperty:      config.SubscriptionProviderIDProperty,
		ProviderSubaccountIdProperty:        config.ProviderSubaccountIDProperty,
		ConsumerTenantIDProperty:            config.ConsumerTenantIDProperty,
		SubscriptionProviderAppNameProperty: config.SubscriptionProviderAppNameProperty,
	}

	request := tenantfetcher.CreateTenantRequest(t, providedTenant, tenantProperties, httpMethod, url, config.ExternalServicesMockURL, config.ClientID, config.ClientSecret)

	t.Log(fmt.Sprintf("Provisioning tenant with ID %s", tenantfetcher.ActualTenantID(providedTenant)))
	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
}

func assertTenant(t *testing.T, tenant *directorSchema.Tenant, tenantID, subdomain string) {
	require.Equal(t, tenantID, tenant.ID)
	if len(subdomain) > 0 {
		require.Equal(t, subdomain, tenant.Labels["subdomain"])
	}
}

const (
	timeout       = time.Minute * 3
	checkInterval = time.Second * 5

	globalAccountCreateSubPath = "global-account-create"
	globalAccountDeleteSubPath = "global-account-delete"
	subaccountMoveSubPath      = "subaccount-move"
	subaccountCreateSubPath    = "subaccount-create"
	subaccountDeleteSubPath    = "subaccount-delete"

	mockEventsPagePattern = `
{
	"totalResults": %d,
	"totalPages": %d,
	"events": [%s]
}`
	mockGlobalAccountEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"customerId": "%s",
		"subdomain": "%s"
	},
	"type": "GlobalAccount"
}`
	mockSubaccountEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"subdomain": "%s",
		"parentGuid": "%s",
		"sourceGlobalAccountGUID": "%s",
		"targetGlobalAccountGUID": "%s",
		"region": "%s"
	},
	"globalAccountGUID": "%s",
	"type": "Subaccount"
}`
)

func TestGlobalAccounts(t *testing.T) {
	ctx := context.TODO()

	externalTenantIDs := []string{"guid1", "guid2"}
	names := []string{"name1", "name2"}
	customerIDs := []string{"customerID1", "customerID2"}
	subdomains := []string{"subdomain1", "subdomain2"}

	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(externalTenantIDs, customerIDs...))

	createEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0])
	createEvent2 := genMockGlobalAccountEvent(externalTenantIDs[1], names[1], customerIDs[1], subdomains[1])
	setMockTenantEvents(t, genMockPage(strings.Join([]string{createEvent1, createEvent2}, ","), 2), globalAccountCreateSubPath)
	defer cleanupMockEvents(t, globalAccountCreateSubPath)

	deleteEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0])
	setMockTenantEvents(t, genMockPage(deleteEvent1, 1), globalAccountDeleteSubPath)
	defer cleanupMockEvents(t, globalAccountDeleteSubPath)

	require.Eventually(t, func() bool {
		tenant1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, externalTenantIDs[0])
		if tenant1 != nil {
			t.Logf("Waiting for tenant %s to be deleted", externalTenantIDs[0])
			return false
		}
		assert.Error(t, err)
		assert.Nil(t, tenant1)

		tenant2, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, externalTenantIDs[1])
		if tenant2 == nil {
			t.Logf("Waiting for tenant %s to be read", externalTenantIDs[1])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, names[1], *tenant2.Name)

		t.Log("TestGlobalAccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestMoveSubaccounts(t *testing.T) {
	ctx := context.TODO()

	gaExternalTenantIDs := []string{"ga1", "ga2"}
	gaNames := []string{"ga1", "ga2"}
	subdomains := []string{"ga1", "ga1"}

	subaccountNames := []string{"sub1", "sub2"}
	subaccountExternalTenants := []string{"sub1", "sub2"}
	subaccountRegion := "test"
	subaccountSubdomain := "sub1"
	subaccountParent := "ga1"
	directoryParentGUID := "test-id" // this is not the global account parent ID but has different semantics

	region := "local"
	provider := "test"

	runtimeNames := []string{"runtime1", "runtime2"}

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           gaNames[0],
			ExternalTenant: gaExternalTenantIDs[0],
			Parent:         nil,
			Subdomain:      &subdomains[0],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
		},
		{
			Name:           gaNames[1],
			ExternalTenant: gaExternalTenantIDs[1],
			Parent:         nil,
			Subdomain:      &subdomains[1],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parent:         &subaccountParent,
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
		},
		{
			Name:           subaccountNames[1],
			ExternalTenant: subaccountExternalTenants[1],
			Parent:         &subaccountParent,
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(gaExternalTenantIDs, subaccountExternalTenants...))

	subaccount1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
	assert.NoError(t, err)
	subaccount2, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[1])
	assert.NoError(t, err)

	runtime1 := registerRuntime(t, ctx, runtimeNames[0], subaccount1.InternalID)
	runtime2 := registerRuntime(t, ctx, runtimeNames[1], subaccount2.InternalID)

	event1 := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion)
	event2 := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion)
	setMockTenantEvents(t, genMockPage(strings.Join([]string{event1, event2}, ","), 2), subaccountMoveSubPath)
	defer cleanupMockEvents(t, subaccountMoveSubPath)

	require.Eventually(t, func() bool {
		tenant1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, gaExternalTenantIDs[0])
		if tenant1 == nil {
			t.Logf("Waiting for global account %s to be read", gaExternalTenantIDs[0])
			return false
		}
		assert.NoError(t, err)

		tenant2, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, gaExternalTenantIDs[1])
		if tenant2 == nil {
			t.Logf("Waiting for global account %s to be read", gaExternalTenantIDs[1])
			return false
		}
		assert.NoError(t, err)

		subaccount1, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
		if subaccount1 == nil || subaccount1.ParentID == tenant1.InternalID {
			t.Logf("Waiting for moved subaccount %s to be read", subaccountExternalTenants[0])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, tenant2.InternalID, subaccount1.ParentID)

		subaccount2, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[1])
		if subaccount2 == nil || subaccount2.ParentID == tenant1.InternalID {
			t.Logf("Waiting for moved subaccount %s to be read", subaccountExternalTenants[1])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, tenant2.InternalID, subaccount2.ParentID)

		rtm1 := fixtures.GetRuntime(t, ctx, directorInternalGQLClient, tenant2.InternalID, runtime1.ID)
		if len(rtm1.Name) == 0 {
			t.Logf("Waiting for runtime %s to be read", runtime1.Name)
			return false
		}
		assert.Equal(t, runtime1.Name, rtm1.Name)

		rtm2 := fixtures.GetRuntime(t, ctx, directorInternalGQLClient, tenant2.InternalID, runtime2.ID)
		if len(rtm2.Name) == 0 {
			t.Logf("Waiting for runtime %s to be read", runtime2.Name)
			return false
		}
		assert.Equal(t, runtime2.Name, rtm2.Name)

		t.Log("TestMoveSubaccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestMoveSubaccountsFailIfSubaccountHasFormationInTheSourceGA(t *testing.T) {
	ctx := context.TODO()

	gaExternalTenantIDs := []string{"ga2"}
	gaNames := []string{"ga2"}
	subdomains := []string{"ga1"}

	subaccountNames := []string{"sub1"}
	subaccountExternalTenants := []string{"sub1"}
	subaccountRegion := "test"
	subaccountSubdomain := "sub1"
	directoryParentGUID := "test-id"

	region := "local"
	provider := "test"

	runtimeNames := []string{"runtime1"}

	defaultTenantID := tenant.TestTenants.GetDefaultTenantID()
	defaultTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, defaultTenantID)
	assert.NoError(t, err)

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           gaNames[0],
			ExternalTenant: gaExternalTenantIDs[0],
			Parent:         nil,
			Subdomain:      &subdomains[0],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parent:         &defaultTenant.InternalID,
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
		},
	}

	err = fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(gaExternalTenantIDs, subaccountExternalTenants...))

	subaccount1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
	assert.NoError(t, err)

	runtime1 := registerRuntime(t, ctx, runtimeNames[0], subaccount1.InternalID)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, defaultTenantID, &runtime1)

	// Add the subaccount to formation
	scenarioName := "testMoveSubaccountScenario"

	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, defaultTenantID, []string{"DEFAULT", scenarioName})
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, defaultTenantID, []string{"DEFAULT"})

	formationInput := graphql.FormationInput{Name: scenarioName}
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountExternalTenants[0], defaultTenantID)
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountExternalTenants[0], defaultTenantID)

	event1 := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, directoryParentGUID, defaultTenantID, defaultTenantID, gaExternalTenantIDs[0], subaccountRegion)
	setMockTenantEvents(t, genMockPage(strings.Join([]string{event1}, ","), 1), subaccountMoveSubPath)
	defer cleanupMockEvents(t, subaccountMoveSubPath)

	require.Eventually(t, func() bool {
		tenant1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, defaultTenantID)
		if tenant1 == nil {
			t.Logf("Waiting for tenant %s to be read", defaultTenantID)
			return false
		}
		assert.NoError(t, err)

		subaccount1, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
		if subaccount1 == nil {
			t.Logf("Waiting for subaccount %s to be read", subaccountExternalTenants[0])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, tenant1.InternalID, subaccount1.ParentID)

		rtm1 := fixtures.GetRuntime(t, ctx, directorInternalGQLClient, tenant1.InternalID, runtime1.ID)
		if len(rtm1.Name) == 0 {
			t.Logf("Waiting for runtime %s to be read", runtime1.Name)
			return false
		}
		assert.Equal(t, runtime1.Name, rtm1.Name)

		t.Log("TestMoveSubaccountsFailIfSubaccountHasFormationInTheSourceGA checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestCreateDeleteSubaccounts(t *testing.T) {
	ctx := context.TODO()

	gaName := "ga1"
	gaExternalTenant := "ga1"
	subdomain1 := "ga1"
	region := "local"

	subaccountNames := []string{"sub1", "sub2"}
	subaccountExternalTenants := []string{"sub1", "sub2"}
	subaccountRegion := "test"
	subaccountParent := "ga1"
	subaccountSubdomain := "sub1"
	directoryParentGUID := "test-id"

	provider := "test"
	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           gaName,
			ExternalTenant: gaExternalTenant,
			Parent:         nil,
			Subdomain:      &subdomain1,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parent:         &subaccountParent,
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)

	// cleanup global account and subaccounts
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(subaccountExternalTenants, gaExternalTenant))

	deleteEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, directoryParentGUID, subaccountParent, "", "", subaccountDeleteSubPath)
	setMockTenantEvents(t, genMockPage(deleteEvent, 1), subaccountDeleteSubPath)
	defer cleanupMockEvents(t, subaccountDeleteSubPath)

	createEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, directoryParentGUID, subaccountParent, "", "", subaccountCreateSubPath)
	setMockTenantEvents(t, genMockPage(createEvent, 1), subaccountCreateSubPath)
	defer cleanupMockEvents(t, subaccountCreateSubPath)

	require.Eventually(t, func() bool {
		subaccount1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
		if subaccount1 != nil {
			t.Logf("Waiting for subaccount %s to be deleted", subaccountExternalTenants[0])
			return false
		}
		assert.Error(t, err)
		assert.Nil(t, subaccount1)

		subaccount2, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[1])
		if subaccount2 == nil {
			t.Logf("Waiting for subaccount %s to be deleted", subaccountExternalTenants[1])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, subaccountNames[1], *subaccount2.Name)

		parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountParent)
		if parent == nil {
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, parent.InternalID, subaccount2.ParentID)

		t.Log("TestCreateDeleteSubaccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestMoveMissingSubaccounts(t *testing.T) {
	ctx := context.TODO()

	gaExternalTenantIDs := []string{"ga1", "ga2"}

	subaccountName := "sub1"
	subaccountExternalTenant := "sub1"
	subaccountRegion := "test"
	subaccountSubdomain := "sub1"
	subaccountParent := "ga1"
	directoryParentGUID := "test-id"

	defer cleanupTenants(t, ctx, directorInternalGQLClient, []string{subaccountExternalTenant, gaExternalTenantIDs[1]})

	event := genMockSubaccountMoveEvent(subaccountExternalTenant, subaccountName, subaccountSubdomain, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion)
	setMockTenantEvents(t, genMockPage(event, 1), subaccountMoveSubPath)
	defer cleanupMockEvents(t, subaccountMoveSubPath)

	require.Eventually(t, func() bool {
		parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, gaExternalTenantIDs[1])
		if parent == nil {
			t.Logf("Waiting for global account %s to be read", gaExternalTenantIDs[1])
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, *parent.Name, gaExternalTenantIDs[1])
		assert.Equal(t, parent.ID, gaExternalTenantIDs[1])

		subaccount, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenant)
		if subaccount == nil {
			t.Logf("Waiting for subaccount %s to be read", subaccountExternalTenant)
			return false
		}
		assert.NoError(t, err)
		assert.Equal(t, subaccount.ID, subaccountExternalTenant)
		assert.Equal(t, subaccount.ParentID, parent.InternalID)

		t.Log("TestMoveMissingSubaccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func genMockGlobalAccountEvent(guid, displayName, customerID, subdomain string) string {
	return fmt.Sprintf(mockGlobalAccountEventPattern, guid, displayName, customerID, subdomain)
}

func genMockSubaccountMoveEvent(guid, displayName, subdomain, directoryParentGUID, parentGuid, sourceGlobalAccountGuid, targetGlobalAccountGuid, region string) string {
	return fmt.Sprintf(mockSubaccountEventPattern, guid, displayName, subdomain, directoryParentGUID, sourceGlobalAccountGuid, targetGlobalAccountGuid, region, parentGuid)
}

func genMockPage(events string, numEvents int) string {
	return fmt.Sprintf(mockEventsPagePattern, numEvents, 1, events)
}

func setMockTenantEvents(t *testing.T, mockEvents string, subPath string) {
	reader := bytes.NewReader([]byte(mockEvents))
	response, err := http.DefaultClient.Post(config.ExternalServicesMockURL+fmt.Sprintf("/tenant-fetcher/%s/configure", subPath), "application/json", reader)
	require.NoError(t, err)
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	if response.StatusCode != http.StatusOK {
		responseBytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to set mock events: %s", string(responseBytes))
	}
}

func cleanupMockEvents(t *testing.T, subPath string) {
	req, err := http.NewRequest(http.MethodDelete, config.ExternalServicesMockURL+fmt.Sprintf("/tenant-fetcher/%s/reset", subPath), nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	if response.StatusCode != http.StatusOK {
		responseBytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to reset mock events: %s", string(responseBytes))
		return
	}
	log.D().Info("Successfully reset mock events")
}

func cleanupTenants(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantExternalIDs []string) {
	var tenantsToDelete []graphql.BusinessTenantMappingInput
	for _, tenantExternalID := range tenantExternalIDs {
		tenantsToDelete = append(tenantsToDelete, graphql.BusinessTenantMappingInput{ExternalTenant: tenantExternalID})
	}
	err := fixtures.DeleteTenants(t, ctx, gqlClient, tenantsToDelete)
	assert.NoError(t, err)
	log.D().Info("Successfully cleanup tenants")
}

func registerRuntime(t *testing.T, ctx context.Context, runtimeName, subaccountInternalID string) graphql.RuntimeExt {
	input := graphql.RuntimeRegisterInput{
		Name: runtimeName,
	}
	return fixtures.RegisterKymaRuntime(t, ctx, directorInternalGQLClient, subaccountInternalID, input, config.GatewayOauth)
}
