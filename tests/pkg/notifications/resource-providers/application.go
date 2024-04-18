package resource_providers

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type ApplicationProvider struct {
	applicationType         string
	applicationTypeLabelKey string
	name                    string
	tenantID                string
	app                     graphql.ApplicationExt
}

func NewApplicationProvider(applicationType, applicationTypeLabelKey, name, tenantID string) *ApplicationProvider {
	p := &ApplicationProvider{
		applicationType:         applicationType,
		applicationTypeLabelKey: applicationTypeLabelKey,
		name:                    name,
		tenantID:                tenantID,
	}

	return p
}

func (p *ApplicationProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	app, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, gqlClient, p.name, p.applicationTypeLabelKey, p.applicationType, p.tenantID)
	require.NoError(t, err)
	require.NotEmpty(t, app.ID)

	p.app = app
	return app.ID
}

func (p *ApplicationProvider) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupApplication(t, ctx, gqlClient, p.tenantID, &p.app)
}

func (p *ApplicationProvider) GetResource() Resource {
	return NewApplicationResource(*p)
}

type ApplicationResource struct {
	appProvider ApplicationProvider
}

func NewApplicationResource(appProvider ApplicationProvider) *ApplicationResource {
	return &ApplicationResource{
		appProvider: appProvider,
	}
}

func (p *ApplicationResource) GetType() graphql.ResourceType {
	return graphql.ResourceTypeApplication
}

func (p *ApplicationResource) GetName() string {
	return p.appProvider.applicationType
}

// GetArtifactKind used only for runtimes, otherwise return empty
func (p *ApplicationResource) GetArtifactKind() *graphql.ArtifactType {
	return nil
}

func (p *ApplicationResource) GetDisplayName() *string {
	return nil
}

func (p *ApplicationResource) GetID() string {
	return p.appProvider.app.ID
}
