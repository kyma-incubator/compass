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

	"github.com/stretchr/testify/assert"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	testPkg "github.com/kyma-incubator/compass/tests/pkg/webhook"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const (
	systemFetcherJobName      = "system-fetcher-test"
	systemFetcherJobNamespace = "compass-system"
	systemFetcherCronJobName  = "compass-system-fetcher"
	mockSystemFormat          = `{
		"systemNumber": "%d",
		"displayName": "name%d",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}`
	defaultMockSystems = `[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {}
	}]`

	nameLabelKey           = "displayName"
	namePlaceholder        = "name"
	displayNamePlaceholder = "display-name"
)

var (
	additionalSystemLabels = directorSchema.Labels{
		nameLabelKey: "{{name}}",
	}
)

func TestSystemFetcherSuccess(t *testing.T) {
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(defaultMockSystems, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID, "")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	description := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	req := fixtures.FixGetApplicationBySystemNumberRequest("1")
	var appResp directorSchema.ApplicationExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &appResp)
	require.NoError(t, err)
	require.Equal(t, "name1", appResp.Name)

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherSuccessExpectORDWebhook(t *testing.T) {
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(defaultMockSystems, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithoutWebhooks(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplateWithoutWebhooks(appTemplateName2, intSys.ID, "")
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	description := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}
	require.ElementsMatch(t, expectedApps, actualApps)

	for _, app := range resp.Data {
		if app.Name == "name1" {
			assert.Equal(t, 1, len(app.Webhooks))
			assert.Equal(t, fmt.Sprintf("%s%s", baseUrl, "/.well-known/open-resource-discovery"), str.PtrStrToStr(app.Webhooks[0].URL))
			assert.Equal(t, "sap:cmp-mtls:v1", str.PtrStrToStr(app.Webhooks[0].Auth.AccessStrategy))
			assert.Equal(t, "OPEN_RESOURCE_DISCOVERY", app.Webhooks[0].Type.String())
		} else {
			assert.Equal(t, 0, len(app.Webhooks))
		}
	}
}

func TestSystemFetcherSuccessMissingORDWebhookEmptyBaseURL(t *testing.T) {
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithoutWebhooks(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplateWithoutWebhooks(appTemplateName2, intSys.ID, "")
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	description := "description"
	baseUrl := ""
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}
	require.ElementsMatch(t, expectedApps, actualApps)

	for _, app := range resp.Data {
		assert.Equal(t, 0, len(app.Webhooks))
	}
}

func TestSystemFetcherSuccessForMoreThanOnePage(t *testing.T) {
	ctx := context.TODO()

	setMultipleMockSystemsResponses(t, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID, "")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID, ""))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)

	req2 := fixtures.FixApplicationsPageableRequest(200, string(resp.PageInfo.EndCursor))
	var resp2 directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req2, &resp2)
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
				SystemNumber:          app.SystemNumber,
				IntegrationSystemID:   app.IntegrationSystemID,
			},
			Labels: app.Labels,
		})
	}
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherDuplicateSystemsForTwoTenants(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(fmt.Sprintf(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {}
	}]`, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	setMockSystems(t, mockSystems, tenant.TestTenants.GetSystemFetcherTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID, "")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	description := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false),
		},
	}

	respDefaultTenant, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range respDefaultTenant.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

// fail
func TestSystemFetcherDuplicateSystems(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(fmt.Sprintf(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "3",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID, "")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, systemFetcherJobName, systemFetcherJobNamespace)

	description := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false),
		},
		{
			Application: directorSchema.Application{
				Name:         "name1",
				Description:  &description,
				SystemNumber: str.Ptr("3"),
			},
			Labels: applicationLabels("name1", "", "", false),
		},
	}

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
				SystemNumber:          app.SystemNumber,
				IntegrationSystemID:   app.IntegrationSystemID,
			},
			Labels: app.Labels,
		})
	}
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherCreateAndDelete(t *testing.T) {
	ctx := context.TODO()

	mockSystems := []byte(fmt.Sprintf(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}, {
		"systemNumber": "3",
		"displayName": "name3",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"%s": "val2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}]`, cfg.SystemInformationSourceKey, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	appTemplateName1 := createAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID, "val1"))
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := createAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID, "val2")
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)
	jobName := "system-fetcher-test"
	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, jobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, jobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, jobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, systemFetcherJobNamespace)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)
	description := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description,
				ApplicationTemplateID: &template.ID,
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true),
		},
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
			Labels: applicationLabels("name2", "", "", false),
		},
		{
			Application: directorSchema.Application{
				Name:                  "name3",
				Description:           &description,
				ApplicationTemplateID: &template2.ID,
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name3", appTemplateName2, intSys.ID, true),
		},
	}

	actualApps := make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
				IntegrationSystemID:   app.IntegrationSystemID,
			},
			Labels: app.Labels,
		})
	}

	require.ElementsMatch(t, expectedApps, actualApps)

	mockSystems = []byte(fmt.Sprintf(`[{
		"systemNumber": "1",
		"displayName": "name1",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type1",
		"%s": "val1",
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
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {}
	}, {
		"systemNumber": "3",
		"displayName": "name3",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"%s": "val2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {},
		"additionalAttributes": {
			"lifecycleStatus": "DELETED"
		}
	}]`, cfg.SystemInformationSourceKey, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())

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
	fixtures.UnregisterAsyncApplicationInTenant(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), idToDelete)

	jobName = "system-fetcher-test2"
	k8s.CreateJobByCronJob(t, ctx, k8sClient, systemFetcherCronJobName, jobName, systemFetcherJobNamespace)
	defer k8s.DeleteJob(t, ctx, k8sClient, jobName, systemFetcherJobNamespace)
	defer k8s.PrintJobLogs(t, ctx, k8sClient, jobName, systemFetcherJobNamespace, cfg.SystemFetcherContainerName, false)

	k8s.WaitForJobToSucceed(t, ctx, k8sClient, jobName, systemFetcherJobNamespace)

	testPkg.UnlockWebhook(t, testPkg.BuildOperationFullPath(cfg.ExternalSvcMockURL+"/"))

	t.Log("Waiting for asynchronous deletion to take place")
	waitForDeleteOperation(ctx, t, idToWaitForDeletion)

	req = fixtures.FixGetApplicationsRequestWithPagination()
	var resp2 directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp2)
	require.NoError(t, err)

	expectedApps = []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description,
			},
			Labels: applicationLabels("name2", "", "", false),
		},
	}

	actualApps = make([]directorSchema.ApplicationExt, 0, len(expectedApps))
	for _, app := range resp2.Data {
		actualApps = append(actualApps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				ApplicationTemplateID: app.ApplicationTemplateID,
				IntegrationSystemID:   app.IntegrationSystemID,
			},
			Labels: app.Labels,
		})
	}

	for _, app := range resp2.Data {
		defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app.ID)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func waitForDeleteOperation(ctx context.Context, t *testing.T, appID string) {
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)

	k8sClient, err := client.NewForConfig(cfg)
	operationsK8sClient := k8sClient.Operations(systemFetcherJobNamespace)
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

