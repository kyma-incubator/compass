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
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/kyma-incubator/compass/tests/pkg/util"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	testPkg "github.com/kyma-incubator/compass/tests/pkg/webhook"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/rest"
)

const (
	systemFetcherJobNamespace = "compass-system"
	tenantHeader              = "Tenant"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {"systemSCPLandscapeID":"cf-eu10"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`

	singleMockSystem = `[{
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
		"additionalAttributes": {"systemSCPLandscapeID":"cf-eu10"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`

	nameLabelKey           = "displayName"
	namePlaceholder        = "name"
	displayNamePlaceholder = "display-name"
	regionLabelKey         = "region"
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

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithDefaultSystemRoles(appTemplateName1, intSys.ID))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	description1 := "name1"
	description2 := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, "cf-eu10"),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
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

func TestSystemFetcherSuccessForCustomerTenant(t *testing.T) {
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(defaultMockSystems, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultCustomerTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), fixApplicationTemplateWithDefaultSystemRoles(appTemplateName1, intSys.ID))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultCustomerTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultCustomerTenantID(), 2)

	description1 := "name1"
	description2 := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, "cf-eu10"),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultCustomerTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), app)
	}

	req := fixtures.FixGetApplicationBySystemNumberRequest("1")
	var appResp directorSchema.ApplicationExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultCustomerTenantID(), req, &appResp)
	require.NoError(t, err)
	require.Equal(t, "name1", appResp.Name)

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherOnNewGASuccess(t *testing.T) {
	gaExternalID := tenant.TestTenants.GetIDByName(t, tenant.TestSystemFetcherOnNewGAName)
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(singleMockSystem, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, gaExternalID)
	defer cleanupMockSystems(t)

	tenantInput := directorSchema.BusinessTenantMappingInput{
		Name:           "ga1",
		ExternalTenant: gaExternalID,
		Parents:        []*string{},
		Subdomain:      str.Ptr("ga1"),
		Region:         str.Ptr("cf-eu10"),
		Type:           string(tenant.Account),
		Provider:       "e2e-test-provider",
		LicenseType:    str.Ptr("LICENSETYPE"),
	}

	err := fixtures.WriteTenant(t, ctx, directorInternalGQLClient, tenantInput)
	assert.NoError(t, err)
	defer cleanupTenant(t, ctx, directorInternalGQLClient, gaExternalID)

	var tenant *directorSchema.Tenant
	require.Eventually(t, func() bool {
		tenant, err = fixtures.GetTenantByExternalID(certSecuredGraphQLClient, gaExternalID)
		if tenant == nil {
			t.Logf("Waiting for global account %s to be read", gaExternalID)
			return false
		}
		assert.NoError(t, err)
		return true
	}, time.Minute*1, time.Second*1, "Waiting for tenants retrieval.")

	t.Logf("Created tenant: %+v", tenant)
	waitForApplicationsToBeProcessed(ctx, t, gaExternalID, 1)
	resp, actualApps := retrieveAppsForTenant(t, ctx, gaExternalID)
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, gaExternalID, app)
	}
	require.Equal(t, 1, len(actualApps))

	req := fixtures.FixGetApplicationBySystemNumberRequest("1")
	var appResp directorSchema.ApplicationExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, gaExternalID, req, &appResp)
	require.NoError(t, err)
	require.Equal(t, "name1", appResp.Name)
}

func TestSystemFetcherSuccessWithMultipleLabelValues(t *testing.T) {
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
		"additionalAttributes": {"systemSCPLandscapeID":"cf-eu10"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"%s": "val2",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithSystemRoles(appTemplateName1, intSys.ID, []interface{}{"val1", "val2"}))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	description1 := "name1"
	description2 := "name2"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, "cf-eu10"),
		},
		{
			Application: directorSchema.Application{
				Name:                  "name2",
				Description:           &description2,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("2"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name2", appTemplateName1, intSys.ID, true, ""),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

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

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithoutWebhooksWithSystemRole(appTemplateName1, intSys.ID, []string{"val1"}))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplateWithoutWebhooks(appTemplateName2, intSys.ID)
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	description1 := "name1"
	description2 := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, "cf-eu10"),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithoutWebhooksWithSystemRole(appTemplateName1, intSys.ID, []string{"val1"}))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplateWithoutWebhooks(appTemplateName2, intSys.ID)
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	description1 := "name1"
	description2 := "description"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, ""),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
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

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplate(appTemplateName1, intSys.ID))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), cfg.SystemFetcherPageSize)

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
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherSuccessForMultipleTenants(t *testing.T) {
	ctx := context.TODO()
	tenants := []string{tenant.TestTenants.GetDefaultTenantID(), tenant.TestTenants.GetDefaultCustomerTenantID()}
	tenantNames := []string{"def", "cus"}
	for index, tenantID := range tenants {
		setMultipleMockSystemsResponses(t, tenantID)
		defer cleanupMockSystems(t)

		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, fmt.Sprintf("int-sys-%s", tenantNames[index]))
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

		appTemplateName2 := fixtures.CreateAppTemplateName(fmt.Sprintf("temp2%s", tenantNames[index]))
		appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
		appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
		template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, appTemplateInput2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, template2)
		require.NoError(t, err)
		require.NotEmpty(t, template2.ID)

		appTemplateName1 := fixtures.CreateAppTemplateName(fmt.Sprintf("temp1%s", tenantNames[index]))
		template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantID, fixApplicationTemplate(appTemplateName1, intSys.ID))
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantID, template)
		require.NoError(t, err)
		require.NotEmpty(t, template.ID)
	}
	for _, tenantID := range tenants {
		triggerSync(t, tenantID)
	}
	for _, tenantID := range tenants {
		waitForApplicationsToBeProcessed(ctx, t, tenantID, cfg.SystemFetcherPageSize)
	}
	for _, tenantID := range tenants {
		req := fixtures.FixGetApplicationsRequestWithPagination()
		var resp directorSchema.ApplicationPageExt
		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, &resp)
		require.NoError(t, err)

		req2 := fixtures.FixApplicationsPageableRequest(200, string(resp.PageInfo.EndCursor))
		var resp2 directorSchema.ApplicationPageExt
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req2, &resp2)
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
			defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, app)
		}
		require.ElementsMatch(t, expectedApps, actualApps)
	}
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	setMockSystems(t, mockSystems, tenant.TestTenants.GetSystemFetcherTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithDefaultSystemRoles(appTemplateName1, intSys.ID))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)
	triggerSync(t, tenant.TestTenants.GetSystemFetcherTenantID())

	description1 := "name1"
	description2 := "description"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, ""),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				BaseURL:      &baseUrl,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
		},
	}

	respDefaultTenant, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range respDefaultTenant.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherSuccessForRegionalAppTemplates(t *testing.T) {
	ctx := context.TODO()
	region1 := "cf-eu10"
	region2 := "cf-eu20"

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
		"additionalAttributes": {"systemSCPLandscapeID":"%s"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {"systemSCPLandscapeID":"%s"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey, region1, cfg.SystemInformationSourceKey, region2))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixRegionalApplicationTemplateWithSystemRoles(appTemplateName1, intSys.ID, []interface{}{"val1"}, region1))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template1)
	require.NoError(t, err)
	require.NotEmpty(t, template1.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp1")
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixRegionalApplicationTemplateWithSystemRoles(appTemplateName2, intSys.ID, []interface{}{"val1"}, region2))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	description1 := "name1"
	description2 := "name2"
	baseUrl := "http://mainurl.com"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template1.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, region1),
		},
		{
			Application: directorSchema.Application{
				Name:                  "name2",
				Description:           &description2,
				BaseURL:               &baseUrl,
				ApplicationTemplateID: &template2.ID,
				SystemNumber:          str.Ptr("2"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name2", appTemplateName2, intSys.ID, true, region2),
		},
	}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherNotFetchMissingRegionForRegionalAppTemplates(t *testing.T) {
	ctx := context.TODO()
	region1 := "cf-eu10"
	region2 := "cf-eu20"

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
		"additionalAttributes": {"systemSCPLandscapeID":"%s"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	},{
		"systemNumber": "2",
		"displayName": "name2",
		"productDescription": "description",
		"productId": "XXX",
		"ppmsProductVersionId": "12345",
		"type": "type2",
		"%s": "val1",
		"baseUrl": "",
		"infrastructureProvider": "",
		"additionalUrls": {"mainUrl":"http://mainurl.com"},
		"additionalAttributes": {"systemSCPLandscapeID":"%s"},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey, region1, cfg.SystemInformationSourceKey, region1))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixRegionalApplicationTemplateWithSystemRoles(appTemplateName, intSys.ID, []interface{}{"val1"}, region2))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 0)

	expectedApps := []directorSchema.ApplicationExt{}

	resp, actualApps := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithDefaultSystemRoles(appTemplateName1, intSys.ID))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 3)

	description1 := "name1"
	description2 := "description"

	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				ApplicationTemplateID: &template.ID,
				SystemNumber:          str.Ptr("1"),
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, ""),
		},
		{
			Application: directorSchema.Application{
				Name:         "name2",
				Description:  &description2,
				SystemNumber: str.Ptr("2"),
			},
			Labels: applicationLabels("name2", "", "", false, ""),
		},
		{
			Application: directorSchema.Application{
				Name:         "name1",
				Description:  &description2,
				SystemNumber: str.Ptr("3"),
			},
			Labels: applicationLabels("name1", "", "", false, ""),
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
	}]`, cfg.SystemInformationSourceKey, cfg.SystemInformationSourceKey))

	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), fixApplicationTemplateWithSystemRoles(appTemplateName1, intSys.ID, []interface{}{"val1"}))
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	appTemplateInput2 := fixApplicationTemplateWithSystemRoles(appTemplateName2, intSys.ID, []interface{}{"val2"})
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 3)

	req := fixtures.FixGetApplicationsRequestWithPagination()
	var resp directorSchema.ApplicationPageExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &resp)
	require.NoError(t, err)

	description1 := "name1"
	description2 := "description"
	description3 := "name3"
	expectedApps := []directorSchema.ApplicationExt{
		{
			Application: directorSchema.Application{
				Name:                  "name1",
				Description:           &description1,
				ApplicationTemplateID: &template.ID,
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name1", appTemplateName1, intSys.ID, true, ""),
		},
		{
			Application: directorSchema.Application{
				Name:        "name2",
				Description: &description2,
			},
			Labels: applicationLabels("name2", "", "", false, ""),
		},
		{
			Application: directorSchema.Application{
				Name:                  "name3",
				Description:           &description3,
				ApplicationTemplateID: &template2.ID,
				IntegrationSystemID:   &intSys.ID,
			},
			Labels: applicationLabels("name3", appTemplateName2, intSys.ID, true, ""),
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
		},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		"additionalAttributes": {},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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
		},
		"businessTypeId": "tbtID",
		"businessTypeDescription": "tbt description name"
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

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

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
				Description: &description2,
			},
			Labels: applicationLabels("name2", "", "", false, ""),
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
		defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app.ID)
	}

	require.ElementsMatch(t, expectedApps, actualApps)
}

