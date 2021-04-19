package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.IntegrationSystemExt {
	intSysRequest := FixGetIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, intSysRequest, &intSys))
	return &intSys
}

func RegisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, name string) *graphql.IntegrationSystemExt {
	input := graphql.IntegrationSystemInput{Name: name}
	in, err := testctx.Tc.Graphqlizer.IntegrationSystemInputToGQL(input)
	if err != nil {
		return nil
	}

	req := FixRegisterIntegrationSystemRequest(in)
	intSys := &graphql.IntegrationSystemExt{}

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys)
	return intSys
}

func UnregisterIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	if id == "" {
		return
	}
	req := FixUnregisterIntegrationSystem(id)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.NoError(t, err)
}

func UnregisterIntegrationSystemWithErr(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixUnregisterIntegrationSystem(id)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "The record cannot be deleted because another record refers to it")
}

func GetSystemAuthsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) []*graphql.IntSysSystemAuth {
	req := FixGetIntegrationSystemRequest(id)
	intSys := graphql.IntegrationSystemExt{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &intSys)
	require.NoError(t, err)
	return intSys.Auths
}

func RequestClientCredentialsForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) *graphql.IntSysSystemAuth {
	req := FixRequestClientCredentialsForIntegrationSystem(id)
	systemAuth := graphql.IntSysSystemAuth{}

	// WHEN
	t.Log("Generate client credentials for integration system")
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	require.NotEmpty(t, systemAuth.Auth)

	t.Log("Check if client credentials were generated")
	assert.NotEmpty(t, systemAuth.Auth.Credential)
	intSysOauthCredentialData, ok := systemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, intSysOauthCredentialData.ClientSecret)
	require.NotEmpty(t, intSysOauthCredentialData.ClientID)
	assert.Equal(t, systemAuth.ID, intSysOauthCredentialData.ClientID)
	return &systemAuth
}

func DeleteSystemAuthForIntegrationSystem(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForIntegrationSystemRequest(id)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, "", req, nil)
	require.NoError(t, err)
}
