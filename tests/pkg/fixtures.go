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

package pkg

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gateway_integration "github.com/kyma-incubator/compass/tests/director/gateway-integration"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

var tc *gateway_integration.TestContext

func init() {
	var err error
	tc, err = gateway_integration.NewTestContext()
	if err != nil {
		panic(errors.Wrap(err, "while test context setup"))
	}
}

func RegisterApplicationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
	return app, err
}

func UnregisterApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant string, applicationID string) graphql.ApplicationExt {
	deleteRequest := fixDeleteApplicationRequest(t, applicationID)
	app := graphql.ApplicationExt{}

	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &app)
	require.NoError(t, err)
	return app
}

func RegisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}
	req := fixRegisterIntegrationSystemRequest(in)
	out := &graphql.IntegrationSystemExt{}
	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, "", req, out)
	require.NotEmpty(t, out)
	require.NoError(t, err)
	return out
}

func UnregisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := fixunregisterIntegrationSystem(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, "", req, nil)
	require.NoError(t, err)
}

func RequestClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.SystemAuth {
	generateIntSysAuthRequest := fixRequestClientCredentialsForIntegrationSystem(id)
	intSysAuth := &graphql.SystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, "", generateIntSysAuthRequest, intSysAuth)
	require.NoError(t, err)
	require.NotEmpty(t, intSysAuth.Auth)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, intSysAuth.Auth.Credential)
	oauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	assert.NotEmpty(t, oauthCredentialData.ClientID)
	assert.NotEmpty(t, oauthCredentialData.ClientSecret)
	assert.Equal(t, intSysAuth.ID, oauthCredentialData.ClientID)
	return intSysAuth
}

func DeleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := fixDeleteSystemAuthForIntegrationSystemRequest(id)
	err := tc.RunOperationWithCustomTenant(ctx, gqlClient, "", req, nil)
	require.NoError(t, err)
}

func fixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.GQLFieldsProvider.ForApplication()))
}

func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, tc.GQLFieldsProvider.ForApplication()))
}

func fixRegisterIntegrationSystemRequest(integrationSystemInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerIntegrationSystem(in: %s) {
					%s
				}
			}`,
			integrationSystemInGQL, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixunregisterIntegrationSystem(intSysID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: unregisterIntegrationSystem(id: "%s") {
					%s
				}
			}`, intSysID, tc.GQLFieldsProvider.ForIntegrationSystem()))
}

func fixRequestClientCredentialsForIntegrationSystem(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: requestClientCredentialsForIntegrationSystem(id: "%s") {
						%s
					}
				}`, id, tc.GQLFieldsProvider.ForSystemAuth()))
}

func fixDeleteSystemAuthForIntegrationSystemRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForIntegrationSystem(authID: "%s") {
					%s
				}
			}`, authID, tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixActiveVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v2",
		Deprecated: ptr.Bool(false),
		ForRemoval: ptr.Bool(false),
	}
}

func FixDecomissionedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:      "v1",
		Deprecated: ptr.Bool(true),
		ForRemoval: ptr.Bool(true),
	}
}

func FixDepracatedVersion() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptr.Bool(true),
		ForRemoval:      ptr.Bool(false),
		DeprecatedSince: ptr.String("v5"),
	}
}
