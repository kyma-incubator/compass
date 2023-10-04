package resource_providers

import (
	"context"
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
	applicationWebhookInput *graphql.WebhookInput
}

func NewApplicationTemplateProvider(applicationType, localTenantID, region, namespace, namePlaceholder, displayNamePlaceholder string, webhookInput *graphql.WebhookInput) *ApplicationTemplateProvider {
	a := &ApplicationTemplateProvider{
		applicationType:         applicationType,
		localTenantID:           localTenantID,
		region:                  region,
		namespace:               namespace,
		namePlaceholder:         namePlaceholder,
		displayNamePlaceholder:  displayNamePlaceholder,
		applicationWebhookInput: webhookInput,
	}

	return a
}

func (a *ApplicationTemplateProvider) Provide(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string) Participant {
	in := fixtures.FixApplicationTemplateWithWebhookInput(a.applicationType, a.localTenantID, a.region, a.namespace, a.namePlaceholder, a.displayNamePlaceholder, a.applicationWebhookInput)
	appTpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, gqlClient, tenant, in)
	require.NoError(t, err)
	a.applicationTemplate = appTpl
	return NewApplicationTemplateParticipant(appTpl)
}

func (a *ApplicationTemplateProvider) TearDown(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string) {
	fixtures.CleanupApplicationTemplate(t, ctx, gqlClient, tenant, a.applicationTemplate)
}
