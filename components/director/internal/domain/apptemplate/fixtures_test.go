package apptemplate_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"

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
	testAppID          = "app-id"
	testConsumerID     = "consumer-id"
	testLabelInputKey  = "applicationType"
	region             = "region-1"
	certSubject        = "C=DE, L=local, O=SAP SE, OU=us10, OU=SAP Cloud Platform Clients, OU=1f538f34-30bf-4d3d-aeaa-02e69eef84ae, CN=non-compass"

	testWebhookID                               = "webhook-id-1"
	testName                                    = "bar"
	testAppName                                 = "app-name"
	updatedAppTemplateTestName                  = "new-app-template-test-name"
	testNameOtherSystemType                     = "Other System Type"
	testPageSize                                = 3
	testCursor                                  = ""
	appInputJSONString                          = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"displayName":"bar","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputJSONStringNoRegionString            = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"displayName":"{{region}}","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputJSONWithRegionString                = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"displayName":"bar","region":"{{region}}","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputJSONWithoutDisplayNameLabelString   = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputJSONWithAppTypeLabelString          = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"applicationType":"%s","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputJSONNonStringDisplayNameLabelString = `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"displayName":false,"test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`

	appInputGQLString = `{name: "foo",providerName: "compass",description: "Lorem ipsum",labels: {displayName:"bar",test:["val","val2"],},webhooks: [ {type: ,url: "webhook1.foo.bar",}, {type: ,url: "webhook2.foo.bar",} ],healthCheckURL: "https://foo.bar",integrationSystemID: "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii",}`
)

var (
	testUUID                       = "b3ea1977-582e-4d61-ae12-b3a837a3858e"
	testDescription                = "Lorem ipsum"
	testJSONPath                   = "$.test"
	testDifferentJSONPath          = "$.invalid.differentPath"
	testDescriptionWithPlaceholder = "Lorem ipsum {{test}}"
	testProviderName               = "provider-display-name"
	testURL                        = "http://valid.url"
	testError                      = errors.New("test error")
	testTableColumns               = []string{"id", "name", "description", "application_namespace", "application_input", "placeholders", "access_level", "created_at", "updated_at"}
	newTestLabels                  = map[string]interface{}{"label1": "test"}
	timestamp                      = time.Now()
	envConsumerSubjects            = []string{"C=DE, L=local, O=SAP SE, OU=Region, OU=SAP Cloud Platform Clients, OU=f8075207-1478-4a80-bd26-24a4785a2bfd, CN=compass"}
)

func fixModelApplicationTemplate(id, name string, webhooks []*model.Webhook) *model.ApplicationTemplate {
	desc := testDescription
	out := model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholders(),
		Webhooks:             modelPtrsToWebhooks(webhooks),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
		CreatedAt:            timestamp,
		UpdatedAt:            timestamp,
	}

	return &out
}

func fixModelApplicationTemplateWithRegionPlaceholders(id, name string, webhooks []*model.Webhook) *model.ApplicationTemplate {
	desc := testDescription
	out := model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholdersWithRegion(),
		Webhooks:             modelPtrsToWebhooks(webhooks),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
		CreatedAt:            timestamp,
		UpdatedAt:            timestamp,
	}

	return &out
}

