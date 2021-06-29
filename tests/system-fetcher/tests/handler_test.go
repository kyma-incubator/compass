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
	"testing"
	"time"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSystemFetcher(t *testing.T) {
	ctx := context.TODO()

	template := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixtures.FixApplicationTemplate("temp1"))
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template.ID)

	k8sClient, err := newK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	cronjob, err := k8sClient.BatchV1beta1().CronJobs("compass-system").Get("compass-system-fetcher", metav1.GetOptions{})
	require.NoError(t, err)
	t.Log("Got the cronjob")
	jobName := "system-fetcher-test"
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
	t.Log("Creating job")
	_, err = k8sClient.BatchV1().Jobs("compass-system").Create(job)
	require.NoError(t, err)

	defer func() {
		err := k8sClient.BatchV1().Jobs("compass-system").Delete(jobName, metav1.NewDeleteOptions(0))
		require.NoError(t, err)
	}()

	elapsed := time.After(time.Minute * 2)
	for {
		select {
		case <-elapsed:
			t.Fatal("Timeout reached waiting for job to complete. Exiting...")
		default:
		}
		t.Log("Waiting for job to finish")
		job, err = k8sClient.BatchV1().Jobs("compass-system").Get(jobName, metav1.GetOptions{})
		require.NoError(t, err)
		if job.Status.Succeeded > 0 {
			break
		}
		time.Sleep(time.Second * 2)
	}

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
			fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app.ID)
		}
	}()

	require.ElementsMatch(t, expectedApps, actualApps)
}