func TestSystemFetcherPreserveSystemStatusOnUpdate(t *testing.T) {
	ctx := context.TODO()
	mockSystems := []byte(fmt.Sprintf(defaultMockSystems, cfg.SystemInformationSourceKey))
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())
	defer cleanupMockSystems(t)

	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "integration-system")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*directorSchema.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, cfg.GatewayOauth)

	appTemplateName1 := fixtures.CreateAppTemplateName("temp1")
	t.Logf("Create Application Template with name %s", appTemplateName1)
	appTemplateInput1 := fixApplicationTemplateWithDefaultSystemRoles(appTemplateName1, intSys.ID)
	template, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template)
	require.NoError(t, err)
	require.NotEmpty(t, template.ID)

	appTemplateName2 := fixtures.CreateAppTemplateName("temp2")
	t.Logf("Create Application Template with name %s", appTemplateName2)
	appTemplateInput2 := fixApplicationTemplate(appTemplateName2, intSys.ID)
	appTemplateInput2.Webhooks = append(appTemplateInput2.Webhooks, testPkg.BuildMockedWebhook(cfg.ExternalSvcMockURL+"/", directorSchema.WebhookTypeUnregisterApplication))
	template2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), template2)
	require.NoError(t, err)
	require.NotEmpty(t, template2.ID)

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	resp, _ := retrieveAppsForTenant(t, ctx, tenant.TestTenants.GetDefaultTenantID())
	for _, app := range resp.Data {
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app)
	}

	t.Log("Get Application with system number 1")
	req := fixtures.FixGetApplicationBySystemNumberRequest("1")
	var appResp1 directorSchema.ApplicationExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &appResp1)
	require.NoError(t, err)

	t.Log("Get Application with system number 2")
	req = fixtures.FixGetApplicationBySystemNumberRequest("2")
	var appResp2 directorSchema.ApplicationExt
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &appResp2)
	require.NoError(t, err)

	// Update the Applications with a CONNECTED status condition
	connectedStatus := directorSchema.ApplicationStatusConditionConnected
	updateInput := fixtures.FixSampleApplicationUpdateInput("after")
	updateInput.StatusCondition = &connectedStatus

	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	requestUpdateApp1 := fixtures.FixUpdateApplicationRequest(appResp1.ID, updateInputGQL)
	requestUpdateApp2 := fixtures.FixUpdateApplicationRequest(appResp2.ID, updateInputGQL)

	t.Log("Update Application with system number 1")
	updatedApp1 := directorSchema.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, requestUpdateApp1, &updatedApp1)
	require.NoError(t, err)
	require.Equal(t, updatedApp1.Status.Condition.String(), connectedStatus.String())

	t.Log("Update Application with system number 2")
	updatedApp2 := directorSchema.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, requestUpdateApp2, &updatedApp2)
	require.NoError(t, err)
	require.Equal(t, updatedApp2.Status.Condition.String(), connectedStatus.String())

	// setup mock systems for second job run
	setMockSystems(t, mockSystems, tenant.TestTenants.GetDefaultTenantID())

	triggerSync(t, tenant.TestTenants.GetDefaultTenantID())
	waitForApplicationsToBeProcessed(ctx, t, tenant.TestTenants.GetDefaultTenantID(), 2)

	// Assert the previously updated Applications still contain their updated StatusCondition
	t.Log("Get Application with system number 1")
	req = fixtures.FixGetApplicationBySystemNumberRequest("1")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &appResp1)
	require.NoError(t, err)

	t.Log("Get Application with system number 2")
	req = fixtures.FixGetApplicationBySystemNumberRequest("2")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), req, &appResp2)
	require.NoError(t, err)

	require.Equal(t, connectedStatus.String(), appResp1.Status.Condition.String())
	require.Equal(t, connectedStatus.String(), appResp2.Status.Condition.String())
}

