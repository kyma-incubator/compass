package fixtures

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateBundleInstanceAuthForRuntime(t *testing.T, ctx context.Context, oauthGraphQLClient *gcli.Client, tenantID, bundleID string) *graphql.BundleInstanceAuth {
	authCtx, inputParams := FixBundleInstanceAuthContextAndInputParams(t)
	bndlInstanceAuthRequestInput := FixBundleInstanceAuthRequestInput(authCtx, inputParams)
	bndlInstanceAuthRequestInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(bndlInstanceAuthRequestInput)
	require.NoError(t, err)

	bndlInstanceAuthCreationRequestReq := FixRequestBundleInstanceAuthCreationRequest(bundleID, bndlInstanceAuthRequestInputStr)
	output := graphql.BundleInstanceAuth{}

	t.Log("Request bundle instance auth creation")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantID, bndlInstanceAuthCreationRequestReq, &output)
	require.NoError(t, err)

	return &output
}

func SetBundleInstanceAuthForRuntime(t *testing.T, ctx context.Context, cli *gcli.Client, tenantID, biaID, clientID string) *graphql.BundleInstanceAuth {
	authInput := FixCertificateOauthAuthWithCustomCredentials(t, FixCertificateOAuthCredentialWithCustomClientID(clientID))
	bndlInstanceAuthSetInput := FixBundleInstanceAuthSetInputSucceeded(authInput)
	bndlInstanceAuthSetInputStr, err := testctx.Tc.Graphqlizer.BundleInstanceAuthSetInputToGQL(bndlInstanceAuthSetInput)
	require.NoError(t, err)

	setBundleInstanceAuthReq := FixSetBundleInstanceAuthRequest(biaID, bndlInstanceAuthSetInputStr)
	output := graphql.BundleInstanceAuth{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, cli, tenantID, setBundleInstanceAuthReq, &output)
	require.NoError(t, err)
	return &output
}

func CreateBundleWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, input graphql.BundleCreateInput) graphql.BundleExt {
	in, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(input)
	require.NoError(t, err)

	req := FixAddBundleRequest(appID, in)
	var resp graphql.BundleExt

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func CreateBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, appID, bndlName string) graphql.BundleExt {
	in, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(FixBundleCreateInput(bndlName))
	require.NoError(t, err)

	req := FixAddBundleRequest(appID, in)
	var resp graphql.BundleExt

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func DeleteBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := FixDeleteBundleRequest(id)

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func GetBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, appID, bundleID string) graphql.BundleExt {
	req := FixBundleRequest(appID, bundleID)
	bundle := graphql.ApplicationExt{}
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &bundle))
	return bundle.Bundle
}

func AddAPIToBundleWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, bndlID string, input graphql.APIDefinitionInput) graphql.APIDefinitionExt {
	inStr, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := FixAddAPIToBundleRequest(bndlID, inStr)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &actualApi)
	require.NoError(t, err)
	return actualApi
}

func AddAPIToBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.APIDefinitionExt {
	return AddAPIToBundleWithInput(t, ctx, gqlClient, tenant.TestTenants.GetDefaultTenantID(), bndlID, FixAPIDefinitionInput())
}

func AddEventToBundleWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string, input graphql.EventDefinitionInput) graphql.EventDefinition {
	inStr, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	event := graphql.EventDefinition{}
	req := FixAddEventAPIToBundleRequest(bndlID, inStr)
	err = testctx.Tc.RunOperation(ctx, gqlClient, req, &event)
	require.NoError(t, err)
	return event
}

func AddEventToBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.EventDefinition {
	return AddEventToBundleWithInput(t, ctx, gqlClient, bndlID, FixEventDefinitionInput())
}

func AddDocumentToBundleWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string, input graphql.DocumentInput) graphql.DocumentExt {
	inStr, err := testctx.Tc.Graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDoc := graphql.DocumentExt{}
	req := FixAddDocumentToBundleRequest(bndlID, inStr)
	err = testctx.Tc.RunOperation(ctx, gqlClient, req, &actualDoc)
	require.NoError(t, err)
	return actualDoc
}

func AddDocumentToBundle(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.DocumentExt {
	return AddDocumentToBundleWithInput(t, ctx, gqlClient, bndlID, FixDocumentInput(t))
}

func CreateBundleInstanceAuth(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, bndlID string) graphql.BundleInstanceAuth {
	authCtx, inputParams := FixBundleInstanceAuthContextAndInputParams(t)
	in, err := testctx.Tc.Graphqlizer.BundleInstanceAuthRequestInputToGQL(FixBundleInstanceAuthRequestInput(authCtx, inputParams))
	require.NoError(t, err)

	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthCreation(bundleID: "%s", in: %s) {
				id
			}}`, bndlID, in))

	var resp graphql.BundleInstanceAuth

	err = testctx.Tc.RunOperation(ctx, gqlClient, req, &resp)
	require.NoError(t, err)

	return resp
}

func DeleteBundleInstanceAuth(t *testing.T, ctx context.Context, gqlClient *gcli.Client, authID string) {
	deleteBundleInstanceAuthReq := FixDeleteBundleInstanceAuthRequest(authID)
	output := graphql.BundleInstanceAuth{}

	t.Log("Deleting the bundle instance auth...")
	err := testctx.Tc.RunOperation(ctx, gqlClient, deleteBundleInstanceAuthReq, &output)

	require.NoError(t, err)
}
