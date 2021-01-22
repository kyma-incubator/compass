package external_services_mock_integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	graphql2 "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
)

func TestAuditlogIntegration(t *testing.T) {
	ctx := context.Background()
	httpClient := http.Client{}
	appName := "app-for-testing-auditlog-mock"
	appInput := graphql.ApplicationRegisterInput{
		Name:         appName,
		ProviderName: ptr.String("compass"),
	}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Create request for registering application")
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	registerRequest := fixRegisterApplicationRequest(appInputGQL)

	t.Log("Register Application through Gateway with Dex id token")
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTenant, registerRequest, &app)
	require.NoError(t, err)

	defer unregisterApplicationInTenant(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	t.Log("Get auditlog service token")
	auditlogToken := getAuditlogMockToken(t, &httpClient, testConfig.ExternalServicesMockBaseURL)

	t.Log("Get auditlog from external services mock")
	auditlogs := searchForAuditlogByString(t, &httpClient, testConfig.ExternalServicesMockBaseURL, auditlogToken, appName)

	for _, auditlog := range auditlogs {
		defer deleteAuditlogByID(t, &httpClient, testConfig.ExternalServicesMockBaseURL, auditlogToken, auditlog.UUID)
	}

	t.Log("Compare request to director with auditlog")
	requestBody := prepareRegisterAppRequestBody(t, registerRequest)
	require.True(t, len(auditlogs[0].Attributes) == 3 || len(auditlogs[0].Attributes) == 4)
	var pre, post model.ConfigurationChange

	for _, v := range auditlogs {
		for _, attr := range v.Attributes {
			if attr.Name == "auditlog_type" && attr.New == "pre-operation" {
				pre = v
			}
			if attr.Name == "auditlog_type" && attr.New == "post-operation" {
				post = v
			}
		}
	}

	var preRequest string
	for _, v := range pre.Attributes {
		if v.Name == "request" {
			preRequest = v.New
		}
	}

	assert.Equal(t, requestBody.String(), preRequest)
	assert.Equal(t, 2, len(auditlogs))
	assert.Equal(t, "admin", pre.Object.ID["consumerID"])
	assert.Equal(t, "Static User", pre.Object.ID["apiConsumer"])

	var postRequest string
	for _, v := range post.Attributes {
		if v.Name == "request" {
			postRequest = v.New
		}
	}

	assert.Equal(t, requestBody.String(), postRequest)
	assert.Equal(t, "admin", post.Object.ID["consumerID"])
	assert.Equal(t, "Static User", post.Object.ID["apiConsumer"])
}

func prepareRegisterAppRequestBody(t *testing.T, registerRequest *graphql2.Request) bytes.Buffer {
	var requestBody bytes.Buffer
	requestBodyObj := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     registerRequest.Query(),
		Variables: registerRequest.Vars(),
	}
	err := json.NewEncoder(&requestBody).Encode(requestBodyObj)
	require.NoError(t, err)

	return requestBody
}
