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
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/tests/pkg/tenant"

	"github.com/google/uuid"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestRegionalOnboardingHandler(t *testing.T) {
	t.Run("Runtime flows", func(t *testing.T) {
		// GIVEN
		ctx := context.TODO()
		input := fixRuntimeInput("runtime-tf-e2e")
		runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &input)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &runtime)
		require.NoError(t, err)

		t.Run("Regional account tenant creation", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

				// THEN
				tenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenantIDs.TenantID)
				require.NoError(t, err)
				assertTenant(t, tenant, providedTenantIDs.TenantID, providedTenantIDs.Subdomain)
				require.Equal(t, tenantfetcher.RegionPathParamValue, tenant.Labels[tenantfetcher.RegionKey])
			})
		})

		t.Run("Regional subaccount tenant creation", func(t *testing.T) {
			t.Run("Success when parent account tenant is pre-existing", func(t *testing.T) {
				// GIVEN
				parentTenant := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}
				childTenant := tenantfetcher.Tenant{
					SubaccountID:           uuid.New().String(),
					TenantID:               parentTenant.TenantID,
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
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
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					CustomerID:             uuid.New().String(),
					SubaccountID:           uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}

				// THEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

				// THEN
				childTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenantIDs.SubaccountID)
				require.NoError(t, err)
				assertTenant(t, childTenant, providedTenantIDs.SubaccountID, providedTenantIDs.Subdomain)
				require.Equal(t, tenantfetcher.RegionPathParamValue, childTenant.Labels[tenantfetcher.RegionKey])

				parentTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenantIDs.TenantID)
				require.NoError(t, err)
				assertTenant(t, parentTenant, providedTenantIDs.TenantID, "")
				require.Empty(t, parentTenant.Labels)

				customerTenant, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenantIDs.CustomerID)
				require.NoError(t, err)
				assertTenant(t, customerTenant, providedTenantIDs.CustomerID, "")
				require.Empty(t, customerTenant.Labels)
			})

			t.Run("Should not fail when tenant already exists", func(t *testing.T) {
				// GIVEN
				parentTenantId := uuid.New().String()
				parentTenant := tenantfetcher.Tenant{
					TenantID:               parentTenantId,
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}
				childTenant := tenantfetcher.Tenant{
					TenantID:               parentTenantId,
					SubaccountID:           uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
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
				providedTenantIDs := tenantfetcher.Tenant{
					CustomerID:             uuid.New().String(),
					SubaccountID:           uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}
				oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

				// THEN
				tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)
				assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
			})

			t.Run("Should fail when subdomain is not provided", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					SubaccountID:           uuid.New().String(),
					CustomerID:             uuid.New().String(),
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}
				oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

				// THEN
				tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)
				assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
			})

			t.Run("Should fail when SubscriptionProviderID is not provided", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:             uuid.New().String(),
					SubaccountID:         uuid.New().String(),
					CustomerID:           uuid.New().String(),
					ProviderSubaccountID: tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:  "app-name",
				}
				oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

				// THEN
				tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)
				assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
			})

			t.Run("Should fail when providerSubaccountID is not provided", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					SubaccountID:           uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubaccountSubdomain,
					CustomerID:             uuid.New().String(),
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					SubscriptionAppName:    "app-name",
				}
				oldTenantState, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusBadRequest)

				// THEN
				tenants, err := fixtures.GetTenants(certSecuredGraphQLClient)
				require.NoError(t, err)
				assert.Equal(t, oldTenantState.TotalCount, tenants.TotalCount)
			})
		})
	})

	t.Run("Application flows", func(t *testing.T) {
		//GIVEN
		tmplName := "app-flow-tmpl-name"
		ctx := context.TODO()
		appTmplInput := fixAppTemplateInput(tmplName)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTmplInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplate)
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate.ID)

		t.Run("Regional account tenant creation", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}

				// WHEN
				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

				// THEN
				tnt, err := fixtures.GetTenantByExternalID(certSecuredGraphQLClient, providedTenantIDs.TenantID)
				apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID())
				require.Greater(t, apps.TotalCount, 0)
				appExists := false

				var appID string
				for _, v := range apps.Data {
					if str.PtrStrToStr(v.ApplicationTemplateID) == appTemplate.ID {
						appExists = true
						appID = v.ID
					}
				}

				defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appID)

				require.True(t, appExists)
				require.NoError(t, err)
				assertTenant(t, tnt, providedTenantIDs.TenantID, providedTenantIDs.Subdomain)
				require.Equal(t, tenantfetcher.RegionPathParamValue, tnt.Labels[tenantfetcher.RegionKey])
			})
		})

		t.Run("Regional account tenant deletion", func(t *testing.T) {
			t.Run("Success", func(t *testing.T) {
				// GIVEN
				providedTenantIDs := tenantfetcher.Tenant{
					TenantID:               uuid.New().String(),
					Subdomain:              tenantfetcher.DefaultSubdomain,
					SubscriptionProviderID: config.SelfRegDistinguishLabelValue,
					ProviderSubaccountID:   tenant.TestTenants.GetDefaultTenantID(),
					SubscriptionAppName:    "app-name",
				}

				addRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

				apps := fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID())
				require.Greater(t, apps.TotalCount, 0)

				appExists := false
				for _, v := range apps.Data {
					if str.PtrStrToStr(v.ApplicationTemplateID) == appTemplate.ID {
						appExists = true
					}
				}
				require.True(t, appExists)

				// WHEN
				removeRegionalTenantExpectStatusCode(t, providedTenantIDs, http.StatusOK)

				// THEN
				apps = fixtures.GetApplicationPage(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID())
				appExists = false
				for _, v := range apps.Data {
					if str.PtrStrToStr(v.ApplicationTemplateID) == appTemplate.ID {
						appExists = true
					}
				}
				require.False(t, appExists)
			})
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

func addRegionalTenantExpectStatusCode(t *testing.T, providedTenantIDs tenantfetcher.Tenant, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodPut, config.TenantFetcherFullRegionalURL, expectedStatusCode)
}

