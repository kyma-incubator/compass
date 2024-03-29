package resource_providers

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type ApplicationTemplateProvider struct {
	applicationTemplate     graphql.ApplicationTemplate
	applicationType         string
	localTenantID           string
	region                  string
	namespace               string
	namePlaceholder         string
	displayNamePlaceholder  string
	tenantID                string
	applicationWebhookInput *graphql.WebhookInput
	condition               graphql.ApplicationStatusCondition
}

func NewApplicationTemplateProvider(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder, tenantID string, webhookInput *graphql.WebhookInput, condition graphql.ApplicationStatusCondition) *ApplicationTemplateProvider {
	p := &ApplicationTemplateProvider{
		applicationType:         applicationType,
		localTenantID:           localTenantID,
		region:                  region,
		namespace:               namespace,
		namePlaceholder:         namePlaceholder,
		displayNamePlaceholder:  displayNamePlaceholder,
		tenantID:                tenantID,
		applicationWebhookInput: webhookInput,
		condition:               condition,
	}

	return p
}

func (p *ApplicationTemplateProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) graphql.ApplicationTemplate {
	in := fixtures.FixApplicationTemplateWithWebhookInput(p.applicationType, p.localTenantID, p.region, p.namespace, p.namePlaceholder, p.displayNamePlaceholder, p.applicationWebhookInput, p.condition)
	appTpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, gqlClient, "", in)
	require.NoError(t, err)
	p.applicationTemplate = appTpl
	return appTpl
}

func (p *ApplicationTemplateProvider) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupApplicationTemplate(t, ctx, gqlClient, "", p.applicationTemplate)
}

func (p *ApplicationTemplateProvider) GetResource() Resource {
	return NewApplicationTemplateResource(p.applicationTemplate)
}

type ApplicationTemplateResource struct {
	tpl graphql.ApplicationTemplate
}

func NewApplicationTemplateResource(tpl graphql.ApplicationTemplate) *ApplicationTemplateResource {
	return &ApplicationTemplateResource{
		tpl: tpl,
	}
}

func (p *ApplicationTemplateResource) GetType() graphql.ResourceType {
	return graphql.ResourceTypeApplication
}

func (p *ApplicationTemplateResource) GetName() string {
	return p.tpl.Name
}

// GetArtifactKind used only for runtimes, otherwise return empty
func (p *ApplicationTemplateResource) GetArtifactKind() *graphql.ArtifactType {
	return nil
}

func (p *ApplicationTemplateResource) GetDisplayName() *string {
	return nil
}

func (p *ApplicationTemplateResource) GetID() string {
	return p.tpl.ID
}
