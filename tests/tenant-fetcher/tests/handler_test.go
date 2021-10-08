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
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/helper"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	tenant_utils "github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestOnboardingHandler(t *testing.T) {
	t.Run("Success with tenant and customerID", func(t *testing.T) {
		tenantWithCustomer := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			CustomerID:             uuid.New().String(),
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}
		// WHEN
		addTenantExpectStatusCode(t, tenantWithCustomer, http.StatusOK)

		tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenantWithCustomer.TenantID)
		require.NoError(t, err)

		parent, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenantWithCustomer.CustomerID)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tenant, tenantWithCustomer.TenantID, tenantWithCustomer.Subdomain)
		assertTenant(t, parent, tenantWithCustomer.CustomerID, "")
	})

	t.Run("Success with only tenant", func(t *testing.T) {
		providedTenantIDs := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

		tnt, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.TenantID)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tnt, providedTenantIDs.TenantID, providedTenantIDs.Subdomain)
	})

	t.Run("Successful account tenant creation with matching account and subaccount tenant IDs", func(t *testing.T) {
		id := uuid.New().String()
		providedTenantIDs := helper.TenantIDs{
			TenantID:               id,
			SubaccountID:           id,
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

		tnt, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.TenantID)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tnt, providedTenantIDs.TenantID, providedTenantIDs.Subdomain)
	})

	t.Run("Successful account tenant creation with matching customer and account tenant IDs", func(t *testing.T) {
		id := uuid.New().String()
		tenant := helper.TenantIDs{
			CustomerID:             id,
			TenantID:               id,
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}

		addTenantExpectStatusCode(t, tenant, http.StatusOK)

		tnt, err := fixtures.GetTenantByExternalID(dexGraphQLClient, tenant.TenantID)
		require.NoError(t, err)

		// THEN
		assertTenant(t, tnt, tenant.TenantID, tenant.Subdomain)
	})

	t.Run("Should not add already existing tenants", func(t *testing.T) {
		tenantWithCustomer := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			CustomerID:             uuid.New().String(),
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
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
		assertTenantExists(t, tenants, tenantWithCustomer.TenantID)
		assertTenantExists(t, tenants, tenantWithCustomer.CustomerID)
	})

	t.Run("Should fail when no tenantID is provided", func(t *testing.T) {
		providedTenantIDs := helper.TenantIDs{
			CustomerID:             uuid.New().String(),
			SubscriptionProviderID: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})

	t.Run("Should fail when no subdomain is provided", func(t *testing.T) {
		providedTenantIDs := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			CustomerID:             uuid.New().String(),
			SubscriptionProviderID: uuid.New().String(),
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})

	t.Run("Should fail when no SubscriptionProviderID is provided", func(t *testing.T) {
		providedTenantIDs := helper.TenantIDs{
			TenantID:   uuid.New().String(),
			CustomerID: uuid.New().String(),
			Subdomain:  helper.DefaultSubdomain,
		}

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

		tenants, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(tenants))
	})

	t.Run("Should fail with subaccount tenant", func(t *testing.T) {
		// GIVEN
		parentTenant := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}
		childTenant := helper.TenantIDs{
			SubaccountID:           uuid.New().String(),
			TenantID:               parentTenant.TenantID,
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}

		addTenantExpectStatusCode(t, parentTenant, http.StatusOK)

		parent, err := fixtures.GetTenantByExternalID(dexGraphQLClient, parentTenant.TenantID)
		require.NoError(t, err)
		assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)

		// THEN
		addTenantExpectStatusCode(t, childTenant, http.StatusInternalServerError)
	})
}

func TestDecommissioningHandler(t *testing.T) {
	t.Run("Success noop", func(t *testing.T) {
		providedTenantIDs := helper.TenantIDs{
			TenantID:               uuid.New().String(),
			Subdomain:              helper.DefaultSubdomain,
			SubscriptionProviderID: uuid.New().String(),
		}

		addTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

		oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		removeTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

		newTenantState, err := fixtures.GetTenants(dexGraphQLClient)
		require.NoError(t, err)

		// THEN
		assert.Equal(t, len(oldTenantState), len(newTenantState))
	})
}

