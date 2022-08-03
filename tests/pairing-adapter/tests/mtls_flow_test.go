package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"

	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/pairing"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/require"
)

func TestGettingTokenWithMTLSWorks(t *testing.T) {
	ctx := context.Background()
	defaultTestTenant := tenant.TestTenants.GetDefaultTenantID()
	templateName := conf.TemplateName
	if !strings.HasPrefix(templateName, "SAP ") {
		templateName = fmt.Sprintf("SAP %s", templateName)
	}
	namePlaceholderKey := "name"
	displayNamePlaceholderKey := "display-name"
	appTemplate := &directorSchema.ApplicationTemplate{}
	newIntSys := &directorSchema.IntegrationSystemExt{}

	if conf.IsLocalEnv {
		newIntSys = createIntSystem(t, ctx, defaultTestTenant)
		defer func() {
			fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, newIntSys)
		}()

		updateAdaptersConfigmap(t, ctx, newIntSys.ID, conf)

		appTemplate = createAppTemplate(t, ctx, defaultTestTenant, newIntSys.ID, templateName, namePlaceholderKey, displayNamePlaceholderKey)
		defer func() {
			fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, defaultTestTenant, appTemplate)
		}()
	}

	appTmplInput := directorSchema.ApplicationFromTemplateInput{
		TemplateName: templateName,
		Values: []*directorSchema.TemplateValueInput{
			{
				Placeholder: namePlaceholderKey,
				Value:       "E2E pairing adapter test app",
			},
			{
				Placeholder: displayNamePlaceholderKey,
				Value:       "E2E pairing adapter test app Display Name",
			},
		},
	}

	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appTmplInput)
	require.NoError(t, err)

	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := directorSchema.ApplicationExt{}

	t.Logf("Registering application from application template with name: %q...", templateName)

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, defaultTestTenant, createAppFromTmplRequest, &outputApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, defaultTestTenant, &outputApp)
	require.NoError(t, err)
	require.NotEmpty(t, outputApp.ID)
	t.Logf("Successfully registered application from application template with name: %q", templateName)

	t.Logf("Getting one time token for application with name: %q and id: %q...", outputApp.Name, outputApp.ID)
	tokenRequest := fixtures.FixRequestOneTimeTokenForApplication(outputApp.ID)
	tokenRequest.Header.Add(conf.ClientIDHeader, "i507827") // needed for the productive test execution
	token := directorSchema.OneTimeTokenForApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, tokenRequest, &token)
	require.NoError(t, err)
	require.NotEmpty(t, token.Token)
	require.Empty(t, token.ConnectorURL)
	require.NotEmpty(t, token.LegacyConnectorURL)
	t.Logf("Successfully got one time token for application with name: %q and id: %q", outputApp.Name, outputApp.ID)

	// only for local setup we need to revert the changes made by the test with the initial/default values so the next execution to be successful
	if conf.IsLocalEnv {
		updateAdaptersConfigmapWithDefaultValues(t, ctx, conf)
	}
}

func TestGettingTokenWithMTLSThroughFQN(t *testing.T) {
	reqData := pairing.RequestData{
		Application: directorSchema.Application{
			Name: conf.TestApplicationName,
			BaseEntity: &directorSchema.BaseEntity{
				ID: conf.TestApplicationID,
			},
		},
		Tenant:     conf.TestTenant,
		ClientUser: conf.TestClientUser,
	}
	jsonReqData, err := json.Marshal(reqData)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, conf.FQDNPairingAdapterURL, strings.NewReader(string(jsonReqData)))
	require.NoError(t, err)

	client := http.Client{
		Transport: director_http.NewServiceAccountTokenTransport(director_http.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respParsed := struct {
		Token string
	}{}

	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
	require.NotEmpty(t, respParsed.Token)
}

func createIntSystem(t *testing.T, ctx context.Context, defaultTestTenant string) *directorSchema.IntegrationSystemExt {
	// GIVEN
	name := "pairing-adapter-int-system"

	// WHEN
	t.Logf("Registering integration system with name %q...", name)
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, name)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)
	require.NotEmpty(t, intSys.Name)

	t.Logf("Successfully registered integration system with name %q", name)
	return intSys
}

