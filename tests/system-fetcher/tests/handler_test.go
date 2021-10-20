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
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"k8s.io/client-go/rest"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/tests/pkg/clients"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testPkg "github.com/kyma-incubator/compass/tests/pkg/webhook"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const mockSystemFormat = `{
		"systemNumber": "%d",
		"displayName": "name%d",
		"productDescription": "description",
		"type": "type1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
}`

func TestSystemFetcherSuccess(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"type": "type1",
		"prop": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`)

	setMockSystems(t, mockSystems)
	defer cleanupMockSystems(t)

	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixtures.FixApplicationTemplate("temp1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateInput2 := fixtures.FixApplicationTemplate("temp2")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "system-fetcher-test"
	namespace := "compass-system"
	createJobByCronJob(t, ctx, k8sClient, "compass-system-fetcher", jobName, namespace)
	defer func() {
		deleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	waitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)
	description := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				ApplicationTemplateID: &template.ID,
			},
		},
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
		},
	}

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
			},
		})
	}
	defer func() {
		for _, app := range resp.Data {
			fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
		}
	}()

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherSuccessForMoreThanOnePage(t *testing.T) {
	ctx := context.TODO()

	setMultipleMockSystemsResponses(t)
	defer cleanupMockSystems(t)

	appTemplateInput2 := fixtures.FixApplicationTemplate("temp2")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixtures.FixApplicationTemplate("temp1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "system-fetcher-test"
	namespace := "compass-system"
	createJobByCronJob(t, ctx, k8sClient, "compass-system-fetcher", jobName, namespace)
	defer func() {
		deleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	waitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)

	req2 := fixtures.FixApplicationsPageableRequest(200, string(resp.PageInfo.EndCursor))
	var resp2 directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req2, &resp2)
	require.NoError(t, err)
	resp.Data = append(resp.Data, resp2.Data...)

	description := "description"
	expectedCount := cfg.SystemFetcherPageSize
	if expectedCount > 1 {
		expectedCount++
	}
	expectedApps := getFixExpectedMockSystems(expectedCount, description)

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
			},
		})
	}
	defer func() {
		for _, app := range resp.Data {
			fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
		}
	}()

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherDuplicateSystems(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"type": "type1",
		"prop": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "3",
		"displayName": "name1",
		"productDescription": "description",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`)

	setMockSystems(t, mockSystems)
	defer cleanupMockSystems(t)

	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixtures.FixApplicationTemplate("temp1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateInput2 := fixtures.FixApplicationTemplate("temp2")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "system-fetcher-test"
	namespace := "compass-system"
	createJobByCronJob(t, ctx, k8sClient, "compass-system-fetcher", jobName, namespace)
	defer func() {
		deleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	waitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	description := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				ApplicationTemplateID: &template.ID,
			},
			Labels: directorSchema.Labels{
				"managed": "true",
			},
		},
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
			Labels: directorSchema.Labels{
				"managed": "true",
			},
		},
		{
			Application: directorSchema.Application{
				Name:        "name1",
				Description: &description,
			},
			Labels: directorSchema.Labels{
				"managed": "true",
			},
		},
	}

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
			},
			Labels: directorSchema.Labels{
				"managed": app.Labels["managed"],
			},
		})
	}
	defer func() {
		for _, app := range resp.Data {
			fixtures.CleanupApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
		}
	}()

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherCreateAndDelete(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"type": "type1",
		"prop": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}, {
		"systemNumber": "3",
		"displayName": "name3",
		"productDescription": "description",
		"prop": "val2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`)

	setMockSystems(t, mockSystems)

	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixtures.FixApplicationTemplate("temp1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateInput2 := fixtures.FixApplicationTemplate("temp2")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "system-fetcher-test"
	namespace := "compass-system"
	createJobByCronJob(t, ctx, k8sClient, "compass-system-fetcher", jobName, namespace)
	defer func(jobName string) {
		deleteJob(t, ctx, k8sClient, jobName, namespace)
	}(jobName)

	waitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)
	description := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				ApplicationTemplateID: &template.ID,
			},
		},
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
		},
		{
			Application: directorSchema.Application{
				Name:                  "name3",
				Description:           &description,
				ApplicationTemplateID: &template2.ID,
			},
		},
	}

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
			},
		})
	}

	require.ElementsMatch(t, expectedApps, actualApps)

	mockSystems = []byte(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"type": "type1",
		"prop": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {
			"lifecycleStatus": "DELETED"
		}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}, {
		"systemNumber": "3",
		"displayName": "name3",
		"productDescription": "description",
		"prop": "val2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {
			"lifecycleStatus": "DELETED"
		}
	}]`)

	setMockSystems(t, mockSystems)

	t.Log("Unlock the mock application webhook")
	testPkg.UnlockWebhook(t, testPkg.BuildOperationFullPath(cfg.ExternalSvcMockURL+"/"))

	var idToDelete string
	var idToWaitForDeletion string
	for _, app := range resp.Data {
		if app.Name == "name3" {
			idToDelete = app.ID
		}
		if app.Name == "name1" {
			idToWaitForDeletion = app.ID
		}
	}
	fixtures.UnregisterAsyncApplicationInTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), idToDelete)

	jobName = "system-fetcher-test2"
	createJobByCronJob(t, ctx, k8sClient, "compass-system-fetcher", jobName, namespace)
	defer func() {
		deleteJob(t, ctx, k8sClient, jobName, namespace)
	}()

	waitForJobToSucceed(t, ctx, k8sClient, jobName, namespace)

	testPkg.UnlockWebhook(t, testPkg.BuildOperationFullPath(cfg.ExternalSvcMockURL+"/"))

	t.Log("Waiting for asynchronous deletion to take place")
	waitForDeleteOperation(ctx, t, idToWaitForDeletion)

	req = fixtures.FixGetApplicationsRequestWithPagination()
	var resp2 directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp2)
	require.NoError(t, err)

	expectedApps = []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
		},
	}

	actualApps = make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp2.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
			},
		})
	}

	defer func() {
		for _, app := range resp2.Data {
			fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app.ID)
		}
	}()

	require.ElementsMatch(t, expectedApps, actualApps)
}

func waitForDeleteOperation(ctx context.Context, t *testing.T, appID string) {
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)

	k8sClient, err := client.NewForConfig(cfg)
	operationsK8sClient := k8sClient.Operations("compass-system")
	opName := fmt.Sprintf("application-%s", appID)

	require.Eventually(t, func() bool {
		op, err := operationsK8sClient.Get(ctx, opName, metav1.GetOptions{})
		if err != nil {
			t.Logf("Error getting operation %s: %s", opName, err)
			return false
		}

		if op.Status.Phase != "Success" {
			t.Logf("Operation %s is not in Success phase. Current state: %s", opName, op.Status.Phase)
			return false
		}
		return true
	}, time.Minute*3, time.Second*5, "Waiting for delete operation timed out.")
}

func waitForJobToSucceed(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	elapsed := time.After(time.Minute * 15)
	for {
		select {
		case <-elapsed:
			t.Fatal("Timeout reached waiting for job to complete. Exiting...")
		default:
		}
		t.Log("Waiting for job to finish")
		job, err := k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		require.NoError(t, err)
		if job.Status.Failed > 0 {
			t.Fatal("Job has failed. Exiting...")
		}
		if job.Status.Succeeded > 0 {
			break
		}
		time.Sleep(time.Second * 2)
	}
}

func deleteJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, jobName, namespace string) {
	t.Log("Deleting test job")
	err := k8sClient.BatchV1().Jobs(namespace).Delete(ctx, jobName, *metav1.NewDeleteOptions(0))
	require.NoError(t, err)

	elapsed := time.After(time.Minute * 2)
	for {
		select {
		case <-elapsed:
			t.Fatal("Timeout reached waiting for job to be deleted. Exiting...")
		default:
		}
		t.Log("Waiting for job to be deleted")
		_, err = k8sClient.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			break
		}
		time.Sleep(time.Second * 2)
	}
	t.Log("Test job deleted")
}

func createJobByCronJob(t *testing.T, ctx context.Context, k8sClient *kubernetes.Clientset, cronJobName, jobName, namespace string) {
	cronjob, err := k8sClient.BatchV1beta1().CronJobs(namespace).Get(ctx, cronJobName, metav1.GetOptions{})
	require.NoError(t, err)
	t.Log("Got the cronjob")

	job := &v1.Job{
		Spec: v1.JobSpec{
			Parallelism:             cronjob.Spec.JobTemplate.Spec.Parallelism,
			Completions:             cronjob.Spec.JobTemplate.Spec.Completions,
			ActiveDeadlineSeconds:   cronjob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds,
			BackoffLimit:            cronjob.Spec.JobTemplate.Spec.BackoffLimit,
			Selector:                cronjob.Spec.JobTemplate.Spec.Selector,
			ManualSelector:          cronjob.Spec.JobTemplate.Spec.ManualSelector,
			Template:                cronjob.Spec.JobTemplate.Spec.Template,
			TTLSecondsAfterFinished: cronjob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
		},
	}
	t.Log("Creating test job")
	_, err = k8sClient.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Log("Test job created")
}

func setMockSystems(t *testing.T, mockSystems []byte) {
	reader := bytes.NewReader(mockSystems)
	response, err := http.DefaultClient.Post(cfg.ExternalSvcMockURL+"/systemfetcher/configure", "application/json", reader)
	require.NoError(t, err)
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	if response.StatusCode != http.StatusOK {
		bytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to set mock systems: %s", string(bytes))
	}
}

func setMultipleMockSystemsResponses(t *testing.T) {
	mockSystems := []byte(getFixMockSystemsJSON(cfg.SystemFetcherPageSize, 0))
	setMockSystems(t, mockSystems)

	mockSystems2 := []byte(getFixMockSystemsJSON(1, cfg.SystemFetcherPageSize))
	setMockSystems(t, mockSystems2)
}

func getFixMockSystemsJSON(count, startingNumber int) string {
	result := "["
	for i := 0; i < count; i++ {
		systemNumber := startingNumber + i
		result = result + fmt.Sprintf(mockSystemFormat, systemNumber, systemNumber)
		if i < count-1 {
			result = result + ","
		}
	}
	return result + "]"
}

func getFixExpectedMockSystems(count int, description string) []directorSchema.ApplicationExt {
	result := make([]directorSchema.ApplicationExt, count)
	for i := 0; i < count; i++ {
		result[i] = directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:        fmt.Sprintf("name%d", i),
				Description: &description,
			},
		}
	}
	return result
}

func cleanupMockSystems(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, cfg.ExternalSvcMockURL+"/systemfetcher/reset", nil)
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
		t.Fatalf("Failed to reset mock systems: %s", string(bytes))
		return
	}
	log.D().Info("Successfully reset mock systems")
}
