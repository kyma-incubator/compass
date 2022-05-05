package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

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
	namePlaceholderKey := "name"
	displayNamePlaceholderKey := "display-name"

	if conf.IsLocalEnv {
		newIntSysID := createIntSystem(t, ctx, defaultTestTenant)
		updateAdaptersConfigmap(t, ctx, newIntSysID, conf)
		createAppTemplate(t, ctx, defaultTestTenant, newIntSysID, templateName, namePlaceholderKey, displayNamePlaceholderKey)
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
	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, defaultTestTenant, createAppFromTmplRequest, &outputApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, defaultTestTenant, &outputApp)
	require.NoError(t, err)
	require.NotEmpty(t, outputApp.ID)

	token := fixtures.RequestOneTimeTokenForApplication(t, ctx, certSecuredGraphQLClient, outputApp.ID)
	require.NotEmpty(t, token.Token)
	require.Empty(t, token.ConnectorURL)
	require.Empty(t, token.LegacyConnectorURL)
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
		Transport: director_http.NewServiceAccountTokenTransport(http.DefaultTransport),
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

func createIntSystem(t *testing.T, ctx context.Context, defaultTestTenant string) string {
	// GIVEN
	name := "pairing-adapter-int-system"

	// WHEN
	t.Logf("Registering integration system with name %s...", name)
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, defaultTestTenant, intSys)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)
	require.NotEmpty(t, intSys.Name)

	t.Logf("Successfully registered integration system with name %s", name)
	return intSys.ID
}

func updateAdaptersConfigmap(t *testing.T, ctx context.Context, newIntSysID string, adapterConfig *config.PairingAdapterConfig) {
	t.Logf("Updating adapters config map with newly created integartion system ID: %s...", newIntSysID)
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

	cmJsonDataUpdated := fmt.Sprintf("{\"%s\":\"%s\"}", newIntSysID, localAdapterFQDN)

	updatedMap := make(map[string]string)
	updatedMap[adapterConfig.ConfigMapKey] = cmJsonDataUpdated
	cm.Data = updatedMap
	cm, err = cmManager.Update(ctx, cm, metav1.UpdateOptions{})
	require.NoError(t, err)
	t.Log("Successfully updated adapters config map")
}

func createAppTemplate(t *testing.T, ctx context.Context, defaultTestTenant, newIntSysID, templateName, namePlaceholderKey, displayNamePlaceholderKey string) {
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
		AccessLevel: directorSchema.ApplicationTemplateAccessLevelGlobal,
	}

	t.Logf("Registering application template with name %s...", templateName)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, defaultTestTenant, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, defaultTestTenant, &appTmpl)
	require.NoError(t, err)
	t.Logf("Successfully registered application template with name %s", templateName)
}
