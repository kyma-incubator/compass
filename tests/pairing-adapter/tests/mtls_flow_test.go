package tests

import (
	"context"
	"testing"

	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/require"
)

func TestGettingTokenWithMTLSWorks(t *testing.T) {
	ctx := context.Background()
	defaultTestTenant := tenant.TestTenants.GetDefaultTenantID()

	expectedProductType := "SAP SuccessFactors"
	namePlaceholderKey := "name"
	displayNamePlaceholderKey := "display-name"
	appTmplInput := directorSchema.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*directorSchema.TemplateValueInput{
			{
				Placeholder: namePlaceholderKey,
				Value:       "E2E test SuccessFactors app",
			},
			{
				Placeholder: displayNamePlaceholderKey,
				Value:       "E2E test SuccessFactors app Display Name",
			},
		},
	}

	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appTmplInput)
	require.NoError(t, err)

	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := directorSchema.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, defaultTestTenant, createAppFromTmplRequest, &outputApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, defaultTestTenant, &outputApp)
	require.NoError(t, err)
	require.NotEmpty(t, outputApp.ID)

	token := fixtures.RequestOneTimeTokenForApplication(t, ctx, certSecuredGraphQLClient, outputApp.ID)
	require.NotEmpty(t, token.Token)
	require.NotEmpty(t, token.ConnectorURL)
	require.NotEmpty(t, token.LegacyConnectorURL)
}