func fixModelApplicationTemplateWithPlaceholdersPayload(id, name string, webhooks []*model.Webhook) *model.ApplicationTemplate {
	desc := testDescription
	out := model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholdersWithPayload(),
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

func fixModelAppTemplateWithAppInputJSONAndLabels(id, name, appInputJSON string, webhooks []*model.Webhook, labels map[string]interface{}) *model.ApplicationTemplate {
	out := fixModelApplicationTemplate(id, name, webhooks)
	out.Labels = labels
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
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput:     appInputGQLString,
		Placeholders:         fixGQLPlaceholders(),
		Webhooks:             gqlPtrsToWebhooks(webhooks),
		AccessLevel:          graphql.ApplicationTemplateAccessLevelGlobal,
		CreatedAt:            graphql.Timestamp(timestamp),
		UpdatedAt:            graphql.Timestamp(timestamp),
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
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholders(),
		Labels:               map[string]interface{}{"test": "test"},
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixRegionalModelAppTemplateInputWithProductLabel(name, appInputString, region string) *model.ApplicationTemplateInput {
	appTemplateInput := fixGlobalModelAppTemplateInputWithProductLabel(name, appInputString)
	appTemplateInput.Labels[RegionKey] = region

	return appTemplateInput
}

func fixGlobalModelAppTemplateInputWithProductLabel(name, appInputString string) *model.ApplicationTemplateInput {
	desc := testDescription

	return &model.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholdersWithRegion(),
		Labels:               map[string]interface{}{"test": "test", AppTemplateProductLabel: []interface{}{"role"}},
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixModelAppTemplateInputWithRegionLabelAndDifferentPlaceholders(name, appInputString, region string) *model.ApplicationTemplateInput {
	placeholderDesc := testDescription
	placeholderJSONPath := testDifferentJSONPath
	isOptional := false

	input := fixRegionalModelAppTemplateInputWithProductLabel(name, appInputString, region)
	input.Placeholders = []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
		{
			Name:        "region",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
	}

	return input
}

func fixModelAppTemplateInputWithOrdWebhook(name string, appInputString string) *model.ApplicationTemplateInput {
	desc := testDescription

	return &model.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholders(),
		Labels:               map[string]interface{}{"test": "test"},
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
		Webhooks: []*model.WebhookInput{
			{
				ID:   testID,
				Type: model.WebhookTypeOpenResourceDiscovery,
				URL:  str.Ptr("foourl"),
			},
		},
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
		ApplicationNamespace: str.Ptr("ns"),
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

func fixModelAppTemplateUpdateInputWithLabels(name string, appInputString string, labels map[string]interface{}) *model.ApplicationTemplateUpdateInput {
	out := fixModelAppTemplateUpdateInput(name, appInputString)
	out.Labels = labels

	return out
}

func fixGQLAppTemplateInput(name string) *graphql.ApplicationTemplateInput {
	desc := testDescription

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		Labels:       map[string]interface{}{"test": "test"},
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputWithRegionPlaceholder(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Labels:      map[string]interface{}{RegionKey: "{{region}}"},
		},
		Placeholders: fixGQLPlaceholderDefinitionWithRegionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGlobalGQLAppTemplateInputWithProductLabel(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
		Labels: graphql.Labels{
			AppTemplateProductLabel: []interface{}{"role"},
		},
	}
}

func fixRegionalGQLAppTemplateInputWithProductLabel(name, region string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Labels:      map[string]interface{}{RegionKey: "{{region}}"},
		},
		Placeholders: fixGQLPlaceholderDefinitionWithRegionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
		Labels: graphql.Labels{
			AppTemplateProductLabel: []interface{}{"role"},
			RegionKey:               region,
		},
	}
}

func fixRegionalGQLAppTemplateInputWithDifferentRegionPlaceholder(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder
	placeholderDesc := testDescription
	placeholderJSONPath := testDifferentJSONPath
	isOptional := false

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Labels:      map[string]interface{}{RegionKey: "{{region}}"},
		},
		Placeholders: []*graphql.PlaceholderDefinitionInput{
			{
				Name:        "test",
				Description: &placeholderDesc,
				JSONPath:    &placeholderJSONPath,
				Optional:    &isOptional,
			},
			{
				Name:        "region",
				Description: &placeholderDesc,
				JSONPath:    &placeholderJSONPath,
				Optional:    &isOptional,
			},
		},
		Labels: graphql.Labels{
			AppTemplateProductLabel:    []interface{}{"role"},
			selfregmanager.RegionLabel: "region",
		},
		AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputWithPlaceholder(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputWithPlaceholderAndProvider(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:         "foo",
			Description:  &desc,
			ProviderName: str.Ptr("SAP"),
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateInputInvalidAppInputURLTemplateMethod(name string) *graphql.ApplicationTemplateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Webhooks: []*graphql.WebhookInput{
				{
					Type:        graphql.WebhookTypeUnregisterApplication,
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
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		Labels:       map[string]interface{}{"label1": "test"},
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInputWithPlaceholder(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateUpdateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Labels: graphql.Labels{
				"displayName": "test",
			},
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInputWithPlaceholderAndProvider(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateUpdateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:         "foo",
			Description:  &desc,
			ProviderName: str.Ptr("SAP"),
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixGQLAppTemplateUpdateInputInvalidAppInput(name string) *graphql.ApplicationTemplateUpdateInput {
	desc := testDescriptionWithPlaceholder

	return &graphql.ApplicationTemplateUpdateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationNamespace: str.Ptr("ns"),
		ApplicationInput: &graphql.ApplicationJSONInput{
			Name:        "foo",
			Description: &desc,
			Webhooks: []*graphql.WebhookInput{
				{
					Type:        graphql.WebhookTypeUnregisterApplication,
					URLTemplate: str.Ptr(`{"path": "https://target.url", "method":"invalid method"}`),
				},
			},
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixEntityApplicationTemplate(t *testing.T, id, name string) *apptemplate.Entity {
	marshalledAppInput := `{"name":"foo","providerName":"compass","description":"Lorem ipsum","labels":{"displayName":"bar","test":["val","val2"]},"healthCheckURL":"https://foo.bar","webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`

	placeholders := fixModelPlaceholders()
	marshalledPlaceholders, err := json.Marshal(placeholders)
	require.NoError(t, err)

	return &apptemplate.Entity{
		ID:                   id,
		Name:                 name,
		Description:          repo.NewValidNullableString(testDescription),
		ApplicationNamespace: repo.NewValidNullableString("ns"),
		ApplicationInputJSON: marshalledAppInput,
		PlaceholdersJSON:     repo.NewValidNullableString(string(marshalledPlaceholders)),
		AccessLevel:          string(model.GlobalApplicationTemplateAccessLevel),
		CreatedAt:            timestamp,
		UpdatedAt:            timestamp,
	}
}

func fixModelPlaceholders() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := testDescription
	placeholderJSONPath := testJSONPath
	isOptional := false
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
	}
}

func fixModelPlaceholdersWithRegion() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := testDescription
	placeholderJSONPath := testJSONPath
	isOptional := false
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
		{
			Name:        "region",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
	}
}

func fixModelPlaceholdersWithPayload() []model.ApplicationTemplatePlaceholder {
	placeholderTestDesc := testDescription
	placeholderTestJSONPath := testJSONPath
	placeholderNameDesc := "Application Name placeholder"
	placeholderNameJSONPath := "$.name"
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderTestDesc,
			JSONPath:    &placeholderTestJSONPath,
		},
		{
			Name:        "name",
			Description: &placeholderNameDesc,
			JSONPath:    &placeholderNameJSONPath,
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
	placeholderJSONPath := testJSONPath
	isOptional := false
	return []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
	}
}

func fixGQLPlaceholderDefinitionWithRegionInput() []*graphql.PlaceholderDefinitionInput {
	placeholderDesc := testDescription
	placeholderJSONPath := testJSONPath
	isOptional := false
	return []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
		{
			Name:        "region",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
		},
	}
}

func fixGQLPlaceholders() []*graphql.PlaceholderDefinition {
	placeholderDesc := testDescription
	placeholderJSONPath := testJSONPath
	isOptional := false
	return []*graphql.PlaceholderDefinition{
		{
			Name:        "test",
			Description: &placeholderDesc,
			JSONPath:    &placeholderJSONPath,
			Optional:    &isOptional,
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
		Labels: map[string]interface{}{
			"key": "value",
		},
	}
}

func fixGQLApplicationFromTemplateWithManagedLabelInput(name, managedLabel string) graphql.ApplicationFromTemplateInput {
	return graphql.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*graphql.TemplateValueInput{
			{Placeholder: "a", Value: "b"},
			{Placeholder: "c", Value: "d"},
		},
		Labels: map[string]interface{}{
			"key":                       "value",
			application.ManagedLabelKey: managedLabel,
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
		Labels: map[string]interface{}{
			"key": "value",
		},
	}
}

func fixModelApplicationFromTemplateWithManagedLabelInput(name, managedLabel string) model.ApplicationFromTemplateInput {
	return model.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*model.ApplicationTemplateValueInput{
			{Placeholder: "a", Value: "b"},
			{Placeholder: "c", Value: "d"},
		},
		Labels: map[string]interface{}{
			"key":                       "value",
			application.ManagedLabelKey: managedLabel,
		},
	}
}

func fixGQLApplicationFromTemplateInputWithPlaceholderPayload(name string) graphql.ApplicationFromTemplateInput {
	placeholdersPayload := `{"name": "appName", "test":"testValue"}`
	return graphql.ApplicationFromTemplateInput{
		TemplateName:        name,
		PlaceholdersPayload: &placeholdersPayload,
	}
}

func fixModelApplicationFromTemplateInputWithPlaceholderPayload(name string) model.ApplicationFromTemplateInput {
	return model.ApplicationFromTemplateInput{
		TemplateName: name,
		Values: []*model.ApplicationTemplateValueInput{
			{Placeholder: "test", Value: "testValue"},
			{Placeholder: "name", Value: "appName"},
		},
	}
}

func fixAppTemplateCreateArgs(entity apptemplate.Entity) []driver.Value {
	return []driver.Value{entity.ID, entity.Name, entity.Description, entity.ApplicationNamespace, entity.ApplicationInputJSON, entity.PlaceholdersJSON, entity.AccessLevel, entity.CreatedAt, entity.UpdatedAt}
}

func fixSQLRows(entities []apptemplate.Entity) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, entity := range entities {
		out.AddRow(entity.ID, entity.Name, entity.Description, entity.ApplicationNamespace, entity.ApplicationInputJSON, entity.PlaceholdersJSON, entity.AccessLevel, entity.CreatedAt, entity.UpdatedAt)
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

func fixModelApplicationWithManagedLabelCreateInput(name, managedLabel string) model.ApplicationRegisterInput {
	return model.ApplicationRegisterInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
		Labels:         map[string]interface{}{application.ManagedLabelKey: managedLabel, "key": "value"},
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

func fixGQLApplicationCreateWithManagedTrueLabelInput(name, managedLabel string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:           name,
		ProviderName:   &testProviderName,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
		Labels:         map[string]interface{}{application.ManagedLabelKey: managedLabel},
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

func fixColumns() []string {
	return []string{"id", "name", "description", "application_namespace", "application_input", "placeholders", "access_level", "created_at", "updated_at"}
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
		{
			Type: graphql.WebhookTypeOpenResourceDiscovery,
			URL:  &testURL,
			Auth: nil,
		},
	}
}

func fixLabelInput(key string, value string, objectID string, objectType model.LabelableObject) *model.LabelInput {
	return &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectID:   objectID,
		ObjectType: objectType,
	}
}

func fixCertSubjectMappingModel(appTemplateID string) *model.CertSubjectMapping {
	return &model.CertSubjectMapping{
		ID:                 testUUID,
		Subject:            certSubject,
		ConsumerType:       string(consumer.ApplicationProvider),
		InternalConsumerID: &appTemplateID,
		TenantAccessLevels: []string{inputvalidation.GlobalAccessLevel},
	}
}
