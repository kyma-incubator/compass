package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func GetApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationExt {
	appRequest := FixGetApplicationRequest(id)
	app := graphql.ApplicationExt{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, appRequest, &app)
	require.NoError(t, err)
	return app
}

func UpdateApplicationWithinTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string, in graphql.ApplicationUpdateInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(in)
	require.NoError(t, err)

	createRequest := FixUpdateApplicationRequest(id, appInputGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createRequest, &app)
	return app, err
}

func RegisterApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, name, tenant string) graphql.ApplicationExt {
	in := FixSampleApplicationRegisterInputWithName("first", name)
	app, err := RegisterApplicationFromInput(t, ctx, gqlClient, tenant, in)
	require.NoError(t, err)
	return app
}

func RegisterApplicationFromInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string, in graphql.ApplicationRegisterInput) (graphql.ApplicationExt, error) {
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := FixRegisterApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &app)
	return app, err
}

func RequestClientCredentialsForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.AppSystemAuth {
	req := FixRequestClientCredentialsForApplication(id)
	systemAuth := graphql.AppSystemAuth{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &systemAuth)
	require.NoError(t, err)
	return systemAuth
}

func UnregisterApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.ApplicationExt {
	if id == "" {
		return graphql.ApplicationExt{}
	}
	deleteRequest := FixUnregisterApplicationRequest(id)
	app := graphql.ApplicationExt{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, deleteRequest, &app)
	require.NoError(t, err)
	return app
}

func UnregisterAsyncApplicationInTenant(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, id string) {
	req := FixAsyncUnregisterApplicationRequest(id)
	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func DeleteApplicationLabel(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id, labelKey string) {
	deleteRequest := FixDeleteApplicationLabelRequest(id, labelKey)

	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, deleteRequest, nil))
}

func SetApplicationLabel(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string, labelKey string, labelValue interface{}) graphql.Label {
	setLabelRequest := FixSetApplicationLabelRequest(id, labelKey, labelValue)
	label := graphql.Label{}
	err := testctx.Tc.RunOperation(ctx, gqlClient, setLabelRequest, &label)
	require.NoError(t, err)

	return label
}

func GenerateClientCredentialsForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) graphql.AppSystemAuth {
	req := FixRequestClientCredentialsForApplication(id)

	out := graphql.AppSystemAuth{}
	err := testctx.Tc.RunOperation(ctx, gqlClient, req, &out)
	require.NoError(t, err)

	return out
}

func DeleteSystemAuthForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) {
	req := FixDeleteSystemAuthForApplicationRequest(id)
	err := testctx.Tc.RunOperation(ctx, gqlClient, req, nil)
	require.NoError(t, err)
}

func SetDefaultEventingForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, appID string, runtimeID string) {
	req := FixSetDefaultEventingForApplication(appID, runtimeID)
	err := testctx.Tc.RunOperation(ctx, gqlClient, req, nil)
	require.NoError(t, err)
}

func RegisterSimpleApp(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID string) func() {
	placeholder := "foo"
	in := FixSampleApplicationRegisterInputWithWebhooks(placeholder)
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	var res graphql.Application
	req := FixRegisterApplicationRequest(appInputGQL)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, req, &res)
	require.NoError(t, err)

	return func() {
		UnregisterApplication(t, ctx, gqlClient, tenantID, res.ID)
	}
}

func RequestOneTimeTokenForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) graphql.OneTimeTokenForApplicationExt {
	tokenRequest := FixRequestOneTimeTokenForApplication(id)
	token := graphql.OneTimeTokenForApplicationExt{}
	err := testctx.Tc.RunOperation(ctx, gqlClient, tokenRequest, &token)
	require.NoError(t, err)
	return token
}

func GenerateOneTimeTokenForApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, id string) graphql.OneTimeTokenForApplicationExt {
	req := FixRequestOneTimeTokenForApplication(id)
	oneTimeToken := graphql.OneTimeTokenForApplicationExt{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &oneTimeToken)
	require.NoError(t, err)

	require.NotEmpty(t, oneTimeToken.ConnectorURL)
	require.NotEmpty(t, oneTimeToken.Token)
	require.NotEmpty(t, oneTimeToken.Raw)
	require.NotEmpty(t, oneTimeToken.RawEncoded)
	require.NotEmpty(t, oneTimeToken.LegacyConnectorURL)
	return oneTimeToken
}

func UnassignApplicationFromScenarios(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID, applicationID string) {
	labelKey := "scenarios"
	defaultValue := "DEFAULT"

	scenarios := []string{defaultValue}
	var labelValue interface{} = scenarios

	setLabelRequest := FixSetApplicationLabelRequest(applicationID, labelKey, labelValue)

	label := graphql.Label{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, setLabelRequest, &label)
	require.NoError(t, err)
}
