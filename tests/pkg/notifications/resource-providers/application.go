package resource_providers

import (
	"context"
	"testing"

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
	tenantID               string
	app                    graphql.ApplicationExt
}

func NewApplicationProvider(applicationType, namePlaceholder, name, displayNamePlaceholder, displayName, tenantID string) *ApplicationProvider {
	p := &ApplicationProvider{
		applicationType:        applicationType,
		namePlaceholder:        namePlaceholder,
		name:                   name,
		displayNamePlaceholder: displayNamePlaceholder,
		displayName:            displayName,
		tenantID:               tenantID,
	}

	return p
}

func (p *ApplicationProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(p.applicationType, p.namePlaceholder, p.name, p.displayNamePlaceholder, p.displayName)
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, p.tenantID, createAppFromTmplFirstRequest, &app)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)
	p.app = app
	return app.ID
}

func (p *ApplicationProvider) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupApplication(t, ctx, gqlClient, p.tenantID, &p.app)
}