func triggerSync(t *testing.T, tenantID string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	systemFetcherClient := &http.Client{
		Transport: httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), util.AuthorizationHeader),
		Timeout:   time.Duration(1) * time.Minute,
	}

	jsonBody := fmt.Sprintf(`{"tenantID":"%s"}`, tenantID)
	sfReq, err := http.NewRequest(http.MethodPost, cfg.SystemFetcherURL+"/sync", bytes.NewBuffer([]byte(jsonBody)))
	require.NoError(t, err)
	sfReq.Header.Add(tenantHeader, tenantID)
	sfResp, err := systemFetcherClient.Do(sfReq)
	defer func() {
		if err := sfResp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, sfResp.StatusCode)
}

func waitForApplicationsToBeProcessed(ctx context.Context, t *testing.T, tenantID string, expectedNumber int) {
	require.Eventually(t, func() bool {
		_, actualApps := retrieveAppsForTenant(t, ctx, tenantID)
		t.Logf("Found %d from %d", len(actualApps), expectedNumber)
		return len(actualApps) >= expectedNumber
	}, time.Minute*1, time.Second*1, "Waiting for Systems to be fetched.")
}

func waitForDeleteOperation(ctx context.Context, t *testing.T, appID string) {
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)

	k8sClient, err := client.NewForConfig(cfg)
	require.NoError(t, err)
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
		bodyBytes, err := io.ReadAll(response.Body)
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
			Labels: applicationLabels(systemName, "", "", false, ""),
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
		bodyBytes, err := io.ReadAll(response.Body)
		require.NoError(t, err)
		t.Fatalf("Failed to reset mock systems: %s", string(bodyBytes))
		return
	}
	log.D().Info("Successfully reset mock systems")
}

