package apptemplate_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

const (
	testTenant         = "tnt"
	testExternalTenant = "external-tnt"
	testID             = "foo"

	testWebhookID      = "webhook-id-1"
	testName           = "bar"
	testPageSize       = 3
	testCursor         = ""
	appInputJSONString = `{"Name":"foo","ProviderName":"compass","Description":"Lorem ipsum","Labels":{"test":["val","val2"]},"HealthCheckURL":"https://foo.bar","Webhooks":[{"Type":"","URL":"webhook1.foo.bar","Auth":null},{"Type":"","URL":"webhook2.foo.bar","Auth":null}],"IntegrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputGQLString  = `{name: "foo",providerName: "compass",description: "Lorem ipsum",labels: {test:["val","val2"],},webhooks: [ {type: ,url: "webhook1.foo.bar",}, {type: ,url: "webhook2.foo.bar",} ],healthCheckURL: "https://foo.bar",integrationSystemID: "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii",}`
)

var (
	testUUID                       = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testDescription                = "Lorem ipsum"
	testDescriptionWithPlaceholder = "Lorem ipsum {{test}}"
	testProviderName               = "provider-display-name"
	testURL                        = "http://valid.url"
	testError                      = errors.New("test error")
	testTableColumns               = []string{"id", "name", "description", "application_input", "placeholders", "access_level"}
)

func fixModelApplicationTemplate(id, name string, webhooks []*model.Webhook) *model.ApplicationTemplate {
	desc := testDescription
	out := model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholders(),
		Webhooks:             modelPtrsToWebhooks(webhooks),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}

	return &out
}

func fixModelAppTemplateWithAppInputJSON(id, name, appInputJSON string, webhooks []*model.Webhook) *model.ApplicationTemplate {
	out := fixModelApplicationTemplate(id, name, webhooks)
	out.ApplicationInputJSON = appInputJSON

	return out
}

func fixModelAppTemplateWithAppInputJSONAndPlaceholders(id, name, appInputJSON string, webhooks []*model.Webhook, placeholders []model.ApplicationTemplatePlaceholder) *model.ApplicationTemplate {
	out := fixModelAppTemplateWithAppInputJSON(id, name, appInputJSON, webhooks)
	out.Placeholders = placeholders

	return out
}

func fixGQLAppTemplate(id, name string, webhooks []*graphql.Webhook) *graphql.ApplicationTemplate {
	desc := testDescription

	return &graphql.ApplicationTemplate{
		ID:               id,
		Name:             name,
		Description:      &desc,
		ApplicationInput: appInputGQLString,
		Placeholders:     fixGQLPlaceholders(),
		Webhooks:         gqlPtrsToWebhooks(webhooks),
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixModelAppTemplatePage(appTemplates []*model.ApplicationTemplate) model.ApplicationTemplatePage {
	return model.ApplicationTemplatePage{
		Data: appTemplates,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(appTemplates),
	}
}

func fixGQLAppTemplatePage(appTemplates []*graphql.ApplicationTemplate) graphql.ApplicationTemplatePage {
	return graphql.ApplicationTemplatePage{
		Data: appTemplates,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(appTemplates),
	}
}

func fixModelAppTemplateInput(name string, appInputString string) *model.ApplicationTemplateInput {
	desc := testDescription

	return &model.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholders(),
		Labels:               map[string]interface{}{"test": "test"},
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixModelAppTemplateWithIDInput(name, appInputString string, id *string) *model.ApplicationTemplateInput {
	model := fixModelAppTemplateInput(name, appInputString)
	model.ID = id

	return model
}

func fixModelAppTemplateUpdateInput(name string, appInputString string) *model.ApplicationTemplateUpdateInput {
	desc := testDescription

	return &model.ApplicationTemplateUpdateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholders(),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixModelAppTemplateUpdateInputWithPlaceholders(name string, appInputString string, placeholders []model.ApplicationTemplatePlaceholder) *model.ApplicationTemplateUpdateInput {
	out := fixModelAppTemplateUpdateInput(name, appInputString)
	out.Placeholders = placeholders

	return out
}

func fixGQLAppTemplateInput(name string) *graphql.ApplicationTemplateInput {
	desc := testDescription

	return &graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		Labels:       map[string]interface{}{"test": "test"},
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputWithPlaceholder(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputInvalidAppInputURLTemplateMethod(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
			Webhooks: []*graphql.WebhookInput{
				{
					Type:        "ASYNC",
					URLTemplate: str.Ptr(`{"path": "https://target.url", "method":"invalid method"}`),
				},
			},
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInput(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescription

	return &graphql.ApplicationTemplateUpdateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInputWithPlaceholder(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateUpdateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInputInvalidAppInput(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateUpdateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationRegisterInput{
			Name:        "foo",
			Description: &desc,
			Webhooks: []*graphql.WebhookInput{
				{
					Type:        "ASYNC",
					URLTemplate: str.Ptr(`{"path": "https://target.url", "method":"invalid method"}`),
				},
			},
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixEntityApplicationTemplate(t *testing.T, id, name string) *apptemplate.Entity {
	marshalledAppInput := `{"Name":"foo","ProviderName":"compass","Description":"Lorem ipsum","Labels":{"test":["val","val2"]},"HealthCheckURL":"https://foo.bar","Webhooks":[{"Type":"","URL":"webhook1.foo.bar","Auth":null},{"Type":"","URL":"webhook2.foo.bar","Auth":null}],"IntegrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`

	placeholders := fixModelPlaceholders()
	marshalledPlaceholders, err := json.Marshal(placeholders)
	require.NoError(t, err)

	return &apptemplate.Entity{
		ID:                   id,
		Name:                 name,
		Description:          repo.NewValidNullableString(testDescription),
		ApplicationInputJSON: marshalledAppInput,
		PlaceholdersJSON:     repo.NewValidNullableString(string(marshalledPlaceholders)),
		AccessLevel:          string(model.GlobalApplicationTemplateAccessLevel),
	}
}

func fixModelPlaceholders() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := testDescription
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixModelApplicationWebhooks(webhookID, applicationID string) []*model.Webhook {
	return []*model.Webhook{
		{
			ObjectID:   applicationID,
			ObjectType: model.ApplicationWebhookReference,
			ID:         webhookID,
			Type:       model.WebhookTypeConfigurationChanged,
			URL:        str.Ptr("foourl"),
			Auth:       &model.Auth{},
		},
	}
}

func fixModelApplicationTemplateWebhooks(webhookID, applicationTemplateID string) []*model.Webhook {
	return []*model.Webhook{
		{
			ObjectID:   applicationTemplateID,
			ObjectType: model.ApplicationTemplateWebhookReference,
			ID:         webhookID,
			Type:       model.WebhookTypeConfigurationChanged,
			URL:        str.Ptr("foourl"),
			Auth:       &model.Auth{},
		},
	}
}

func fixGQLPlaceholderDefinitionInput() []*graphql.PlaceholderDefinitionInput {
	placeholderDesc := testDescription
	return []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixGQLPlaceholders() []*graphql.PlaceholderDefinition {
	placeholderDesc := testDescription
	return []*graphql.PlaceholderDefinition{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixGQLApplicationWebhooks(webhookID, applicationID string) []*graphql.Webhook {
	return []*graphql.Webhook{
		{
			ID:            webhookID,
			ApplicationID: str.Ptr(applicationID),
			Type:          graphql.WebhookTypeConfigurationChanged,
			URL:           str.Ptr("foourl"),
			Auth:          &graphql.Auth{},
		},
	}
}

func fixGQLApplicationTemplateWebhooks(webhookID, applicationTemplateID string) []*graphql.Webhook {
	return []*graphql.Webhook{
		{
			ID:                    webhookID,
			ApplicationTemplateID: str.Ptr(applicationTemplateID),
			Type:                  graphql.WebhookTypeConfigurationChanged,
			URL:                   str.Ptr("foourl"),
			Auth:                  &graphql.Auth{},
		},
	}
}

func fixGQLApplicationFromTemplateInput(name string) graphql.ApplicationFromTemplateInput {
	return graphql.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*graphql.TemplateValueInput{
			{Placeholder: "a", Value: "b"},
			{Placeholder: "c", Value: "d"},
		},
	}
}

func fixModelApplicationFromTemplateInput(name string) model.ApplicationFromTemplateInput {
	return model.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*model.ApplicationTemplateValueInput{
			{Placeholder: "a", Value: "b"},
			{Placeholder: "c", Value: "d"},
		},
	}
}

func fixAppTemplateCreateArgs(entity apptemplate.Entity) []driver.Value {
	return []driver.Value{entity.ID, entity.Name, entity.Description, entity.ApplicationInputJSON, entity.PlaceholdersJSON, entity.AccessLevel}
}

func fixSQLRows(entities []apptemplate.Entity) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, entity := range entities {
		out.AddRow(entity.ID, entity.Name, entity.Description, entity.ApplicationInputJSON, entity.PlaceholdersJSON, entity.AccessLevel)
	}
	return out
}

func fixJSONApplicationCreateInput(name string) string {
	return fmt.Sprintf(`{"name": "%s", "providerName": "%s", "description": "%s", "healthCheckURL": "%s"}`, name, testProviderName, testDescription, testURL)
}

func fixModelApplicationCreateInput(name string) model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixModelApplicationWithLabelCreateInput(name string) model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
		Labels:         map[string]interface{}{"managed": "false"},
	}
}

func fixGQLApplicationCreateInput(name string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:           name,
		ProviderName:   &testProviderName,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixModelApplication(id, name string) model.Application {
	return model.Application{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
		BaseEntity:     &model.BaseEntity{ID: id},
	}
}

func fixGQLApplication(id, name string) graphql.Application {
	return graphql.Application{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func modelPtrsToWebhooks(in []*model.Webhook) []model.Webhook {
	if in == nil {
		return nil
	}
	webhookPtrs := []model.Webhook{}
	for i := range in {
		webhookPtrs = append(webhookPtrs, *in[i])
	}
	return webhookPtrs
}

func gqlPtrsToWebhooks(in []*graphql.Webhook) (webhookPtrs []graphql.Webhook) {
	for i := range in {
		webhookPtrs = append(webhookPtrs, *in[i])
	}
	return
}