func updateAdaptersConfigmap(t *testing.T, ctx context.Context, newIntSysID string, adapterConfig *config.PairingAdapterConfig) {
	t.Logf("Updating adapters configmap with newly created integration system ID: %q...", newIntSysID)
	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	cmManager := k8sClient.CoreV1().ConfigMaps(adapterConfig.ConfigMapNamespace)
	cm, err := cmManager.Get(ctx, adapterConfig.ConfigMapName, metav1.GetOptions{})
	require.NoError(t, err)

	adaptersMap := make(map[string]string)
	err = json.Unmarshal([]byte(cm.Data[adapterConfig.ConfigMapKey]), &adaptersMap)
	require.NoError(t, err)

	localAdapterFQDN := adapterConfig.LocalAdapterFQDN
	adapterURL, found := adaptersMap[adapterConfig.IntegrationSystemID]
	require.True(t, found)
	require.Equal(t, localAdapterFQDN, adapterURL)

	require.NotEmpty(t, newIntSysID)
	cmJsonDataUpdated := fmt.Sprintf("{\"%s\":\"%s\"}", newIntSysID, localAdapterFQDN)
	updatedMap := make(map[string]string)
	updatedMap[adapterConfig.ConfigMapKey] = cmJsonDataUpdated
	cm.Data = updatedMap
	cm, err = cmManager.Update(ctx, cm, metav1.UpdateOptions{})
	require.NoError(t, err)
	t.Log("Successfully updated adapters configmap")
}

func updateAdaptersConfigmapWithDefaultValues(t *testing.T, ctx context.Context, adapterConfig *config.PairingAdapterConfig) {
	t.Log("Updating adapters configmap with with default values")
	k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	require.NoError(t, err)

	cmManager := k8sClient.CoreV1().ConfigMaps(adapterConfig.ConfigMapNamespace)
	cm, err := cmManager.Get(ctx, adapterConfig.ConfigMapName, metav1.GetOptions{})
	require.NoError(t, err)

	cmJsonDataUpdated := fmt.Sprintf("{\"%s\":\"%s\"}", conf.IntegrationSystemID, adapterConfig.LocalAdapterFQDN)
	updatedMap := make(map[string]string)
	updatedMap[adapterConfig.ConfigMapKey] = cmJsonDataUpdated
	cm.Data = updatedMap
	cm, err = cmManager.Update(ctx, cm, metav1.UpdateOptions{})
	require.NoError(t, err)
	t.Log("Successfully updated adapters configmap with default values")
}

func createAppTemplate(t *testing.T, ctx context.Context, defaultTestTenant, newIntSysID, templateName, namePlaceholderKey, displayNamePlaceholderKey string) *directorSchema.ApplicationTemplate {
	appTemplateDesc := "pairing-adapter-app-template-desc"
	providerName := "compass-e2e-tests"
	namePlaceholderDescription := "name-description"
	displayNamePlaceholderDescription := "display-name-description"
	integrationSystemID := newIntSysID

	appTemplateInput := directorSchema.ApplicationTemplateInput{
		Name:        templateName,
		Description: &appTemplateDesc,
		ApplicationInput: &directorSchema.ApplicationRegisterInput{
			Name:                fmt.Sprintf("{{%s}}", namePlaceholderKey),
			ProviderName:        &providerName,
			Description:         ptr.String(fmt.Sprintf("test {{%s}}", displayNamePlaceholderKey)),
			IntegrationSystemID: &integrationSystemID,
		},
		Placeholders: []*directorSchema.PlaceholderDefinitionInput{
			{
				Name:        namePlaceholderKey,
				Description: &namePlaceholderDescription,
			},
			{
				Name:        displayNamePlaceholderKey,
				Description: &displayNamePlaceholderDescription,
			},
		},
		Labels: directorSchema.Labels{
			conf.SelfRegDistinguishLabelKey: []interface{}{conf.SelfRegDistinguishLabelValue},
		},
		AccessLevel: directorSchema.ApplicationTemplateAccessLevelGlobal,
	}

	t.Logf("Registering application template with name %q...", templateName)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, defaultTestTenant, appTemplateInput)
	require.NoError(t, err)
	require.NotEmpty(t, appTmpl.ID)
	require.NotEmpty(t, appTmpl.Name)

	t.Log("Check if application template was created...")

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(appTmpl.ID)
	appTemplateOutput := directorSchema.ApplicationTemplate{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOutput)

	appTemplateInput.Labels[conf.SelfRegLabelKey] = appTemplateOutput.Labels[conf.SelfRegLabelKey]
	appTemplateInput.Labels[tenantfetcher.RegionKey] = appTemplateOutput.Labels[tenantfetcher.RegionKey]
	appTemplateInput.Labels["global_subaccount_id"] = tenant.TestTenants.GetIDByName(t, tenant.TestProviderSubaccount)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": fmt.Sprintf("%s (%s)", templateName, conf.SelfRegRegion)}
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)

	t.Logf("Successfully registered application template with name %q", templateName)

	return &appTmpl
}
