package webhook_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var fixColumns = []string{"id", "app_id", "app_template_id", "type", "url", "auth", "runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}

var emptyTemplate = `{}`

func stringPtr(s string) *string {
	return &s
}

func fixApplicationModelWebhook(id, appID, tenant, url string) *model.Webhook {
	return &model.Webhook{
		ID:             id,
		ObjectID:       appID,
		ObjectType:     model.ApplicationWebhookReference,
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           fixBasicAuth(),
		Mode:           &modelWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixApplicationTemplateModelWebhook(id, appTemplateID, url string) *model.Webhook {
	return &model.Webhook{
		ID:             id,
		ObjectID:       appTemplateID,
		ObjectType:     model.ApplicationTemplateWebhookReference,
		Type:           model.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           fixBasicAuth(),
		Mode:           &modelWebhookMode,
		URLTemplate:    &emptyTemplate,
		InputTemplate:  &emptyTemplate,
		HeaderTemplate: &emptyTemplate,
		OutputTemplate: &emptyTemplate,
	}
}

func fixGQLWebhook(id, appID, url string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:             id,
		ApplicationID:  &appID,
		Type:           graphql.WebhookTypeConfigurationChanged,
		URL:            &url,
		Auth:           &graphql.Auth{},
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

func fixBasicAuth() *model.Auth {
	return &model.Auth{
		Credential: model.CredentialData{
			Basic: &model.BasicCredentialData{
				Username: "aaa",
				Password: "bbb",
			},
		},
	}
}

func fixAuthAsAString(t *testing.T) string {
	b, err := json.Marshal(fixBasicAuth())
	require.NoError(t, err)
	return string(b)
}

func fixApplicationWebhookEntity(t *testing.T) *webhook.Entity {
	return fixApplicationWebhookEntityWithID(t, givenID())
}

func fixApplicationWebhookEntityWithID(t *testing.T, id string) *webhook.Entity {
	return &webhook.Entity{
		ID:             id,
		ApplicationID:  repo.NewValidNullableString(givenApplicationID()),
		Type:           string(model.WebhookTypeConfigurationChanged),
		URL:            repo.NewValidNullableString("http://kyma.io"),
		Mode:           repo.NewValidNullableString(string(model.WebhookModeSync)),
		Auth:           sql.NullString{Valid: true, String: fixAuthAsAString(t)},
		URLTemplate:    repo.NewValidNullableString(emptyTemplate),
		InputTemplate:  repo.NewValidNullableString(emptyTemplate),
		HeaderTemplate: repo.NewValidNullableString(emptyTemplate),
		OutputTemplate: repo.NewValidNullableString(emptyTemplate),
	}
}

func fixApplicationTemplateWebhookEntity(t *testing.T) *webhook.Entity {
	return &webhook.Entity{
		ID:                    givenID(),
		ApplicationTemplateID: repo.NewValidNullableString(givenApplicationTemplateID()),
		Type:                  string(model.WebhookTypeConfigurationChanged),
		URL:                   repo.NewValidNullableString("http://kyma.io"),
		Mode:                  repo.NewValidNullableString(string(model.WebhookModeSync)),
		Auth:                  sql.NullString{Valid: true, String: fixAuthAsAString(t)},
		URLTemplate:           repo.NewValidNullableString(emptyTemplate),
		InputTemplate:         repo.NewValidNullableString(emptyTemplate),
		HeaderTemplate:        repo.NewValidNullableString(emptyTemplate),
		OutputTemplate:        repo.NewValidNullableString(emptyTemplate),
	}
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func anotherID() string {
	return "dddddddd-dddd-dddd-dddd-dddddddddddd"
}

func givenTenant() string {
	return "b91b59f7-2563-40b2-aba9-fef726037aa3"
}

func givenExternalTenant() string {
	return "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
}

func givenApplicationID() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}

func givenApplicationTemplateID() string {
	return "ffffffff-ffff-ffff-ffff-ffffffffffff"
}

func givenError() error {
	return errors.New("some error")
}
