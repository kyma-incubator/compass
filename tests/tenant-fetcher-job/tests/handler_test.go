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

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	globalAccountCreateSubPath = "global-account-create"
	globalAccountDeleteSubPath = "global-account-delete"
	subaccountMoveSubPath      = "subaccount-move"
	subaccountCreateSubPath    = "subaccount-create"
	subaccountDeleteSubPath    = "subaccount-delete"

	namespace                 = "compass-system"
	globalAccountsJobName     = "tenant-fetcher-account-test"
	globalAccountsCronJobName = "compass-tenant-fetcher-account-fetcher"
	subaccountsJobName        = "tenant-fetcher-subaccount-test"
	subaccountsCronJobName    = "compass-tenant-fetcher-subaccount-fetcher"

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
		"global_subaccount_id": "%s",
		"sourceGlobalAccountGUID": "%s",
		"targetGlobalAccountGUID": "%s",
		"region": "%s"
	},
	"type": "Subaccount"
}`
)

func TestGlobalAccounts(t *testing.T) {
	ctx := context.TODO()

	externalTenantIDs := []string{"guid1", "guid2"}
	names := []string{"name1", "name2"}
	customerIDs := []string{"customerID1", "customerID2"}
	subdomains := []string{"subdomain1", "subdomain2"}

	createEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0])
	createEvent2 := genMockGlobalAccountEvent(externalTenantIDs[1], names[1], customerIDs[1], subdomains[1])
	setMockTenantEvents(t, []byte(genMockPage(strings.Join([]string{createEvent1, createEvent2}, ","), 2)), globalAccountCreateSubPath)
	defer cleanupMockEvents(t, globalAccountCreateSubPath)

	deleteEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], names[0], customerIDs[0], subdomains[0])
	setMockTenantEvents(t, []byte(genMockPage(deleteEvent1, 1)), globalAccountDeleteSubPath)
	defer cleanupMockEvents(t, globalAccountDeleteSubPath)

	defer cleanupTenants(t, ctx, directorInternalGQLClient, externalTenantIDs)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, globalAccountsCronJobName, globalAccountsJobName, namespace)
	defer func() {
		k8s.DeleteJob(t, ctx, k8sClient, globalAccountsJobName, namespace)
	}()

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, globalAccountsJobName, namespace)

	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, externalTenantIDs[0])
	assert.Error(t, err)
	assert.Nil(t, tenant1)

	tenant2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, externalTenantIDs[1])
	assert.NoError(t, err)
	assert.Equal(t, names[1], *tenant2.Name)
}

func TestMoveSubaccounts(t *testing.T) {
	ctx := context.TODO()

	gaExternalTenantIDs := []string{"ga1", "ga1"}
	gaNames := []string{"ga1", "ga2"}
	subdomains := []string{"ga1", "ga1"}

	subaccountNames := []string{"sub1", "sub2"}
	subaccountExternalTenants := []string{"sub1", "sub2"}
	subaccountRegion := "test"
	subaccountSubdomain := "sub1"
	subaccountParent := "ga1"

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

	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, gaExternalTenantIDs[0])
	assert.NoError(t, err)
	tenant2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, gaExternalTenantIDs[1])
	assert.NoError(t, err)

	runtime1 := registerRuntime(t, ctx, runtimeNames[0], subaccountExternalTenants[0], tenant1.InternalID)
	runtime2 := registerRuntime(t, ctx, runtimeNames[1], subaccountExternalTenants[1], tenant1.InternalID)

	event1 := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, subaccountParent, cfg.MovedRuntimeLabelKey, tenant1.ID, tenant2.ID, subaccountRegion)
	event2 := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, subaccountParent, cfg.MovedRuntimeLabelKey, tenant1.ID, tenant2.ID, subaccountRegion)
	setMockTenantEvents(t, []byte(genMockPage(strings.Join([]string{event1, event2}, ","), 2)), subaccountMoveSubPath)
	defer cleanupMockEvents(t, subaccountMoveSubPath)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	assert.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, subaccountsCronJobName, subaccountsJobName, namespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, subaccountsJobName, namespace)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, subaccountsJobName, namespace)

	subaccount1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, subaccountExternalTenants[0])
	assert.NoError(t, err)
	assert.Equal(t, tenant2.InternalID, subaccount1.ParentID)

	subaccount2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, subaccountExternalTenants[1])
	assert.NoError(t, err)
	assert.Equal(t, tenant2.InternalID, subaccount2.ParentID)

	rtm1 := fixtures.GetRuntime(t, ctx, directorInternalGQLClient, tenant2.InternalID, runtime1.ID)
	assert.Equal(t, runtime1.Name, rtm1.Name)

	rtm2 := fixtures.GetRuntime(t, ctx, directorInternalGQLClient, tenant2.InternalID, runtime2.ID)
	assert.Equal(t, runtime2.Name, rtm2.Name)
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

	deleteEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[0], subaccountNames[0], subaccountSubdomain, subaccountParent, "", "", "", subaccountDeleteSubPath)
	setMockTenantEvents(t, []byte(genMockPage(deleteEvent, 1)), subaccountDeleteSubPath)
	defer cleanupMockEvents(t, subaccountDeleteSubPath)

	createEvent := genMockSubaccountMoveEvent(subaccountExternalTenants[1], subaccountNames[1], subaccountSubdomain, subaccountParent, "", "", "", subaccountCreateSubPath)
	setMockTenantEvents(t, []byte(genMockPage(createEvent, 1)), subaccountCreateSubPath)
	defer cleanupMockEvents(t, subaccountCreateSubPath)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	assert.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, subaccountsCronJobName, subaccountsJobName, namespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, subaccountsJobName, namespace)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, subaccountsJobName, namespace)

	subaccount1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, subaccountExternalTenants[0])
	assert.Error(t, err)
	assert.Nil(t, subaccount1)

	subaccount2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, subaccountExternalTenants[1])
	assert.NoError(t, err)
	assert.Equal(t, subaccountNames[1], *subaccount2.Name)
}

func genMockGlobalAccountEvent(guid, displayName, customerID, subdomain string) string {
	return fmt.Sprintf(mockGlobalAccountEventPattern, guid, displayName, customerID, subdomain)
}

func genMockSubaccountMoveEvent(guid, displayName, subdomain, parentGuid, movedLabelKey, sourceGlobalAccountGuid, targetGlobalAccountGuid, region string) string {
	return fmt.Sprintf(mockSubaccountEventPattern, guid, displayName, subdomain, parentGuid, movedLabelKey, sourceGlobalAccountGuid, targetGlobalAccountGuid, region)
}

func genMockPage(events string, numEvents int) string {
	return fmt.Sprintf(mockEventsPagePattern, numEvents, 1, events)
}

func setMockTenantEvents(t *testing.T, mockEvents []byte, subPath string) {
	reader := bytes.NewReader(mockEvents)
	response, err := http.DefaultClient.Post(cfg.ExternalSvcMockURL+fmt.Sprintf("/tenant-fetcher/%s/configure", subPath), "application/json", reader)
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
	req, err := http.NewRequest(http.MethodDelete, cfg.ExternalSvcMockURL+fmt.Sprintf("/tenant-fetcher/%s/reset", subPath), nil)
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

func registerRuntime(t require.TestingT, ctx context.Context, runtimeName, externalTenantID, globalAccountInternalID string) graphql.RuntimeExt {
	input := &graphql.RuntimeInput{
		Name:   runtimeName,
		Labels: map[string]interface{}{cfg.MovedRuntimeLabelKey: externalTenantID},
	}
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorInternalGQLClient, globalAccountInternalID, input)
	assert.NoError(t, err)
	return runtime
}