func setMockSystems(t *testing.T, mockSystems []byte, tenant string) {
	reader := bytes.NewReader(mockSystems)
	url := cfg.ExternalSvcMockURL + fmt.Sprintf("/systemfetcher/configure?tenant=%s", tenant)
	response, err := http.DefaultClient.Post(url, "application/json", reader)
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	if response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to set mock systems: %s", string(bodyBytes))
	}
}

func setMultipleMockSystemsResponses(t *testing.T, tenant string) {
	mockSystems := []byte(getFixMockSystemsJSON(cfg.SystemFetcherPageSize, 0))
	setMockSystems(t, mockSystems, tenant)

	mockSystems2 := []byte(getFixMockSystemsJSON(1, cfg.SystemFetcherPageSize))
	setMockSystems(t, mockSystems2, tenant)
}

func retrieveAppsForTenant(t *testing.T, ctx context.Context, tenant string) (directorSchema.ApplicationPageExt, []directorSchema.ApplicationExt) {
	req := fixtures.FixGetApplicationsRequestWithPagination()

	var resp directorSchema.ApplicationPageExt
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant, req, &resp)
	require.NoError(t, err)

	apps := make([]directorSchema.ApplicationExt, 0)
	for _, app := range resp.Data {
		apps = append(apps, directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:                  app.Application.Name,
				Description:           app.Application.Description,
				BaseURL:               app.Application.BaseURL,
				ApplicationTemplateID: app.ApplicationTemplateID,
				SystemNumber:          app.SystemNumber,
				IntegrationSystemID:   app.IntegrationSystemID,
			},
			Labels: app.Labels,
		})
	}
	return resp, apps
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
		systemName := fmt.Sprintf("name%d", i)
		result[i] = directorSchema.ApplicationExt{
			Application: directorSchema.Application{
				Name:         systemName,
				Description:  &description,
				SystemNumber: str.Ptr(fmt.Sprintf("%d", i)),
			},
			Labels: applicationLabels(systemName, "", "", false),
		}
	}
	return result
}