func applicationLabels(name, appTemplateName, integrationSystemID string, fromTemplate bool, regionLabel string) directorSchema.Labels {
	labels := directorSchema.Labels{
		"managed":                "true",
		"name":                   fmt.Sprintf("mp-%s", name),
		"ppmsProductVersionId":   "12345",
		"productId":              "XXX",
		"integrationSystemID":    integrationSystemID,
		"tenantBusinessTypeCode": "tbtID",
		"tenantBusinessTypeName": "tbt description name",
	}

	if fromTemplate {
		labels[nameLabelKey] = name
		labels["applicationType"] = appTemplateName
	}

	if len(regionLabel) > 0 {
		labels[regionLabelKey] = regionLabel
	}

	return labels
}

func fixApplicationTemplate(name, intSystemID string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        name,
		Description: str.Ptr("template description"),
		ApplicationInput: &directorSchema.ApplicationJSONInput{
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
	}

	return appTemplateInput
}

func fixRegionalApplicationTemplate(name, intSystemID, region string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        name,
		Description: str.Ptr("template description"),
		ApplicationInput: &directorSchema.ApplicationJSONInput{
			Name:        fmt.Sprintf("{{%s}}", namePlaceholder),
			Description: ptr.String(fmt.Sprintf("{{%s}}", displayNamePlaceholder)),
			Labels: directorSchema.Labels{
				nameLabelKey:   "{{name}}",
				regionLabelKey: "{{region}}",
			},
			Webhooks: []*directorSchema.WebhookInput{{
				Type: directorSchema.WebhookTypeConfigurationChanged,
				URL:  ptr.String("http://url.com"),
			}},
			HealthCheckURL:      ptr.String("http://url.valid"),
			IntegrationSystemID: &intSystemID,
		},
		Labels: directorSchema.Labels{
			regionLabelKey: region,
		},
		Placeholders: []*directorSchema.PlaceholderDefinitionInput{
			{
				Name: namePlaceholder,
			},
			{
				Name: displayNamePlaceholder,
			},
			{
				Name:     regionLabelKey,
				JSONPath: str.Ptr("$.additionalAttributes.systemSCPLandscapeID"),
			},
		},
		AccessLevel: directorSchema.ApplicationTemplateAccessLevelGlobal,
	}

	return appTemplateInput
}

