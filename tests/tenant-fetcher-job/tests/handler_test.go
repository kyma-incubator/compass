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

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockEventsPagePattern = `
{
	"totalResults": %d,
	"totalPages": %d,
	"events": [%s]
}`
	mockEventPattern = `
{
	"eventData": {
		"guid": "%s",
		"displayName": "%s",
		"customerId": "%s",
		"subdomain": "%s"
	},
	"type": "GlobalAccount"
}`
	mockSubaccountMoveEventPattern = `
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

const (
	globalAccountCreateSubPath = "global-account-create"
	globalAccountDeleteSubPath = "global-account-delete"
	globalAccountUpdateSubPath = "global-account-update"
	subaccountMoveSubPath      = "subaccount-move"
)

func TestGlobalAccounts(t *testing.T) {
	ctx := context.TODO()

	externalTenantIDs := []string{"guid1", "guid2", "guid3"}
	createEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], "name1", "customerID1", "subdomain1")
	createEvent2 := genMockGlobalAccountEvent(externalTenantIDs[1], "name2", "customerID2", "subdomain2")
	setMockTenantEvents(t, []byte(genMockPage(strings.Join([]string{createEvent1, createEvent2}, ","), 3)), globalAccountCreateSubPath)
	defer cleanupMockEvents(t, globalAccountCreateSubPath)

	deleteEvent1 := genMockGlobalAccountEvent(externalTenantIDs[0], "name1", "customerID1", "subdomain1")
	setMockTenantEvents(t, []byte(genMockPage(deleteEvent1, 1)), globalAccountDeleteSubPath)
	defer cleanupMockEvents(t, globalAccountDeleteSubPath)

	defer func(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenants []graphql.BusinessTenantMappingInput) {
		err := fixtures.DeleteTenants(t, ctx, gqlClient, tenants)
		assert.NoError(t, err)
	}(t, ctx, directorInternalGQLClient, []graphql.BusinessTenantMappingInput{{ExternalTenant: externalTenantIDs[0]}, {ExternalTenant: externalTenantIDs[1]}})

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "tenant-fetcher-test"
	namespace := "compass-system"

	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-external-mock", jobName, namespace)
	defer func() {
		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "guid1")
	assert.Error(t, err)
	assert.Nil(t, tenant1)

	tenant2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "guid2")
	assert.NoError(t, err)
	assert.Equal(t, *tenant2.Name, "name2")
}

//func TestUpdateGlobalAccounts(t *testing.T) {
//	ctx := context.TODO()
//
//	event1 := genMockGlobalAccountEvent("guid1", "name1", "customerID1", "subdomain1")
//	event2 := genMockGlobalAccountEvent("guid2", "name2", "customerID2", "subdomain2")
//	event3 := genMockGlobalAccountEvent("guid2", "name2", "customerID2", "subdomain2")
//	setMockTenantEvents(t, []byte(genMockPage(strings.Join([]string{event1, event2}, ","), 3)), globalAccountCreateSubPath)
//	defer cleanupMockEvents(t, globalAccountCreateSubPath)
//
//	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
//	require.NoError(t, err)
//	jobName := "tenant-fetcher-test"
//	namespace := "compass-system"
//
//	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-external-mock", jobName, namespace)
//	defer func() {
//		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
//	}()
//
//	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)
//
//	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "guid1")
//	assert.NoError(t, err)
//	assert.Equal(t, *tenant1.Name, "name1")
//	tenant2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "guid2")
//	assert.NoError(t, err)
//	assert.Equal(t, *tenant2.Name, "name2")
//
//	setMockTenantEvents(t, []byte(genMockPage(event1, 1)), globalAccountUpdateSubPath)
//	defer cleanupMockEvents(t, globalAccountUpdateSubPath)
//	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-external-mock", jobName, namespace)
//	defer func() {
//		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
//	}()
//
//	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)
//	tenant2, err = fixtures.GetTenantByExternalID(dexGraphQLClient, "guid2")
//	assert.NoError(t, err)
//	assert.Equal(t, *tenant2.Name, "name2")
//	tenant1, err = fixtures.GetTenantByExternalID(dexGraphQLClient, "guid1")
//	assert.Error(t, err)
//	assert.Nil(t, tenant1)
//}

func TestMoveSubaccounts(t *testing.T) {
	ctx := context.TODO()

	subdomain1 := "ga1"
	subdomain2 := "ga2"
	region := "local"

	subaccountRegion := "test"
	subaccountParent := "ga1"
	subaccountSubdomain := "sub1"
	// 2 global accounts, 1 subaccount, 1 runtime
	globalAccounts := []graphql.BusinessTenantMappingInput{
		{
			Name:           "ga1",
			ExternalTenant: "ga1",
			Parent:         nil,
			Subdomain:      &subdomain1,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       "test",
		},
		{
			Name:           "ga2",
			ExternalTenant: "ga2",
			Parent:         nil,
			Subdomain:      &subdomain2,
			Region:         &region,
			Type:           string(tenant.Account),
			Provider:       "test",
		},
		{
			Name:           "sub1",
			ExternalTenant: "sub1",
			Parent:         &subaccountParent,
			Subdomain:      &subaccountSubdomain,
			Region:         &subaccountRegion,
			Type:           string(tenant.Subaccount),
			Provider:       "test",
		},
	}
	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, globalAccounts)
	assert.NoError(t, err)
	defer func(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenants []graphql.BusinessTenantMappingInput) {
		err := fixtures.DeleteTenants(t, ctx, gqlClient, tenants)
		assert.NoError(t, err)
	}(t, ctx, directorInternalGQLClient, globalAccounts)

	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "ga1")
	assert.NoError(t, err)
	tenant2, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "ga2")
	assert.NoError(t, err)

	input := &graphql.RuntimeInput{
		Name:   "runtime",
		Labels: map[string]interface{}{cfg.MovedRuntimeLabelKey: "sub1"},
	}
	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant1.ID, input)
	assert.NoError(t, err)

	event1 := genMockSubaccountMoveEvent("sub1", "sub1", subaccountSubdomain, subaccountParent, cfg.MovedRuntimeLabelKey, tenant1.ID, tenant2.ID, subaccountRegion)
	setMockTenantEvents(t, []byte(genMockPage(event1, 1)), subaccountMoveSubPath)
	defer cleanupMockEvents(t, subaccountMoveSubPath)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	assert.NoError(t, err)
	jobName := "tenant-fetcher-subaccount-test"
	namespace := "compass-system"

	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-subaccount-fetcher", jobName, namespace)
	defer func() {
		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	subaccount, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "sub1")
	assert.NoError(t, err)
	log.C(ctx).Infof("actual subaccount: %+v", subaccount)
	assert.Equal(t, tenant2.InternalID, subaccount.ParentID)

	rtm := fixtures.GetRuntime(t, ctx, dexGraphQLClient, tenant2.ID, runtime.ID)
	assert.Equal(t, runtime.Name, rtm.Name)
}

//func TestMoveSubaccountsFailure(t *testing.T) {
//	ctx := context.TODO()
//
//	subdomain1 := "ga1"
//	region := "local"
//
//	subaccountRegion := "test"
//	subaccountParent := "ga1"
//	subaccountSubdomain := "sub1"
//	// global account, 2 runtimes
//	globalAccounts := []graphql.BusinessTenantMappingInput{
//		{
//			Name:           "ga1",
//			ExternalTenant: "ga1",
//			Parent:         nil,
//			Subdomain:      &subdomain1,
//			Region:         &region,
//			Type:           string(tenant.Account),
//			Provider:       "test",
//		},
//		{
//			Name:           "sub1",
//			ExternalTenant: "sub1",
//			Parent:         &subaccountParent,
//			Subdomain:      &subaccountSubdomain,
//			Region:         &subaccountRegion,
//			Type:           string(tenant.Subaccount),
//			Provider:       "test",
//		},
//	}
//	err := fixtures.WriteTenants(t, ctx, directorInternalGQLClient, globalAccounts)
//	assert.NoError(t, err)
//	defer fixtures.DeleteTenants(t, ctx, directorInternalGQLClient, globalAccounts)
//
//	tenant1, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "ga1")
//	assert.NoError(t, err)
//	input := &graphql.RuntimeInput{
//		Name:   "runtime",
//		Labels: map[string]interface{}{cfg.MovedRuntimeLabelKey: tenant1.InternalID},
//	}
//	_, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorInternalGQLClient, tenant1.InternalID, input)
//	assert.NoError(t, err)
//	_, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, directorInternalGQLClient, tenant1.InternalID, input)
//	assert.NoError(t, err)
//
//	event1 := genMockSubaccountMoveEvent("sub1", "sub1", subaccountSubdomain, subaccountParent, cfg.MovedRuntimeLabelKey, "ga1", "ga2", subaccountRegion)
//	setMockTenantEvents(t, []byte(genMockPage(event1, 1)), subaccountMoveSubPath)
//	defer cleanupMockEvents(t, subaccountMoveSubPath)
//
//	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
//	assert.NoError(t, err)
//	jobName := "tenant-fetcher-test"
//	namespace := "compass-system"
//
//	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-external-mock", jobName, namespace)
//	defer func() {
//		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
//	}()
//
//	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)
//}

func genMockGlobalAccountEvent(guid, displayName, customerID, subdomain string) string {
	return fmt.Sprintf(mockEventPattern, guid, displayName, customerID, subdomain)
}

func genMockSubaccountMoveEvent(guid, displayName, subdomain, parentGuid, movedLabelKey, sourceGlobalAccountGuid, targetGlobalAccountGuid, region string) string {
	return fmt.Sprintf(mockSubaccountMoveEventPattern, guid, displayName, subdomain, parentGuid, movedLabelKey, sourceGlobalAccountGuid, targetGlobalAccountGuid, region)
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
		bytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to set mock events: %s", string(bytes))
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
		bytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to reset mock events: %s", string(bytes))
		return
	}
	log.D().Info("Successfully reset mock events")
}