func TestRegionalOnboardingHandler(t *testing.T) {
	t.Run("Regional account tenant creation", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// GIVEN
			providedTenantIDs := helper.TenantIDs{
				TenantID:               uuid.New().String(),
				Subdomain:              helper.DefaultSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.TenantID)
			require.NoError(t, err)
			assertTenant(t, tenant, providedTenantIDs.TenantID, providedTenantIDs.Subdomain)
			require.Equal(t, helper.RegionPathParamValue, tenant.Labels[helper.RegionKey])
		})
	})

	t.Run("Regional subaccount tenant creation", func(t *testing.T) {
		t.Run("Success when parent account tenant is pre-existing", func(t *testing.T) {
			// GIVEN
			parentTenant := helper.TenantIDs{
				TenantID:               uuid.New().String(),
				Subdomain:              helper.DefaultSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}
			childTenant := helper.TenantIDs{
				SubaccountID:           uuid.New().String(),
				TenantID:               parentTenant.TenantID,
				Subdomain:              helper.DefaultSubaccountSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}

			addRegionalTenantExpectStatusCode(t, parentTenant, http.StatusOK)

			parent, err := fixtures.GetTenantByExternalID(dexGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)
			require.Equal(t, helper.RegionPathParamValue, parent.Labels[helper.RegionKey])

			// WHEN
			addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
			require.Equal(t, helper.RegionPathParamValue, tenant.Labels[helper.RegionKey])

			parentTenantAfterInsert, err := fixtures.GetTenantByExternalID(dexGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenantAfterInsert, parentTenant.TenantID, parentTenant.Subdomain)
			require.Equal(t, helper.RegionPathParamValue, parentTenantAfterInsert.Labels[helper.RegionKey])
		})

		t.Run("Success when parent account tenant does not exist", func(t *testing.T) {
			// GIVEN
			providedTenantIDs := helper.TenantIDs{
				TenantID:               uuid.New().String(),
				CustomerID:             uuid.New().String(),
				SubaccountID:           uuid.New().String(),
				Subdomain:              helper.DefaultSubaccountSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}

			// THEN
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

			// THEN
			childTenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.SubaccountID)
			require.NoError(t, err)
			assertTenant(t, childTenant, providedTenantIDs.SubaccountID, providedTenantIDs.Subdomain)
			require.Equal(t, helper.RegionPathParamValue, childTenant.Labels[helper.RegionKey])

			parentTenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenant, providedTenantIDs.TenantID, "")
			require.Empty(t, parentTenant.Labels)

			customerTenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, providedTenantIDs.CustomerID)
			require.NoError(t, err)
			assertTenant(t, customerTenant, providedTenantIDs.CustomerID, "")
			require.Empty(t, customerTenant.Labels)
		})

		t.Run("Should not fail when tenant already exists", func(t *testing.T) {
			// GIVEN
			parentTenantId := uuid.New().String()
			parentTenant := helper.TenantIDs{
				TenantID:               parentTenantId,
				Subdomain:              helper.DefaultSubaccountSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}
			childTenant := helper.TenantIDs{
				TenantID:               parentTenantId,
				SubaccountID:           uuid.New().String(),
				Subdomain:              helper.DefaultSubaccountSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}
			oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)

			addTenantExpectStatusCode(t, parentTenant, http.StatusOK)
			parent, err := fixtures.GetTenantByExternalID(dexGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain)

			// WHEN
			for i := 0; i < 10; i++ {
				addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)
			}

			tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)

			tenants, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)

			// THEN
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain)
			assert.Equal(t, len(oldTenantState)+2, len(tenants))
		})

		t.Run("Should fail when parent tenantID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenantIDs := helper.TenantIDs{
				CustomerID:             uuid.New().String(),
				SubaccountID:           uuid.New().String(),
				Subdomain:              helper.DefaultSubaccountSubdomain,
				SubscriptionProviderID: uuid.New().String(),
			}
			oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, len(oldTenantState), len(tenants))
		})

		t.Run("Should fail when subdomain is not provided", func(t *testing.T) {
			// GIVEN
			providedTenantIDs := helper.TenantIDs{
				TenantID:               uuid.New().String(),
				SubaccountID:           uuid.New().String(),
				CustomerID:             uuid.New().String(),
				SubscriptionProviderID: uuid.New().String(),
			}
			oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, len(oldTenantState), len(tenants))
		})

		t.Run("Should fail when SubscriptionProviderID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenantIDs := helper.TenantIDs{
				TenantID:     uuid.New().String(),
				SubaccountID: uuid.New().String(),
				CustomerID:   uuid.New().String(),
			}
			oldTenantState, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

			// THEN
			tenants, err := fixtures.GetTenants(dexGraphQLClient)
			require.NoError(t, err)
			assert.Equal(t, len(oldTenantState), len(tenants))
		})
	})

	t.Run("Regional subaccount tenant subscription flows", func(t *testing.T) {
		consumerSubaccountIDsLabelKey := config.ConsumerSubaccountIDsLabelKey
		t.Run("Subscribe tenant for correct consumer", func(t *testing.T) {
			// GIVEN
			ctx := context.Background()
			subscriptionConsumerID := uuid.New().String()
			subaccountID := uuid.New().String()
			accountID := uuid.New().String()

			tenantId := tenant_utils.TestTenants.GetDefaultTenantID()

			subscribedRuntime := registerRuntime(t, ctx, "subscribed-runtime", subscriptionConsumerID)
			defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, tenantId, &subscribedRuntime)

			notSubscribedRuntime := registerRuntime(t, ctx, "not-subscribed-runtime", "fake_subscrioption_id")
			defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, tenantId, &notSubscribedRuntime)

			// WHEN
			providedTenantIDs := helper.TenantIDs{
				TenantID:               accountID,
				SubaccountID:           subaccountID,
				Subdomain:              helper.DefaultSubdomain,
				SubscriptionProviderID: subscriptionConsumerID,
			}

			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

			// THEN
			helper.AssertRuntimeSubscription(t, ctx, subscribedRuntime.ID, providedTenantIDs, dexGraphQLClient, consumerSubaccountIDsLabelKey)

			assertNoRuntimeSubscription(t, ctx, notSubscribedRuntime.ID)
		})

		t.Run("Unsubscribe tenant from correct consumer", func(t *testing.T) {
			// GIVEN
			ctx := context.Background()
			subscriptionConsumerID := uuid.New().String()
			secondSubscriptionConsumerID := uuid.New().String()
			subaccountID := uuid.New().String()
			accountID := uuid.New().String()

			tenantId := tenant_utils.TestTenants.GetDefaultTenantID()

			subscribedRuntime := registerRuntime(t, ctx, "subscribed-runtime", subscriptionConsumerID)
			defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, tenantId, &subscribedRuntime)

			secondSubscribedRuntime := registerRuntime(t, ctx, "second-subscribed-runtime", secondSubscriptionConsumerID)
			defer fixtures.CleanupRuntime(t, ctx, dexGraphQLClient, tenantId, &secondSubscribedRuntime)

			providedTenantIDs := helper.TenantIDs{
				TenantID:               accountID,
				SubaccountID:           subaccountID,
				Subdomain:              helper.DefaultSubdomain,
				SubscriptionProviderID: subscriptionConsumerID,
			}
			addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)
			helper.AssertRuntimeSubscription(t, ctx, subscribedRuntime.ID, providedTenantIDs, dexGraphQLClient, consumerSubaccountIDsLabelKey)

			tenantSecondSubscription := helper.TenantIDs{
				TenantID:               accountID,
				SubaccountID:           subaccountID,
				Subdomain:              helper.DefaultSubdomain,
				SubscriptionProviderID: secondSubscriptionConsumerID,
			}
			addRegionalTenantExpectStatusCode(t, tenantSecondSubscription, http.StatusOK)
			helper.AssertRuntimeSubscription(t, ctx, secondSubscribedRuntime.ID, tenantSecondSubscription, dexGraphQLClient, consumerSubaccountIDsLabelKey)

			// WHEN
			removeRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

			// THEN
			assertNoRuntimeSubscription(t, ctx, subscribedRuntime.ID)
			helper.AssertRuntimeSubscription(t, ctx, secondSubscribedRuntime.ID, tenantSecondSubscription, dexGraphQLClient, consumerSubaccountIDsLabelKey)
		})
	})
}

