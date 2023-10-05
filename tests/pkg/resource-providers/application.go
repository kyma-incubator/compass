package resource_providers

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

type ApplicationProvider struct {
	applicationType        string
	namePlaceholder        string
	name                   string
	displayNamePlaceholder string
	displayName            string
	tenantID               string
	app                    graphql.ApplicationExt
}

func NewApplicationProvider(applicationType, namePlaceholder, name, displayNamePlaceholder, displayName, tenantID string) *ApplicationProvider {
	a := &ApplicationProvider{
		applicationType:        applicationType,
		namePlaceholder:        namePlaceholder,
		name:                   name,
		displayNamePlaceholder: displayNamePlaceholder,
		displayName:            displayName,
		tenantID:               tenantID,
	}

	return a
}

func (a *ApplicationProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(a.applicationType, a.namePlaceholder, a.name, a.displayNamePlaceholder, a.displayName)
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, a.tenantID, createAppFromTmplFirstRequest, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)
	a.app = app
	return app.ID
}

func (a *ApplicationProvider) TearDown(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupApplication(t, ctx, gqlClient, a.tenantID, &a.app)
}