func removeRegionalTenantExpectStatusCode(t *testing.T, providedTenantIDs tenantfetcher.Tenant, expectedStatusCode int) {
	makeTenantRequestExpectStatusCode(t, providedTenantIDs, http.MethodDelete, config.TenantFetcherFullRegionalURL, expectedStatusCode)
}

func makeTenantRequestExpectStatusCode(t *testing.T, providedTenantIDs tenantfetcher.Tenant, httpMethod, url string, expectedStatusCode int) {
	tenantProperties := tenantfetcher.TenantIDProperties{
		TenantIDProperty:               config.TenantIDProperty,
		SubaccountTenantIDProperty:     config.SubaccountTenantIDProperty,
		CustomerIDProperty:             config.CustomerIDProperty,
		SubdomainProperty:              config.SubdomainProperty,
		SubscriptionProviderIDProperty: config.SubscriptionProviderIDProperty,
		ProviderSubaccountIdProperty:   config.ProviderSubaccountIDProperty,
		SubscriptionAppNameProperty:    config.SubscriptionAppNameProperty,
	}

	request := tenantfetcher.CreateTenantRequest(t, providedTenantIDs, tenantProperties, httpMethod, url, config.ExternalServicesMockURL, config.ClientID, config.ClientSecret)

	t.Log(fmt.Sprintf("Provisioning tenant with ID %s", tenantfetcher.ActualTenantID(providedTenantIDs)))
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

func fixRuntimeInput(name string) directorSchema.RuntimeRegisterInput {
	input := fixtures.FixRuntimeRegisterInput(name)
	input.Labels[config.SelfRegDistinguishLabelKey] = []interface{}{config.SelfRegDistinguishLabelValue}
	input.Labels[tenantfetcher.RegionKey] = config.SelfRegRegion
	delete(input.Labels, "placeholder")

	return input
}

func fixAppTemplateInput(name string) directorSchema.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[config.SelfRegDistinguishLabelKey] = []interface{}{config.SelfRegDistinguishLabelValue}
	input.Labels[tenantfetcher.RegionKey] = config.SelfRegRegion
	input.ApplicationInput.Name = "{{name}}"
	input.ApplicationInput.Description = str.Ptr("{{display-name}}")
	input.Placeholders = []*directorSchema.PlaceholderDefinitionInput{
		{Name: "name", Description: str.Ptr("description")},
		{Name: "display-name", Description: str.Ptr("description")},
	}

	return input
}
