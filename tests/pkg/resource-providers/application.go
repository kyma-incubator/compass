package resource_providers

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type ApplicationProvider struct {
	applicationType        string
	namePlaceholder        string
	name                   string
	displayNamePlaceholder string
	displayName            string
	app                    graphql.ApplicationExt
}

func NewApplicationProvider(applicationType, namePlaceholder, name, displayNamePlaceholder, displayName string) *ApplicationProvider {
	a := &ApplicationProvider{
		applicationType:        applicationType,
		namePlaceholder:        namePlaceholder,
		name:                   name,
		displayNamePlaceholder: displayNamePlaceholder,
		displayName:            displayName,
	}

	return a
}

func (a *ApplicationProvider) Provide(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string) graphql.ApplicationExt {
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(a.applicationType, a.namePlaceholder, a.name, a.displayNamePlaceholder, a.displayName)
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, createAppFromTmplFirstRequest, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)
	a.app = app
	return app
}

func (a *ApplicationProvider) TearDown(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string) {
	fixtures.CleanupApplication(t, ctx, gqlClient, tenant, &a.app)
}
