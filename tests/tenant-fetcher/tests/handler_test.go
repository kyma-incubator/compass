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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/utils/strings/slices"

	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	gcli "github.com/machinebox/graphql"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"
)

var (
	testLicenseType        = "LICENSETYPE"
	customerIDLabelKey     = "customerId"
	costObjectTypeLabelKey = "costObjectType"
	costObjectIdLabelKey   = "costObjectId"
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
				SubscriptionLicenseType:     &testLicenseType,
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, tenant, providedTenant.TenantID, providedTenant.Subdomain, providedTenant.SubscriptionLicenseType)
			require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])
		})

		t.Run("Success with cost object tenant", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				CostObjectID:                uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			// WHEN
			addRegionalTenantExpectStatusCode(t, providedTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, tenant, providedTenant.TenantID, providedTenant.Subdomain, providedTenant.SubscriptionLicenseType)
			require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])

			costObjectTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.CostObjectID)
			require.NoError(t, err)
			require.Equal(t, tenant.Parents, []string{costObjectTenant.InternalID})
		})
	})

	t.Run("Regional subaccount tenant creation", func(t *testing.T) {
		t.Run("Success when parent account tenant is pre-existing", func(t *testing.T) {
			// GIVEN
			parentTenant := tenantfetcher.Tenant{
				TenantID:                    uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			childTenant := tenantfetcher.Tenant{
				SubaccountID:                uuid.New().String(),
				TenantID:                    parentTenant.TenantID,
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}

			addRegionalTenantExpectStatusCode(t, parentTenant, http.StatusOK)

			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain, parentTenant.SubscriptionLicenseType)
			require.Equal(t, tenantfetcher.RegionPathParamValue, parent.Labels[tenantfetcher.RegionKey])

			// WHEN
			addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)

			// THEN
			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain, childTenant.SubscriptionLicenseType)
			require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])

			parentTenantAfterInsert, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenantAfterInsert, parentTenant.TenantID, parentTenant.Subdomain, parentTenant.SubscriptionLicenseType)
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
			assertTenant(t, childTenant, providedTenant.SubaccountID, providedTenant.Subdomain, providedTenant.SubscriptionLicenseType)
			require.Equal(t, tenantfetcher.RegionPathParamValue, childTenant.Labels[tenantfetcher.RegionKey])

			parentTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parentTenant, providedTenant.TenantID, "", providedTenant.SubscriptionLicenseType)
			require.Empty(t, parentTenant.Labels)

			customerTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenant.CustomerID)
			require.NoError(t, err)
			assertTenant(t, customerTenant, providedTenant.CustomerID, "", providedTenant.SubscriptionLicenseType)
			require.Empty(t, customerTenant.Labels)
		})

		t.Run("Should not fail when tenant already exists", func(t *testing.T) {
			// GIVEN
			parentTenantId := uuid.New().String()
			parentTenant := tenantfetcher.Tenant{
				TenantID:                    parentTenantId,
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			childTenant := tenantfetcher.Tenant{
				TenantID:                    parentTenantId,
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
				ProviderSubaccountID:        tenant.TestTenants.GetDefaultTenantID(),
				ConsumerTenantID:            uuid.New().String(),
				SubscriptionProviderAppName: tenantfetcher.SubscriptionProviderAppName,
			}
			oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			addRegionalTenantExpectStatusCode(t, parentTenant, http.StatusOK)
			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, parentTenant.TenantID)
			require.NoError(t, err)
			assertTenant(t, parent, parentTenant.TenantID, parentTenant.Subdomain, parentTenant.SubscriptionLicenseType)

			// WHEN
			for i := 0; i < 10; i++ {
				addRegionalTenantExpectStatusCode(t, childTenant, http.StatusOK)
			}

			tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, childTenant.SubaccountID)
			require.NoError(t, err)

			tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
			require.NoError(t, err)

			// THEN
			assertTenant(t, tenant, childTenant.SubaccountID, childTenant.Subdomain, childTenant.SubscriptionLicenseType)
			assert.Equal(t, oldTenantState.TotalCount+2, tenants.TotalCount)
		})

		t.Run("Should fail when parent tenantID is not provided", func(t *testing.T) {
			// GIVEN
			providedTenant := tenantfetcher.Tenant{
				CustomerID:                  uuid.New().String(),
				SubaccountID:                uuid.New().String(),
				Subdomain:                   tenantfetcher.DefaultSubaccountSubdomain,
				SubscriptionProviderID:      uuid.New().String(),
				SubscriptionLicenseType:     &testLicenseType,
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
				SubscriptionLicenseType:     &testLicenseType,
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
				SubscriptionLicenseType:     &testLicenseType,
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
				SubscriptionLicenseType:     &testLicenseType,
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
				SubscriptionLicenseType:     &testLicenseType,
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
				TenantID:                uuid.New().String(),
				SubaccountID:            uuid.New().String(),
				Subdomain:               tenantfetcher.DefaultSubaccountSubdomain,
				CustomerID:              uuid.New().String(),
				SubscriptionProviderID:  uuid.New().String(),
				SubscriptionLicenseType: &testLicenseType,
				ConsumerTenantID:        uuid.New().String(),
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

	t.Run("Should succeed when a valid region is provided with omit flag", func(t *testing.T) {
		// GIVEN
		// It currently just calls https://compass-gateway.local.kyma.dev/tenants/v1/regional/{region}/dependencies
		regionalEndpoint := strings.Replace(config.DependenciesEndpoint, "{region}", config.SelfRegRegion, 1)
		validRegionUrl := config.TenantFetcherURL + config.RootAPI + regionalEndpoint
		request, err := http.NewRequest(http.MethodGet, validRegionUrl, nil)

		q := request.URL.Query()
		q.Add(config.OmitDependenciesCallbackParam, config.OmitDependenciesCallbackParamValue)
		request.URL.RawQuery = q.Encode()

		require.NoError(t, err)

		tkn := token.GetClientCredentialsToken(t, context.Background(), config.ExternalServicesMockURL+"/secured/oauth/token", config.ClientID,
			config.ClientSecret, claims.TenantFetcherClaimKey)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

		// WHEN
		response, err := httpClient.Do(request)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.Equal(t, "[]", string(body))

		// THEN
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should succeed when a valid region is provided", func(t *testing.T) {
		// GIVEN
		// It currently just calls https://compass-gateway.local.kyma.dev/tenants/v1/regional/{region}/dependencies
		regionalEndpoint := strings.Replace(config.DependenciesEndpoint, "{region}", config.SelfRegRegion, 1)
		validRegionUrl := config.TenantFetcherURL + config.RootAPI + regionalEndpoint
		request, err := http.NewRequest(http.MethodGet, validRegionUrl, nil)
		require.NoError(t, err)

		tkn := token.GetClientCredentialsToken(t, context.Background(), config.ExternalServicesMockURL+"/secured/oauth/token", config.ClientID,
			config.ClientSecret, claims.TenantFetcherClaimKey)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

		// WHEN
		response, err := httpClient.Do(request)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.NotEqual(t, "[]", string(body))

		// THEN
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should succeed when a valid region is provided with invalid omit flag", func(t *testing.T) {
		// GIVEN
		// It currently just calls https://compass-gateway.local.kyma.dev/tenants/v1/regional/{region}/dependencies
		regionalEndpoint := strings.Replace(config.DependenciesEndpoint, "{region}", config.SelfRegRegion, 1)
		validRegionUrl := config.TenantFetcherURL + config.RootAPI + regionalEndpoint
		request, err := http.NewRequest(http.MethodGet, validRegionUrl, nil)
		require.NoError(t, err)

		q := request.URL.Query()
		q.Add(config.OmitDependenciesCallbackParam, "invalid")
		request.URL.RawQuery = q.Encode()

		tkn := token.GetClientCredentialsToken(t, context.Background(), config.ExternalServicesMockURL+"/secured/oauth/token", config.ClientID,
			config.ClientSecret, claims.TenantFetcherClaimKey)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

		// WHEN
		response, err := httpClient.Do(request)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)

		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.NotEqual(t, "[]", string(body))

		// THEN
		require.Equal(t, http.StatusOK, response.StatusCode)
	})

	t.Run("Should fail when invalid region is provided", func(t *testing.T) {

		// GIVEN
		// It currently just calls https://compass-gateway.local.kyma.dev/tenants/v1/regional/{region}/dependencies

		regionalEndpoint := strings.Replace(config.DependenciesEndpoint, "{region}", "invalid", 1)
		invalidRegionUrl := config.TenantFetcherURL + config.RootAPI + regionalEndpoint
		request, err := http.NewRequest(http.MethodGet, invalidRegionUrl, nil)
		require.NoError(t, err)

		tkn := token.GetClientCredentialsToken(t, context.Background(), config.ExternalServicesMockURL+"/secured/oauth/token", config.ClientID,
			config.ClientSecret, claims.TenantFetcherClaimKey)
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tkn))

		// WHEN
		response, err := httpClient.Do(request)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)

		// THEN
		require.Equal(t, http.StatusBadRequest, response.StatusCode)
		body, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		require.Equal(t, "Invalid region provided: invalid\n", string(body))
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
		CostObjectIDProperty:                config.CostObjectIDProperty,
		SubdomainProperty:                   config.SubdomainProperty,
		SubscriptionProviderIDProperty:      config.SubscriptionProviderIDProperty,
		SubscriptionLicenseTypeProperty:     config.SubscriptionLicenseTypeProperty,
		ProviderSubaccountIdProperty:        config.ProviderSubaccountIDProperty,
		ConsumerTenantIDProperty:            config.ConsumerTenantIDProperty,
		SubscriptionProviderAppNameProperty: config.SubscriptionProviderAppNameProperty,
	}

	request := tenantfetcher.CreateTenantRequest(t, providedTenant, tenantProperties, httpMethod, url, config.ExternalServicesMockURL, config.ClientID, config.ClientSecret)

	t.Logf("Provisioning tenant with ID %s", tenantfetcher.ActualTenantID(providedTenant))
	response, err := httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, expectedStatusCode, response.StatusCode)
}

