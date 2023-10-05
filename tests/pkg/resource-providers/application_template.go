package resource_providers

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
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
}

func NewApplicationTemplateProvider(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder, tenantID string, webhookInput *graphql.WebhookInput) *ApplicationTemplateProvider {
	a := &ApplicationTemplateProvider{
		applicationType:         applicationType,
		localTenantID:           localTenantID,
		region:                  region,
		namespace:               namespace,
		namePlaceholder:         namePlaceholder,
		displayNamePlaceholder:  displayNamePlaceholder,
		tenantID:                tenantID,
		applicationWebhookInput: webhookInput,
	}

	return a
}

func (a *ApplicationTemplateProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	in := fixtures.FixApplicationTemplateWithWebhookInput(a.applicationType, a.localTenantID, a.region, a.namespace, a.namePlaceholder, a.displayNamePlaceholder, a.applicationWebhookInput)
	appTpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, gqlClient, a.tenantID, in)
	require.NoError(t, err)
	a.applicationTemplate = appTpl
	return appTpl.ID
}

func (a *ApplicationTemplateProvider) TearDown(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupApplicationTemplate(t, ctx, gqlClient, a.tenantID, a.applicationTemplate)
}

func (a *ApplicationTemplateProvider) GetResource() Resource {
	return NewApplicationTemplateParticipant(a.applicationTemplate)
}
