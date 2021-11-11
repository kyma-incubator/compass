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
)

func TestTenantFetcherSuccess(t *testing.T) {
	ctx := context.TODO()

	event1 := genMockEvent("guid1", "name1", "customerID1", "subdomain1")
	event2 := genMockEvent("guid2", "name2", "customerID2", "subdomain2")
	event3 := genMockEvent("guid3", "name3", "customerID3", "subdomain3")
	setMockTenantEvents(t, []byte(genMockPage(strings.Join([]string{event1, event2, event3}, ","), 3)))
	defer cleanupMockTenantEvents(t)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "tenant-fetcher-test"
	namespace := "compass-system"

	k8s.CreateJobByCronJob(t, ctx, k8sClient, "compass-tenant-fetcher-external-mock", jobName, namespace)
	defer func() {
		k8s.DeleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	tenant, err := fixtures.GetTenantByExternalID(dexGraphQLClient, "guid1")
	fmt.Printf(">>>>> %+v\n", tenant)
	assert.NoError(t, err)
	assert.Equal(t, tenant.Name, "name1")
}

func genMockEvent(guid, displayName, customerID, subdomain string) string {
	return fmt.Sprintf(mockEventPattern, guid, displayName, customerID, subdomain)
}

func genMockPage(events string, numEvents int) string {
	return fmt.Sprintf(mockEventsPagePattern, numEvents, 1, events)
}

func setMockTenantEvents(t *testing.T, mockEvents []byte) {
	reader := bytes.NewReader(mockEvents)
	fmt.Println(">>>>>", string(mockEvents))
	response, err := http.DefaultClient.Post(cfg.ExternalSvcMockURL+"/tenant-fetcher/global-account-create/configure", "application/json", reader)
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

func cleanupMockTenantEvents(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, cfg.ExternalSvcMockURL+"/tenant-fetcher/global-account-create/reset", nil)
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
