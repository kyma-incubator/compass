package webhook_test

import (
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var emptyTemplate = `{}`

func stringPtr(s string) *string {
	return &s
}

var fixAccessStrategy = "accessStrategy"

func fixApplicationModelWebhook(id, appID, tenant, url string) *model.Webhook {
	return &model.Webhook{
		ID:             id,
		ApplicationID:  &appID,
		TenantID:       &tenant,
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           &model.Auth{},
		AccessStrategy: &fixAccessStrategy,
		Mode:           &modelWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixApplicationTemplateModelWebhook(id, appTemplateID, url string) *model.Webhook {
	return &model.Webhook{
		ID:                    id,
		ApplicationTemplateID: &appTemplateID,
		Type:                  model.WebhookTypeConfigurationChanged,
		URL:                   &url,
		Auth:                  &model.Auth{},
		AccessStrategy:        &fixAccessStrategy,
		Mode:                  &modelWebhookMode,
		URLTemplate:           &emptyTemplate,
		InputTemplate:         &emptyTemplate,
		HeaderTemplate:        &emptyTemplate,
		OutputTemplate:        &emptyTemplate,
	}
}

func fixGQLWebhook(id, appID, url string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:             id,
		ApplicationID:  &appID,
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           &graphql.Auth{},
		AccessStrategy: &fixAccessStrategy,
		Mode:           &graphqlWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixModelWebhookInput(url string) *model.WebhookInput {
	return &model.WebhookInput{
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           &model.AuthInput{},
		AccessStrategy: &fixAccessStrategy,
		Mode:           &modelWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixGQLWebhookInput(url string) *graphql.WebhookInput {
	return &graphql.WebhookInput{
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           &graphql.AuthInput{},
		AccessStrategy: &fixAccessStrategy,
		Mode:           &graphqlWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixApplicationModelWebhookWithType(id, appID, tenant, url string, webhookType model.WebhookType) (w *model.Webhook) {
	w = fixApplicationModelWebhook(id, appID, tenant, url)
	w.Type = webhookType
	return
}

func fixApplicationTemplateModelWebhookWithType(id, appTemplateID, url string, webhookType model.WebhookType) (w *model.Webhook) {
	w = fixApplicationTemplateModelWebhook(id, appTemplateID, url)
	w.Type = webhookType
	return
}

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
