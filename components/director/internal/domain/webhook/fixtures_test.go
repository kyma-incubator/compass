package webhook_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var fixColumns = []string{"id", "app_id", "app_template_id", "type", "url", "auth", "runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template", "created_at", "formation_template_id"}

var (
	emptyTemplate = `{}`
	testURL       = "testURL"
)

func stringPtr(s string) *string {
	return &s
}

func fixApplicationModelWebhook(id, appID, tenant, url string, createdAt time.Time) *model.Webhook {
	appWebhook := fixGenericModelWebhook(id, appID, url)
	appWebhook.ObjectType = model.ApplicationWebhookReference
	appWebhook.CreatedAt = &createdAt
	return appWebhook
}

func fixRuntimeModelWebhook(id, runtimeID, url string) *model.Webhook {
	runtimeWebhook := fixGenericModelWebhook(id, runtimeID, url)
	runtimeWebhook.ObjectType = model.RuntimeWebhookReference
	return runtimeWebhook
}

func fixFormationTemplateModelWebhook(id, formationTemplateID, url string) *model.Webhook {
	formationTmplWebhook := fixGenericModelWebhook(id, formationTemplateID, url)
	formationTmplWebhook.ObjectType = model.FormationTemplateWebhookReference
	formationTmplWebhook.Type = model.WebhookTypeFormationLifecycle
	return formationTmplWebhook
}

func fixApplicationTemplateModelWebhook(id, appTemplateID, url string) *model.Webhook {
	appTmplWebhook := fixGenericModelWebhook(id, appTemplateID, url)
	appTmplWebhook.ObjectType = model.ApplicationTemplateWebhookReference
	return appTmplWebhook
}

func fixIntegrationSystemModelWebhook(id, intSysID, url string) *model.Webhook {
	intSysWebhook := fixGenericModelWebhook(id, intSysID, url)
	intSysWebhook.ObjectType = model.IntegrationSystemWebhookReference
	intSysWebhook.Type = ""
	return intSysWebhook
}

