package fixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func FixRuntimeRegisterInput(placeholder string) graphql.RuntimeRegisterInput {
	return graphql.RuntimeRegisterInput{
		Name:        placeholder,
		Description: ptr.String(fmt.Sprintf("%s-description", placeholder)),
		Labels:      graphql.Labels{"placeholder": []interface{}{"placeholder"}},
	}
}

func FixRuntimeRegisterInputWithWebhooks(placeholder string) graphql.RuntimeRegisterInput {
	return graphql.RuntimeRegisterInput{
		Name:        placeholder,
		Description: ptr.String(fmt.Sprintf("%s-description", placeholder)),
		Labels:      graphql.Labels{"placeholder": []interface{}{"placeholder"}},
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.WebhookTypeConfigurationChanged,
			URL:  ptr.String(webhookURL),
		}},
	}
}

func FixRuntimeUpdateInput(placeholder string) graphql.RuntimeUpdateInput {
	return graphql.RuntimeUpdateInput{
		Name:        placeholder,
		Description: ptr.String(fmt.Sprintf("%s-description", placeholder)),
		Labels:      graphql.Labels{"placeholder": []interface{}{"placeholder"}},
	}
}

func FixRegisterRuntimeRequest(runtimeInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerRuntime(in: %s) {
					%s
				}
			}`,
			runtimeInGQL, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRequestClientCredentialsForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestClientCredentialsForRuntime(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}

func FixUnregisterRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{unregisterRuntime(id: "%s") {
				%s
			}
		}`, id, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixDeleteRuntimeLabel(id, key string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation{deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
				%s
			}
		}`, id, key, testctx.Tc.GQLFieldsProvider.ForLabel()))
}

func FixGetRuntimeRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}}`, id, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRequestOneTimeTokenForRuntime(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestOneTimeTokenForRuntime(id: "%s") {
					%s
				}
			}`, id, testctx.Tc.GQLFieldsProvider.ForOneTimeTokenForRuntime()))
}

func FixUpdateRuntimeRequest(id, updateInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`,
			id, updateInputInGQL, testctx.Tc.GQLFieldsProvider.ForRuntime()))
}

func FixRuntimeRequestWithPaginationRequest(after int, cursor string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes(first:%d, after:"%s") {
					%s
				}
			}`, after, cursor, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixGetRuntimesRequestWithPagination() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes {
						%s
					}
				}`,
			testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixRuntimesFilteredPageableRequest(labelFilterInGQL string, first int, after string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: runtimes(filter: %s, first: %d, after: "%s") {
						%s
					}
				}`,
			labelFilterInGQL, first, after, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForRuntime())))
}

func FixDeleteSystemAuthForRuntimeRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteSystemAuthForRuntime(authID: "%s") {
					%s
				}
			}`, authID, testctx.Tc.GQLFieldsProvider.ForSystemAuth()))
}
func RegisterKymaRuntime(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID string, runtimeInput graphql.RuntimeRegisterInput, oauthPath string) graphql.RuntimeExt {
	intSysName := "runtime-integration-system"

	t.Logf("Creating integration system with name: %q", intSysName)
	intSys, err := RegisterIntegrationSystem(t, ctx, gqlClient, tenantID, intSysName)
	defer CleanupIntegrationSystem(t, ctx, gqlClient, tenantID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := RequestClientCredentialsForIntegrationSystem(t, ctx, gqlClient, tenantID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer DeleteSystemAuthForIntegrationSystem(t, ctx, gqlClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, oauthPath)

	t.Logf("Registering runtime with name %q with integration system credentials...", runtimeInput.Name)
	kymaRuntime, err := RegisterRuntimeFromInputWithinTenant(t, ctx, oauthGraphQLClient, tenantID, &runtimeInput)
	require.NoError(t, err)
	require.NotEmpty(t, kymaRuntime.ID)

	return kymaRuntime
}