func fixApplicationTemplateWithDefaultSystemRoles(name, intSystemID string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := fixApplicationTemplate(name, intSystemID)

	appTemplateInput.Labels = map[string]interface{}{
		cfg.TemplateLabelFilter: []interface{}{"val1"},
	}

	return appTemplateInput
}

func fixApplicationTemplateWithSystemRoles(name, intSystemID string, systemRoles []interface{}) directorSchema.ApplicationTemplateInput {
	appTemplateInput := fixApplicationTemplate(name, intSystemID)

	appTemplateInput.Labels = map[string]interface{}{
		cfg.TemplateLabelFilter: systemRoles,
	}

	return appTemplateInput
}

func fixRegionalApplicationTemplateWithSystemRoles(name, intSystemID string, systemRoles []interface{}, region string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := fixRegionalApplicationTemplate(name, intSystemID, region)
	appTemplateInput.Labels[cfg.TemplateLabelFilter] = systemRoles

	return appTemplateInput
}

func fixApplicationTemplateWithoutWebhooks(name, intSystemID string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        name,
		Description: str.Ptr("template description"),
		ApplicationInput: &directorSchema.ApplicationJSONInput{
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
	}

	return appTemplateInput
}

func fixApplicationTemplateWithoutWebhooksWithSystemRole(name, intSystemID string, systemRoles []string) directorSchema.ApplicationTemplateInput {
	appTemplateInput := fixApplicationTemplateWithoutWebhooks(name, intSystemID)

	appTemplateInput.Labels = map[string]interface{}{
		cfg.TemplateLabelFilter: systemRoles,
	}

	return appTemplateInput
}

func cleanupTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantExternalID string) {
	tenantsToDelete := []directorSchema.BusinessTenantMappingInput{
		{
			ExternalTenant: tenantExternalID,
		},
	}
	err := fixtures.DeleteTenants(t, ctx, gqlClient, tenantsToDelete)
	assert.NoError(t, err)
	log.D().Info("Successfully cleanup tenants")
}