func fixGenericModelWebhook(id, objectID, url string) *model.Webhook {
	return &model.Webhook{
		ID:             id,
		ObjectID:       objectID,
		ObjectType:     model.UnknownWebhookReference,
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

func fixApplicationGQLWebhook(id, appID, url string) *graphql.Webhook {
	appWebhook := fixGenericGQLWebhook(id, url)
	appWebhook.ApplicationID = &appID
	appWebhook.CreatedAt = &graphql.Timestamp{}
	return appWebhook
}

func fixRuntimeGQLWebhook(id, rtmID, url string) *graphql.Webhook {
	rtmWebhook := fixGenericGQLWebhook(id, url)
	rtmWebhook.RuntimeID = &rtmID
	return rtmWebhook
}

func fixApplicationTemplateGQLWebhook(id, appTmplID, url string) *graphql.Webhook {
	appTmplWebhook := fixGenericGQLWebhook(id, url)
	appTmplWebhook.ApplicationTemplateID = &appTmplID
	return appTmplWebhook
}

func fixFormationTemplateGQLWebhook(id, formationTmplID, url string) *graphql.Webhook {
	formationTmplWebhook := fixGenericGQLWebhook(id, url)
	formationTmplWebhook.FormationTemplateID = &formationTmplID
	formationTmplWebhook.Type = graphql.WebhookTypeFormationLifecycle
	return formationTmplWebhook
}

func fixIntegrationSystemGQLWebhook(id, intSysID, url string) *graphql.Webhook {
	intSysWebhook := fixGenericGQLWebhook(id, url)
	intSysWebhook.IntegrationSystemID = &intSysID
	intSysWebhook.Type = ""
	return intSysWebhook
}

func fixGenericGQLWebhook(id, url string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:             id,
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

func fixApplicationModelWebhookWithType(id, appID, tenant, url string, webhookType model.WebhookType, createdAt time.Time) (w *model.Webhook) {
	w = fixApplicationModelWebhook(id, appID, tenant, url, createdAt)
	w.Type = webhookType
	return
}

func fixApplicationTemplateModelWebhookWithType(id, appTemplateID, url string, webhookType model.WebhookType) (w *model.Webhook) {
	w = fixApplicationTemplateModelWebhook(id, appTemplateID, url)
	w.Type = webhookType
	return
}

func fixApplicationTemplateModelWebhookWithTypeAndTimestamp(id, appTemplateID, url string, webhookType model.WebhookType, createdAt time.Time) (w *model.Webhook) {
	w = fixApplicationTemplateModelWebhookWithType(id, appTemplateID, url, webhookType)
	w.CreatedAt = &createdAt
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

func fixApplicationWebhookEntity(t *testing.T, createdAt time.Time) *webhook.Entity {
	return fixApplicationWebhookEntityWithID(t, givenID(), createdAt)
}

func fixApplicationWebhookEntityWithID(t *testing.T, id string, createdAt time.Time) *webhook.Entity {
	return fixApplicationWebhookEntityWithIDAndWebhookType(t, id, model.WebhookTypeConfigurationChanged, createdAt)
}

func fixApplicationWebhookEntityWithIDAndWebhookType(t *testing.T, id string, whType model.WebhookType, createdAt time.Time) *webhook.Entity {
	return &webhook.Entity{
		ID:             id,
		ApplicationID:  repo.NewValidNullableString(givenApplicationID()),
		Type:           string(whType),
		URL:            repo.NewValidNullableString("http://kyma.io"),
		Mode:           repo.NewValidNullableString(string(model.WebhookModeSync)),
		Auth:           sql.NullString{Valid: true, String: fixAuthAsAString(t)},
		URLTemplate:    repo.NewValidNullableString(emptyTemplate),
		InputTemplate:  repo.NewValidNullableString(emptyTemplate),
		HeaderTemplate: repo.NewValidNullableString(emptyTemplate),
		OutputTemplate: repo.NewValidNullableString(emptyTemplate),
		CreatedAt:      &createdAt,
	}
}

func fixRuntimeWebhookEntityWithID(t *testing.T, id string) *webhook.Entity {
	return &webhook.Entity{
		ID:             id,
		RuntimeID:      repo.NewValidNullableString(givenRuntimeID()),
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

func fixFormationTemplateWebhookEntityWithID(t *testing.T, id string) *webhook.Entity {
	return &webhook.Entity{
		ID:                  id,
		FormationTemplateID: repo.NewValidNullableString(givenFormationTemplateID()),
		Type:                string(model.WebhookTypeFormationLifecycle),
		URL:                 repo.NewValidNullableString("http://kyma.io"),
		Mode:                repo.NewValidNullableString(string(model.WebhookModeSync)),
		Auth:                sql.NullString{Valid: true, String: fixAuthAsAString(t)},
		URLTemplate:         repo.NewValidNullableString(emptyTemplate),
		InputTemplate:       repo.NewValidNullableString(emptyTemplate),
		HeaderTemplate:      repo.NewValidNullableString(emptyTemplate),
		OutputTemplate:      repo.NewValidNullableString(emptyTemplate),
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
		CreatedAt:             nil,
	}
}

func fixApplicationTemplateWebhookEntityWithTimestamp(t *testing.T, createdAt time.Time) *webhook.Entity {
	w := fixApplicationTemplateWebhookEntity(t)
	w.CreatedAt = &createdAt
	return w
}

func newModelBusinessTenantMappingWithType(tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             givenTenant(),
		Name:           "name",
		ExternalTenant: "external",
		Parent:         givenParentTenant(),
		Type:           tenantType,
		Provider:       "test-provider",
		Status:         tenant.Active,
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

func givenParentTenant() string {
	return "b92b59f7-2563-40b2-aba9-fef726037aa3"
}

func givenExternalTenant() string {
	return "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
}

func givenApplicationID() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}

func givenRuntimeID() string {
	return "rrrrrrrr-rrrr-rrrr-rrrr-rrrrrrrrrrrr"
}

func givenFormationTemplateID() string {
	return "rrrrrrrr-rrrr-rrrr-rrrr-rrrrrrrrrrrr"
}

func givenApplicationTemplateID() string {
	return "ffffffff-ffff-ffff-ffff-ffffffffffff"
}

func givenError() error {
	return errors.New("some error")
}

func fixEmptyTenantMappingConfig() map[string]interface{} {
	tenantMappingJSON := "{}"
	return GetTenantMappingConfig(tenantMappingJSON)
}

func fixTenantMappingConfig() map[string]interface{} {
	tenantMappingJSON := "{\"SYNC\": {\"v1.0\": [{ \"type\": \"CONFIGURATION_CHANGED\",\"urlTemplate\": \"%s\",\"inputTemplate\": \"input template\",\"headerTemplate\": \"header template\",\"outputTemplate\": \"output template\"}]}}"
	return GetTenantMappingConfig(tenantMappingJSON)
}

func fixTenantMappingConfigForAsyncCallback() map[string]interface{} {
	tenantMappingJSON := "{\"ASYNC_CALLBACK\": {\"v1.0\": [{ \"type\": \"CONFIGURATION_CHANGED\",\"urlTemplate\": \"%s\",\"inputTemplate\": \"input template\",\"headerTemplate\": \"%s\",\"outputTemplate\": \"output template\"}]}}"
	return GetTenantMappingConfig(tenantMappingJSON)
}

func fixInvalidTenantMappingConfig() map[string]interface{} {
	tenantMappingJSON := "{\"SYNC\": []}"
	return GetTenantMappingConfig(tenantMappingJSON)
}

func fixTenantMappedWebhooks() []*graphql.WebhookInput {
	syncMode := graphql.WebhookModeSync

	return []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			Auth:    nil,
			Mode:    &syncMode,
			URL:     &testURL,
			Version: str.Ptr("v1.0"),
		},
	}
}

func fixTenantMappedWebhooksForAsyncCallbackMode() []*graphql.WebhookInput {
	asyncMode := graphql.WebhookModeAsyncCallback

	return []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			Auth:    nil,
			Mode:    &asyncMode,
			URL:     &testURL,
			Version: str.Ptr("v1.0"),
		},
	}
}

func fixTenantMappedWebhooksWithInvalidVersion() []*graphql.WebhookInput {
	syncMode := graphql.WebhookModeSync

	return []*graphql.WebhookInput{
		{
			Type:    graphql.WebhookTypeConfigurationChanged,
			Auth:    nil,
			Mode:    &syncMode,
			URL:     &testURL,
			Version: str.Ptr("notfound"),
		},
	}
}

func fixEnrichedTenantMappedWebhooks() []*graphql.WebhookInput {
	syncMode := graphql.WebhookModeSync

	return []*graphql.WebhookInput{
		{
			Type:           graphql.WebhookTypeConfigurationChanged,
			Auth:           nil,
			Mode:           &syncMode,
			URLTemplate:    &testURL,
			InputTemplate:  str.Ptr("input template"),
			HeaderTemplate: str.Ptr("header template"),
			OutputTemplate: str.Ptr("output template"),
		},
	}
}

func fixEnrichedTenantMappedWebhooksForAsyncCallbackMode(callbackURL string) []*graphql.WebhookInput {
	asyncMode := graphql.WebhookModeAsyncCallback

	return []*graphql.WebhookInput{
		{
			Type:           graphql.WebhookTypeConfigurationChanged,
			Auth:           nil,
			Mode:           &asyncMode,
			URLTemplate:    &testURL,
			InputTemplate:  str.Ptr("input template"),
			HeaderTemplate: str.Ptr(callbackURL),
			OutputTemplate: str.Ptr("output template"),
		},
	}
}
