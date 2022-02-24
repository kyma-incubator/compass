package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"

	graphql2 "github.com/machinebox/graphql"

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

	t.Log("Create request for registering application")
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	registerRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)

	t.Log("Register Application through Gateway with Dex id Token")
	app := graphql.ApplicationExt{}

	timeFrom := time.Now()
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
	require.NoError(t, err)
	timeTo := timeFrom.Add(1 * time.Minute)

	t.Log("Get auditlog service Token")
	auditlogToken := token.GetClientCredentialsToken(t, context.Background(), testConfig.Auditlog.TokenURL+"/oauth/token", testConfig.Auditlog.ClientID, testConfig.Auditlog.ClientSecret, "")

	t.Log("Get auditlog from auditlog API")
	auditlogs := fixtures.SearchForAuditlogByTimestampAndString(t, &httpClient, testConfig.Auditlog, auditlogToken, appName, timeFrom, timeTo)

	assert.Eventually(t, func() bool {
		auditlogs = fixtures.SearchForAuditlogByTimestampAndString(t, &httpClient, testConfig.Auditlog, auditlogToken, appName, timeFrom, timeTo)
		t.Logf("Waiting for auditlog items to be %d, but currently are: %d", 2, len(auditlogs))
		return len(auditlogs) == 2
	}, time.Minute, time.Millisecond*500)

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
	assert.Equal(t, consumerID, pre.Object.ID["consumerID"])
	assert.Equal(t, "Static User", pre.Object.ID["apiConsumer"])

	var postRequest string
	for _, v := range post.Attributes {
		if v.Name == "request" {
			postRequest = v.New
		}
	}

	assert.Equal(t, requestBody.String(), postRequest)
	assert.Equal(t, consumerID, post.Object.ID["consumerID"])
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