func assertTenant(t *testing.T, tenant *graphql.Tenant, tenantID, subdomain string, licenseType *string) {
	require.Equal(t, tenantID, tenant.ID)
	if len(subdomain) > 0 {
		require.Equal(t, subdomain, tenant.Labels["subdomain"])
	}
	if licenseType != nil {
		require.Equal(t, *licenseType, tenant.Labels["licensetype"])
	}
}

const (
	testProvider = "e2e-test-provider"

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
		"subdomain": "%s",
		"licenseType": "%s"
	},
	"type": "GlobalAccount"
}`
	mockGlobalAccountWithCustomerEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"customerId": "%s",
		"subdomain": "%s",
		"licenseType": "%s"
	},
	"type": "GlobalAccount"
}`
	mockGlobalAccountWithCostObjectEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"costObject": "%s",
		"subdomain": "%s",
		"licenseType": "%s"
	},
	"type": "GlobalAccount"
}`
	mockSubaccountEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"subdomain": "%s",
		"licenseType": "%s",
		"parentGuid": "%s",
		"sourceGlobalAccountGUID": "%s",
		"targetGlobalAccountGUID": "%s",
		"region": "%s",
		"labels": {
			"customerId": ["%s"]
		}
	},
	"globalAccountGUID": "%s",
	"type": "Subaccount"
}`
	mockSubaccountWithCostObjectEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"subdomain": "%s",
		"licenseType": "%s",
		"parentGuid": "%s",
		"region": "%s",
		"costObjectId": "%s",
		"costObjectType": "%s",
		"labels": {
			"customerId": ["%s"]
		}
	},
	"globalAccountGUID": "%s",
	"type": "Subaccount"
}`
)

func TestGlobalAccounts(t *testing.T) {
	ctx := context.TODO()
	externalTenantIDs := []string{"guid1", "guid2"}
	names := []string{"name1", "name2"}
	subdomains := []string{"subdomain1", "subdomain2"}

	t.Run("Having customer parents", func(t *testing.T) {
		customerIDs := []string{"customerID1", "customerID2"}

		defer cleanupTenants(t, ctx, directorInternalGQLClient, append(externalTenantIDs, customerIDs...))

		createEvent1 := genMockGlobalAccountWithCustomerEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0], testLicenseType)
		createEvent2 := genMockGlobalAccountWithCustomerEvent(externalTenantIDs[1], names[1], customerIDs[1], subdomains[1], testLicenseType)
		setMockTenantEvents(t, genMockPage(strings.Join([]string{createEvent1, createEvent2}, ","), 2), globalAccountCreateSubPath)
		defer cleanupMockEvents(t, globalAccountCreateSubPath)

		deleteEvent1 := genMockGlobalAccountWithCustomerEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0], testLicenseType)
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

			t.Log("TestGlobalAccounts/Having_customer_parents checks are successful")
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")
	})

	t.Run("Having cost object parents", func(t *testing.T) {
		costObjectIDs := []string{"costObj1", "costObj2"}

		defer cleanupTenants(t, ctx, directorInternalGQLClient, append(externalTenantIDs, costObjectIDs...))

		createEvent1 := genMockGlobalAccountWithCostObjectEvent(externalTenantIDs[0], names[0], costObjectIDs[0], subdomains[0], testLicenseType)
		createEvent2 := genMockGlobalAccountWithCostObjectEvent(externalTenantIDs[1], names[1], costObjectIDs[1], subdomains[1], testLicenseType)
		setMockTenantEvents(t, genMockPage(strings.Join([]string{createEvent1, createEvent2}, ","), 2), globalAccountCreateSubPath)
		defer cleanupMockEvents(t, globalAccountCreateSubPath)

		deleteEvent1 := genMockGlobalAccountWithCostObjectEvent(externalTenantIDs[0], names[0], costObjectIDs[0], subdomains[0], testLicenseType)
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

			costObject, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, costObjectIDs[1])
			if costObject == nil {
				t.Logf("Waiting for tenant %s to be read", costObjectIDs[1])
				return false
			}
			assert.NoError(t, err)
			assert.Equal(t, tenant2.Parents, []string{costObject.InternalID})

			t.Log("TestGlobalAccounts/Having_cost_object_parents checks are successful")
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")
	})

	t.Run("Having cost object parents brownfield", func(t *testing.T) {
		costObjectIDs := []string{"costObj1"}

		defer cleanupTenants(t, ctx, directorInternalGQLClient, append(externalTenantIDs, costObjectIDs...))

		createEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], names[0], subdomains[0], testLicenseType)
		setMockTenantEvents(t, genMockPage(strings.Join([]string{createEvent1}, ","), 1), globalAccountCreateSubPath)
		defer cleanupMockEvents(t, globalAccountCreateSubPath)

		require.Eventually(t, func() bool {
			account, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, externalTenantIDs[0])
			if account == nil {
				t.Logf("Waiting for tenant %s to be created", externalTenantIDs[0])
				return false
			}
			assert.NoError(t, err)
			assert.Equal(t, names[0], *account.Name)
			assert.Empty(t, account.Parents)

			t.Logf("Account tenant %s is inserted", externalTenantIDs[0])
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")

		createEvent2 := genMockGlobalAccountWithCostObjectEvent(externalTenantIDs[0], names[0], costObjectIDs[0], subdomains[0], testLicenseType)
		setMockTenantEvents(t, genMockPage(strings.Join([]string{createEvent2}, ","), 1), globalAccountCreateSubPath)
		defer cleanupMockEvents(t, globalAccountCreateSubPath)

		require.Eventually(t, func() bool {
			costObject, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, costObjectIDs[0])
			if costObject == nil {
				t.Logf("Waiting for tenant %s to be created", costObjectIDs[0])
				return false
			}
			assert.NoError(t, err)
			assert.NotNil(t, costObject)

			account, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, externalTenantIDs[0])
			assert.NoError(t, err)
			assert.Equal(t, account.Parents, []string{costObject.InternalID})

			t.Logf("Cost Object tenant %s is inserted", costObjectIDs[0])
			t.Logf("TestGlobalAccounts/Having_cost_object_parents_brownfield checks are successful")
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")
	})
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

	runtimeNames := []string{"runtime1", "runtime2"}

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           gaNames[0],
			ExternalTenant: gaExternalTenantIDs[0],
			Parents:        []*string{},
			Subdomain:      &subdomains[0],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           gaNames[1],
			ExternalTenant: gaExternalTenantIDs[1],
			Parents:        []*string{},
			Subdomain:      &subdomains[1],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parents:        []*string{&subaccountParent},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountNames[1],
			ExternalTenant: subaccountExternalTenants[1],
			Parents:        []*string{&subaccountParent},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(gaExternalTenantIDs, subaccountExternalTenants...))

	subaccount1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
	assert.NoError(t, err)
	subaccount2, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[1])
	assert.NoError(t, err)

	var runtime1 graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount1.InternalID, &runtime1)
	runtime1 = registerRuntime(t, ctx, runtimeNames[0], subaccount1.InternalID)

	var runtime2 graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, subaccount2.InternalID, &runtime2)
	runtime2 = registerRuntime(t, ctx, runtimeNames[1], subaccount2.InternalID)

	event1 := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion, "")
	event2 := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion, "")
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
		if subaccount1 == nil || slices.Contains(subaccount1.Parents, tenant1.InternalID) {
			t.Logf("Waiting for moved subaccount %s to be read", subaccountExternalTenants[0])
			return false
		}
		assert.NoError(t, err)
		assert.True(t, slices.Contains(subaccount1.Parents, tenant2.InternalID))

		subaccount2, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[1])
		if subaccount2 == nil || slices.Contains(subaccount2.Parents, tenant1.InternalID) {
			t.Logf("Waiting for moved subaccount %s to be read", subaccountExternalTenants[1])
			return false
		}
		assert.NoError(t, err)
		assert.True(t, slices.Contains(subaccount2.Parents, tenant2.InternalID))

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
			Parents:        []*string{},
			Subdomain:      &subdomains[0],
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountExternalTenants[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parents:        []*string{&defaultTenant.ID},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
			LicenseType:    &testLicenseType,
		},
	}

	err = fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(gaExternalTenantIDs, subaccountExternalTenants...))

	subaccount1, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
	assert.NoError(t, err)

	var runtime1 graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, defaultTenantID, &runtime1)
	input := graphql.RuntimeRegisterInput{
		Name: runtimeNames[0],
	}
	runtime1 = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, subaccount1.ID, input, config.GatewayOauth)

	// Add the subaccount to formation
	scenarioName := "testMoveSubaccountScenario"

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, defaultTenantID, scenarioName)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, defaultTenantID, scenarioName)

	formationInput := graphql.FormationInput{Name: scenarioName}
	defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput.Name, subaccountExternalTenants[0], defaultTenantID)
	fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationInput, subaccountExternalTenants[0], defaultTenantID)

	event1 := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountExternalTenants[0], subaccountSubdomain, testLicenseType, directoryParentGUID, defaultTenantID, defaultTenantID, gaExternalTenantIDs[0], subaccountRegion, "")
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
		assert.True(t, slices.Contains(subaccount1.Parents, tenant1.InternalID))

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
	customerIDs := []string{"0022b8c3-bfda-47b4-8d1b-4c717b9940a3", "000000customerID2"}
	customerIDsTrimmed := []string{"0022b8c3-bfda-47b4-8d1b-4c717b9940a3", "customerID2"}

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
			Parents:        []*string{},
			Subdomain:      &subdomain1,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       provider,
			LicenseType:    &testLicenseType,
		},
		{
			Name:           subaccountNames[0],
			ExternalTenant: subaccountExternalTenants[0],
			Parents:        []*string{&subaccountParent},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       provider,
			LicenseType:    &testLicenseType,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	require.NoError(t, err)

	// cleanup global account and subaccounts
	defer cleanupTenants(t, ctx, directorInternalGQLClient, append(subaccountExternalTenants, gaExternalTenant))

	deleteEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, "", "", subaccountDeleteSubPath, customerIDs[0])
	setMockTenantEvents(t, genMockPage(deleteEvent, 1), subaccountDeleteSubPath)
	defer cleanupMockEvents(t, subaccountDeleteSubPath)
	createEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, "", "", subaccountCreateSubPath, customerIDs[1])
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

		customerIDLabel, exists := subaccount2.Labels[customerIDLabelKey]
		assert.True(t, exists)
		assert.Equal(t, customerIDsTrimmed[1], customerIDLabel)

		parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountParent)
		if parent == nil {
			return false
		}
		assert.NoError(t, err)
		assert.True(t, slices.Contains(subaccount2.Parents, parent.InternalID))

		t.Log("TestCreateDeleteSubaccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestCreateSubaccountsWithCostObject(t *testing.T) {
	ctx := context.TODO()

	gaName := "ga1"
	gaExternalTenant := "ga1"
	subdomain1 := "ga1"
	region := "local"
	customerIDs := []string{"0022b8c3-bfda-47b4-8d1b-4c717b9940a3"}
	customerIDsTrimmed := []string{"0022b8c3-bfda-47b4-8d1b-4c717b9940a3"}

	subaccountNames := []string{"sub1"}
	subaccountExternalTenants := []string{"sub1"}
	subaccountParent := "ga1"
	subaccountSubdomain := "sub1"
	directoryParentGUID := "test-id"
	costObjectId := "a0361a89-f0b0-4f6a-9f32-dd5492477d15"
	costObjectType := "random-type"

	t.Run("Greenfield scenario", func(t *testing.T) {
		provider := "test"
		tenants := []graphql.BusinessTenantMappingInput{
			{
				Name:           gaName,
				ExternalTenant: gaExternalTenant,
				Parents:        []*string{},
				Subdomain:      &subdomain1,
				Region:         &region,
				Type:           string(tenant.Account),
				Provider:       provider,
				LicenseType:    &testLicenseType,
			},
		}
		err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
		require.NoError(t, err)

		// cleanup global account and subaccounts
		defer cleanupTenants(t, ctx, directorInternalGQLClient, append(subaccountExternalTenants, gaExternalTenant, costObjectId))

		createEvent := genMockSubaccountWithCostObjectEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, costObjectId, costObjectType, region, customerIDs[0])
		setMockTenantEvents(t, genMockPage(createEvent, 1), subaccountCreateSubPath)
		defer cleanupMockEvents(t, subaccountCreateSubPath)

		require.Eventually(t, func() bool {
			subaccount, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
			if subaccount == nil {
				t.Logf("Waiting for subaccount %s to be created", subaccountExternalTenants[0])
				return false
			}
			assert.NoError(t, err)
			assert.Equal(t, subaccountNames[0], *subaccount.Name)

			customerIDLabel, exists := subaccount.Labels[customerIDLabelKey]
			assert.True(t, exists)
			assert.Equal(t, customerIDsTrimmed[0], customerIDLabel)

			actualCostObjectId, exists := subaccount.Labels[costObjectIdLabelKey]
			assert.True(t, exists)
			assert.Equal(t, costObjectId, actualCostObjectId)

			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountParent)
			if parent == nil {
				return false
			}
			assert.NoError(t, err)
			assert.True(t, slices.Contains(subaccount.Parents, parent.InternalID))

			costObjectTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, costObjectId)
			if costObjectTenant == nil {
				t.Logf("Waiting for cost object tenant %s to be created", costObjectId)
				return false
			}
			assert.NoError(t, err)

			actualCostObjectType, exists := costObjectTenant.Labels[costObjectTypeLabelKey]
			assert.True(t, exists)
			assert.Equal(t, costObjectType, actualCostObjectType)

			t.Log("TestCreateSubaccountsWithCostObject/Greenfield_scenario checks are successful")
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")
	})

	t.Run("Brownfield scenario", func(t *testing.T) {
		provider := "test"
		tenants := []graphql.BusinessTenantMappingInput{
			{
				Name:           gaName,
				ExternalTenant: gaExternalTenant,
				Parents:        []*string{},
				Subdomain:      &subdomain1,
				Region:         &region,
				Type:           string(tenant.Account),
				Provider:       provider,
				LicenseType:    &testLicenseType,
			},
		}
		err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
		require.NoError(t, err)

		defer cleanupTenants(t, ctx, directorInternalGQLClient, append(subaccountExternalTenants, gaExternalTenant, costObjectId))

		createEvent1 := genMockSubaccountWithCostObjectEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, "", "", region, customerIDs[0])
		setMockTenantEvents(t, genMockPage(createEvent1, 1), subaccountCreateSubPath)

		require.Eventually(t, func() bool {
			subaccount, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
			if subaccount == nil {
				t.Logf("Waiting for subaccount %s to be created", subaccountExternalTenants[0])
				return false
			}
			assert.NoError(t, err)
			assert.Equal(t, subaccountNames[0], *subaccount.Name)

			customerIDLabel, exists := subaccount.Labels[customerIDLabelKey]
			assert.True(t, exists)
			assert.Equal(t, customerIDsTrimmed[0], customerIDLabel)

			_, exists = subaccount.Labels[costObjectIdLabelKey]
			assert.False(t, exists)

			parent, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountParent)
			if parent == nil {
				return false
			}
			assert.NoError(t, err)
			assert.True(t, slices.Contains(subaccount.Parents, parent.InternalID))

			_, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, costObjectId)
			assert.Error(t, err)

			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")

		cleanupMockEvents(t, subaccountCreateSubPath)
		createEvent2 := genMockSubaccountWithCostObjectEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, costObjectId, costObjectType, region, customerIDs[0])
		setMockTenantEvents(t, genMockPage(createEvent2, 1), subaccountCreateSubPath)
		defer cleanupMockEvents(t, subaccountCreateSubPath)

		require.Eventually(t, func() bool {
			costObjectTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, costObjectId)
			if costObjectTenant == nil {
				t.Logf("Waiting for cost object tenant %s to be created", costObjectId)
				return false
			}
			assert.NoError(t, err)

			actualCostObjectType, exists := costObjectTenant.Labels[costObjectTypeLabelKey]
			assert.True(t, exists)
			assert.Equal(t, costObjectType, actualCostObjectType)

			subaccount, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenants[0])
			if subaccount == nil {
				t.Logf("Waiting for subaccount %s to be created", subaccountExternalTenants[0])
				return false
			}
			assert.NoError(t, err)
			assert.Equal(t, subaccountNames[0], *subaccount.Name)

			actualCostObjectId, exists := subaccount.Labels[costObjectIdLabelKey]
			assert.True(t, exists)
			assert.Equal(t, costObjectId, actualCostObjectId)

			t.Log("TestCreateSubaccountsWithCostObject/Brownfield_scenario checks are successful")
			return true
		}, timeout, checkInterval, "Waiting for tenants retrieval.")
	})
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

	tenants := []graphql.BusinessTenantMappingInput{
		{
			Name:           gaExternalTenantIDs[1],
			ExternalTenant: gaExternalTenantIDs[1],
			Parents:        []*string{},
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Account),
			Provider:       testProvider,
			LicenseType:    &testLicenseType,
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, tenants)
	assert.NoError(t, err)

	defer cleanupTenants(t, ctx, directorInternalGQLClient, []string{subaccountExternalTenant, gaExternalTenantIDs[1]})

	event := genMockSubaccountMoveEvent(subaccountExternalTenant, subaccountName, subaccountSubdomain, testLicenseType, directoryParentGUID, subaccountParent, gaExternalTenantIDs[0], gaExternalTenantIDs[1], subaccountRegion, "")
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
		assert.True(t, slices.Contains(subaccount.Parents, parent.InternalID))

		t.Log("TestMoveMissingSubaccounts checks are successful")
		return true
	}, timeout, checkInterval, "Waiting for tenants retrieval.")
}

func TestGetSubaccountOnDemandIfMissing(t *testing.T) {
	ctx := context.TODO()
	subaccountExternalTenant := config.OnDemandTenant

	tenantsToDelete := []graphql.BusinessTenantMappingInput{
		{
			ExternalTenant: subaccountExternalTenant,
		},
	}

	defer func() {
		err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenantsToDelete)
		assert.NoError(t, err)
	}()

	t.Logf("Deleting tenant %q", subaccountExternalTenant)
	err := fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, tenantsToDelete)
	require.NoError(t, err)

	t.Logf("Retrieving tenant %q by external id", subaccountExternalTenant)
	subaccount, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, subaccountExternalTenant)
	region, regionExists := subaccount.Labels[tenantfetcher.RegionKey]

	require.NoError(t, err)
	require.NotNil(t, subaccount)
	require.Equal(t, subaccount.ID, subaccountExternalTenant)
	require.True(t, regionExists)
	require.Equal(t, region, config.TenantRegionPrefix+config.TenantRegion)

	t.Log("TestGetSubaccountOnDemandIfMissing checks are successful")
}

func genMockGlobalAccountWithCustomerEvent(guid, displayName, customerID, subdomain, licenseType string) string {
	return fmt.Sprintf(mockGlobalAccountWithCustomerEventPattern, guid, displayName, customerID, subdomain, licenseType)
}

func genMockGlobalAccountWithCostObjectEvent(guid, displayName, costObjectID, subdomain, licenseType string) string {
	return fmt.Sprintf(mockGlobalAccountWithCostObjectEventPattern, guid, displayName, costObjectID, subdomain, licenseType)
}

func genMockGlobalAccountEvent(guid, displayName, subdomain, licenseType string) string {
	return fmt.Sprintf(mockGlobalAccountEventPattern, guid, displayName, subdomain, licenseType)
}

func genMockSubaccountMoveEvent(guid, displayName, subdomain, licenseType, directoryParentGUID, parentGuid, sourceGlobalAccountGuid, targetGlobalAccountGuid, region, customerID string) string {
	return fmt.Sprintf(mockSubaccountEventPattern, guid, displayName, subdomain, licenseType, directoryParentGUID, sourceGlobalAccountGuid, targetGlobalAccountGuid, region, customerID, parentGuid)
}

func genMockSubaccountWithCostObjectEvent(guid, displayName, subdomain, licenseType, directoryParentGUID, parentGuid, costObjectId, costObjectType, region, customerID string) string {
	return fmt.Sprintf(mockSubaccountWithCostObjectEventPattern, guid, displayName, subdomain, licenseType, directoryParentGUID, region, costObjectId, costObjectType, customerID, parentGuid)
}

func genMockPage(events string, numEvents int) string {
	return fmt.Sprintf(mockEventsPagePattern, numEvents, 1, events)
}

func setMockTenantEvents(t *testing.T, mockEvents string, subPath string) {
	reader := bytes.NewReader([]byte(mockEvents))
	response, err := http.DefaultClient.Post(config.ExternalServicesMockURL+fmt.Sprintf("/tenant-fetcher/%s/configure", subPath), "application/json", reader)
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
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
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
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

func registerRuntime(t require.TestingT, ctx context.Context, runtimeName, subaccountInternalID string) graphql.RuntimeExt {
	input := &graphql.RuntimeRegisterInput{
		Name: runtimeName,
	}
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorInternalGQLClient, subaccountInternalID, input)
	assert.NoError(t, err)
	return runtime
}
