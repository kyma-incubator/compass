package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/tests/pkg"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	graphql2 "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
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
	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	registerRequest := pkg.FixRegisterApplicationRequest(appInputGQL)

	t.Log("Register Application through Gateway with Dex id Token")
	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTenant, registerRequest, &app)
	require.NoError(t, err)

	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)

	t.Log("Get auditlog service Token")
	auditlogToken := pkg.GetAuditlogMockToken(t, &httpClient, testConfig.ExternalServicesMockBaseURL)

	t.Log("Get auditlog from external services mock")
	auditlogs := pkg.SearchForAuditlogByString(t, &httpClient, testConfig.ExternalServicesMockBaseURL, auditlogToken, appName)

	for _, auditlog := range auditlogs {
		defer pkg.DeleteAuditlogByID(t, &httpClient, testConfig.ExternalServicesMockBaseURL, auditlogToken, auditlog.UUID)
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
