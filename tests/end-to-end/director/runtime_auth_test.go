package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryRuntimeAuths(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	defaultAuth := fixBasicAuth()
	customAuth := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "custom",
				Password: "auth",
			}},
		AdditionalHeaders: &graphql.HttpHeaders{
			"customHeader": []string{"custom", "custom"},
		},
	}

	exampleSaved := false
	rtmsToCreate := 3

	testCases := []struct {
		Name string
		Apis []*graphql.APIDefinitionInput
	}{
		{
			Name: "API without default auth",
			Apis: []*graphql.APIDefinitionInput{
				{
					Name:      "without-default-auth",
					TargetURL: "http://mywordpress.com/comments",
				},
			},
		},
		{
			Name: "API with default auth",
			Apis: []*graphql.APIDefinitionInput{
				{
					Name:        "with-default-auth",
					TargetURL:   "http://mywordpress.com/comments",
					DefaultAuth: defaultAuth,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appInput := graphql.ApplicationInput{
				Name: "test-app",
				Apis: testCase.Apis,
				Labels: &graphql.Labels{
					"scenarios": []interface{}{"DEFAULT"},
				},
			}

			app := createApplicationFromInputWithinTenant(t, ctx, appInput, defaultTenant)
			defer deleteApplication(t, app.ID)

			var rtmIDs []string
			for i := 0; i < rtmsToCreate; i++ {
				rtm := createRuntime(t, ctx, fmt.Sprintf("test-rtm-%d", i))
				rtmIDs = append(rtmIDs, rtm.ID)
				defer deleteRuntime(t, rtm.ID)
			}
			require.Len(t, rtmIDs, rtmsToCreate)

			rtmIDWithUnsetRtmAuth := rtmIDs[0]
			rtmIDWithSetRtmAuth := rtmIDs[1]

			setAPIAuth(t, ctx, app.Apis.Data[0].ID, rtmIDWithSetRtmAuth, customAuth)
			defer deleteAPIAuth(t, ctx, app.Apis.Data[0].ID, rtmIDWithSetRtmAuth)

			innerTestCases := []struct {
				Name         string
				QueriedRtmID string
			}{
				{
					Name:         "Query set Runtime Auth",
					QueriedRtmID: rtmIDWithSetRtmAuth,
				},
				{
					Name:         "Query unset Runtime Auth",
					QueriedRtmID: rtmIDWithUnsetRtmAuth,
				},
			}

			for _, innerTestCase := range innerTestCases {
				t.Run(innerTestCase.Name, func(t *testing.T) {
					result := graphql.ApplicationExt{}
					request := fixRuntimeAuthRequest(app.ID, innerTestCase.QueriedRtmID)

					// WHEN
					err := tc.RunQuery(ctx, request, &result)

					// THEN
					require.NoError(t, err)
					require.NotEmpty(t, result.ID)
					assertApplication(t, appInput, result)
					require.Len(t, result.Apis.Data, 1)

					assert.Equal(t, len(rtmIDs), len(result.Apis.Data[0].Auths))
					assert.Equal(t, innerTestCase.QueriedRtmID, result.Apis.Data[0].Auth.RuntimeID)
					if innerTestCase.QueriedRtmID == rtmIDWithSetRtmAuth {
						assertAuth(t, &customAuth, result.Apis.Data[0].Auth.Auth)
					} else {
						assert.Equal(t, result.Apis.Data[0].DefaultAuth, result.Apis.Data[0].Auth.Auth)
					}

					for _, auth := range result.Apis.Data[0].Auths {
						if auth.RuntimeID == rtmIDWithSetRtmAuth {
							assertAuth(t, &customAuth, auth.Auth)
						} else {
							assert.Equal(t, result.Apis.Data[0].DefaultAuth, auth.Auth)
						}
					}

					if !exampleSaved {
						saveQueryInExamples(t, request.Query(), "query runtime auths")
						exampleSaved = true
					}
				})
			}
		})
	}
}
