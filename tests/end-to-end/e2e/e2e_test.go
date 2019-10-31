package e2e

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/common"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/idtokenprovider"
	"github.com/stretchr/testify/require"
)

func init() {
	var err error
	tc, err = NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

func TestCompassAuth(t *testing.T) {
	ctx := context.Background()
	t.Log("Get Dex id token")
	config, err := idtokenprovider.LoadConfig()
	require.NoError(t, err)

	dexToken, err := idtokenprovider.Authenticate(config.IdProviderConfig)
	require.NoError(t, err)

	tc.cli = common.NewAuthorizedGraphQLClient(dexToken)
	t.Log("tokkenn", dexToken)

	t.Log("Create Integration System with Dex id token")

	//intSys :=
	createIntegrationSystem(t, ctx, "int-sys")

	//t.Log("Generate Client Credentials for Integration System")
	//intSysAuth := generateClientCredentialsForIntegrationSystem(t, ctx, intSys.ID)
	//intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(graphql.OAuthCredentialData)
	//require.True(t, ok)

	//t.Log("Issue a Hydra dexToken with Client Credentials")
	//oauthCredentials := intSysOauthCredentialData.ClientID + ":" + intSysOauthCredentialData.ClientSecret
	//encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	//
	//form := url.Values{
	//	"grant_type": {"cllient_credentials"},
	//	"scope":      {},
	//}
	//req, err := http.NewRequest("POST", "url", strings.NewReader(form.Encode()))
	//require.NoError(t, err)
	//req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))
	////resp, err := http.PostForm("", url.Values{
	////})
	//
	//resp, err := client.Do(req)
	//require.NoError(t, err)
	//defer resp.Body.Close()
	//
	//require.Equal(t, http.StatusOK, resp.StatusCode)
	//intSysToken, err := ioutil.ReadAll(resp.Body)
	//
	//t.Log("Create an application as Integration System")
	//tc.cli = common.NewAuthorizedGraphQLClient(string(intSysToken))
	//appByIntSys := createApplication(t, ctx, "app-created-by-integration-system")
	//
	//t.Log("Add API Spec to Application")
	//apiInput := graphql.APIDefinitionInput{
	//	Name:      "new-api-name",
	//	TargetURL: "new-api-url",
	//	Spec: &graphql.APISpecInput{
	//		Data:         nil,
	//		Type:         "",
	//		Format:       "",
	//		FetchRequest: nil,
	//	},
	//}
	//require.NoError(t, err)
	//
	//addApi(t, ctx, apiInput, appByIntSys.ID)
	//
	//t.Log("Remove application using Dex id token")
	//tc.cli = common.NewAuthorizedGraphQLClient(dexToken)
	//
	//deleteApplication(t, ctx, appByIntSys.ID)
	//
	//// Remove this step when https://github.com/kyma-incubator/compass/issues/375 is done
	//t.Log("Remove System Auth with client credentials")
	//deleteSystemAuthForIntegrationSystem(t, ctx, intSysAuth.ID)
	//
	//t.Log("Remove Integration System")
	//deleteIntegrationSystem(t, ctx, intSys.ID)
	//
	//t.Log("Check if token granted for Integration System is invalid")
	//appInput := graphql.ApplicationInput{
	//	Name: "app-which-should-be-not-created",
	//}
	//appInputGQL, err := tc.Graphqlizer.ApplicationInputToGQL(appInput)
	//require.NoError(t, err)
	//createRequest := fixCreateApplicationRequest(appInputGQL)
	//require.Error(t, tc.RunOperation(ctx, createRequest, nil))
	//
	//t.Log("Check if token can not be fetched with old client credentials")
	//
	//resp, err = client.Do(req)
	//require.NoError(t, err)
	//defer resp.Body.Close()
	//
	//require.Equal(t, http.StatusForbidden, resp.StatusCode)
}