func TestGetDependenciesHandler(t *testing.T) {
	t.Run("Returns empty body", func(t *testing.T) {
		// GIVEN
		request, err := http.NewRequest(http.MethodGet, config.TenantFetcherFullDependenciesURL, nil)
		require.NoError(t, err)
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
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.FetchTokenFromExternalServicesMock(t, config.ExternalServicesMockURL, claims)))

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

func addTenantExpectStatusCode(t *testing.T, providedTenantIDs helper.TenantIDs, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodPut, config.TenantFetcherFullURL, expectedStatusCode)
}

func addRegionalTenantExpectStatusCode(t *testing.T, providedTenantIDs helper.TenantIDs, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodPut, config.TenantFetcherFullRegionalURL, expectedStatusCode)
}

func removeTenantExpectStatusCode(t *testing.T, providedTenantIDs helper.TenantIDs, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodDelete, config.TenantFetcherFullURL, expectedStatusCode)
}

func removeRegionalTenantExpectStatusCode(t *testing.T, providedTenantIDs helper.TenantIDs, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodDelete, config.TenantFetcherFullRegionalURL, expectedStatusCode)
}

func makeTenantRequestExpectStatusCode(t *testing.T, providedTenantIDs helper.TenantIDs, httpMethod, url string, expectedStatusCode int) {
	tenantProperties := helper.TenantIDProperties{
		TenantIDProperty:               config.TenantIDProperty,
		SubaccountTenantIDProperty:     config.SubaccountTenantIDProperty,
		CustomerIDProperty:             config.CustomerIDProperty,
		SubdomainProperty:              config.SubdomainProperty,
		SubscriptionProviderIDProperty: config.SubscriptionProviderIDProperty,
	}

	request := helper.CreateTenantRequest(t, providedTenantIDs, tenantProperties, httpMethod, url, config.ExternalServicesMockURL)

	t.Log(fmt.Sprintf("Provisioning tenant with ID %s", helper.ActualTenantID(providedTenantIDs)))
	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
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

func registerRuntime(t *testing.T, ctx context.Context, name, subscriptionConsumerID string) graphql.RuntimeExt {
	runtimeInput := graphql.RuntimeInput{
		Name:        name,
		Description: ptr.String(name),
		Labels:      graphql.Labels{helper.RegionKey: helper.RegionPathParamValue, config.SubscriptionProviderLabelKey: subscriptionConsumerID},
	}
	runtimeInGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)
	require.NoError(t, err)
	actualRuntime := graphql.RuntimeExt{}

	registerReq := fixtures.FixRegisterRuntimeRequest(runtimeInGQL)
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerReq, &actualRuntime)
	require.NoError(t, err)
	return actualRuntime
}

func assertNoRuntimeSubscription(t *testing.T, ctx context.Context, runtimeID string) {
	notSubscribedRuntime := graphql.RuntimeExt{}
	getNotSubscribedReq := fixtures.FixGetRuntimeRequest(runtimeID)
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getNotSubscribedReq, &notSubscribedRuntime)
	require.NoError(t, err)

	if _, ok := notSubscribedRuntime.Labels[config.ConsumerSubaccountIDsLabelKey]; !ok {
		return
	}
	require.Len(t, notSubscribedRuntime.Labels[config.ConsumerSubaccountIDsLabelKey], 0)
}