func cleanupMockSystems(t *testing.T) {
	req, err := http.NewRequest(http.MethodDelete, cfg.ExternalSvcMockURL+"/systemfetcher/reset", nil)
	require.NoError(t, err)

	response, err := http.DefaultClient.Do(req)
	defer func() {
		if err := response.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to reset mock systems: %s", string(bodyBytes))
		return
	}
	log.D().Info("Successfully reset mock systems")
}

func applicationLabels(name, appTemplateName, integrationSystemID string, fromTemplate bool) directorSchema.Labels {
	labels := directorSchema.Labels{
		"managed":              "true",
		"name":                 fmt.Sprintf("mp-%s", name),
		"ppmsProductVersionId": "12345",
		"productId":            "XXX",
		"integrationSystemID":  integrationSystemID,
	}

	if fromTemplate {
		labels[nameLabelKey] = name
		labels["applicationType"] = appTemplateName
	}

	return labels
}

func fixApplicationTemplate(name, intSystemID, systemRole string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        name,
		Description: str.Ptr("template description"),
		ApplicationInput: &directorSchema.ApplicationRegisterInput{
			Name:        fmt.Sprintf("{{%s}}", namePlaceholder),
			Description: ptr.String(fmt.Sprintf("{{%s}}", displayNamePlaceholder)),
			Labels:      additionalSystemLabels,
			Webhooks: []*directorSchema.WebhookInput{{
				Type: directorSchema.WebhookTypeConfigurationChanged,
				URL:  ptr.String("http://url.com"),
			}},
			HealthCheckURL:      ptr.String("http://url.valid"),
			IntegrationSystemID: &intSystemID,
		},
		Placeholders: []*directorSchema.PlaceholderDefinitionInput{
			{
				Name: namePlaceholder,
			},
			{
				Name: displayNamePlaceholder,
			},
		},
		AccessLevel: directorSchema.ApplicationTemplateAccessLevelGlobal,
		Labels: directorSchema.Labels{
			cfg.SelfRegDistinguishLabelKey: []interface{}{cfg.SelfRegDistinguishLabelValue},
		},
	}

	if len(systemRole) > 0 {
		appTemplateInput.Labels[cfg.TemplateLabelFilter] = systemRole
	}

	return appTemplateInput
}

func fixApplicationTemplateWithoutWebhooks(name, intSystemID, systemRole string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        name,
		Description: str.Ptr("template description"),
		ApplicationInput: &directorSchema.ApplicationRegisterInput{
			Name:                fmt.Sprintf("{{%s}}", namePlaceholder),
			Description:         ptr.String(fmt.Sprintf("{{%s}}", displayNamePlaceholder)),
			Labels:              additionalSystemLabels,
			HealthCheckURL:      ptr.String("http://url.valid"),
			IntegrationSystemID: &intSystemID,
		},
		Placeholders: []*directorSchema.PlaceholderDefinitionInput{
			{
				Name: namePlaceholder,
			},
			{
				Name: displayNamePlaceholder,
			},
		},
		AccessLevel: directorSchema.ApplicationTemplateAccessLevelGlobal,
		Labels: directorSchema.Labels{
			cfg.SelfRegDistinguishLabelKey: []interface{}{cfg.SelfRegDistinguishLabelValue},
		},
	}

	if len(systemRole) > 0 {
		appTemplateInput.Labels[cfg.TemplateLabelFilter] = systemRole
	}

	return appTemplateInput
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}
