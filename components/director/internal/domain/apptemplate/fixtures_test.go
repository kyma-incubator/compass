package apptemplate_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"
)

const (
	testTenant      = "tnt"
	testID          = "foo"
	testName        = "bar"
	testDescription = "Lorem ipsum"
	testPageSize    = 3
	testCursor      = ""
)

var (
	testError        = errors.New("test error")
	testTableColumns = []string{"id", "name", "description", "application_input", "placeholders", "access_level"}
)

func fixModelAppTemplate(id, name string) *model.ApplicationTemplate {
	desc := testDescription

	return &model.ApplicationTemplate{
		ID:               id,
		Name:             name,
		Description:      &desc,
		ApplicationInput: fixModelApplicationCreateInput(),
		Placeholders:     fixModelPlaceholders(),
		AccessLevel:      model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixGQLAppTemplate(id, name string) *graphql.ApplicationTemplate {
	desc := testDescription

	return &graphql.ApplicationTemplate{
		ID:               id,
		Name:             name,
		Description:      &desc,
		ApplicationInput: fixGQLApplicationCreateInputString(),
		Placeholders:     fixGQLPlaceholders(),
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

func fixModelAppTemplateInput(name string) *model.ApplicationTemplateInput {
	desc := testDescription

	return &model.ApplicationTemplateInput{
		Name:             name,
		Description:      &desc,
		ApplicationInput: fixModelApplicationCreateInput(),
		Placeholders:     fixModelPlaceholders(),
		AccessLevel:      model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixGQLAppTemplateInput(name string) *graphql.ApplicationTemplateInput {
	desc := testDescription

	return &graphql.ApplicationTemplateInput{
		Name:             name,
		Description:      &desc,
		ApplicationInput: fixGQLApplicationCreateInput(),
		Placeholders:     fixGQLPlaceholderDefinitionInput(),
		AccessLevel:      graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixEntityAppTemplate(t *testing.T, id, name string) *apptemplate.Entity {
	appInput := fixModelApplicationCreateInput()
	marshalledAppInput, err := json.Marshal(appInput)
	require.NoError(t, err)

	placeholders := fixModelPlaceholders()
	marshalledPlaceholders, err := json.Marshal(placeholders)
	require.NoError(t, err)

	return &apptemplate.Entity{
		ID:               id,
		Name:             name,
		Description:      repo.NewValidNullableString(testDescription),
		ApplicationInput: string(marshalledAppInput),
		Placeholders:     repo.NewValidNullableString(string(marshalledPlaceholders)),
		AccessLevel:      string(model.GlobalApplicationTemplateAccessLevel),
	}
}

func fixModelApplicationCreateInput() *model.ApplicationCreateInput {
	desc := "Sample"
	kind := "test"
	testURL := "https://foo.bar"
	intSysID := "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	return &model.ApplicationCreateInput{
		Name:        "foo",
		Description: &desc,
		Labels: map[string]interface{}{
			"test": []interface{}{"val", "val2"},
		},
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		Webhooks: []*model.WebhookInput{
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
		},
		Apis: []*model.APIDefinitionInput{
			{Name: "api1", TargetURL: "foo.bar"},
			{Name: "api2", TargetURL: "foo.bar2"},
		},
		EventAPIs: []*model.EventAPIDefinitionInput{
			{Name: "event1", Description: &desc},
			{Name: "event2", Description: &desc},
		},
		Documents: []*model.DocumentInput{
			{DisplayName: "doc1", Kind: &kind},
			{DisplayName: "doc2", Kind: &kind},
		},
	}
}

func fixGQLApplicationCreateInput() *graphql.ApplicationCreateInput {
	desc := "Sample"
	kind := "test"
	testURL := "https://foo.bar"
	intSysID := "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	return &graphql.ApplicationCreateInput{
		Name:        "foo",
		Description: &desc,
		Labels: &graphql.Labels{
			"test": []interface{}{"val", "val2"},
		},
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		Webhooks: []*graphql.WebhookInput{
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
		},
		Apis: []*graphql.APIDefinitionInput{
			{Name: "api1", TargetURL: "foo.bar"},
			{Name: "api2", TargetURL: "foo.bar2"},
		},
		EventAPIs: []*graphql.EventAPIDefinitionInput{
			{Name: "event1", Description: &desc},
			{Name: "event2", Description: &desc},
		},
		Documents: []*graphql.DocumentInput{
			{DisplayName: "doc1", Kind: &kind},
			{DisplayName: "doc2", Kind: &kind},
		},
	}
}

func fixGQLApplicationCreateInputString() string {
	return `
	{
		"name":"foo",
		"description":"Sample",
		"labels":{"test":["val","val2"]},
		"webhooks":[{"type":"","url":"webhook1.foo.bar","auth":null},{"type":"","url":"webhook2.foo.bar","auth":null}],
		"healthCheckURL":"https://foo.bar",
		"apis":[{"name":"api1","description":null,"targetURL":"foo.bar","group":null,"spec":null,"version":null,"defaultAuth":null},{"name":"api2","description":null,"targetURL":"foo.bar2","group":null,"spec":null,"version":null,"defaultAuth":null}],
		"eventAPIs":[{"name":"event1","description":"Sample","spec":null,"group":null,"version":null},{"name":"event2","description":"Sample","spec":null,"group":null,"version":null}],
		"documents":[{"title":"","displayName":"doc1","description":"","format":"","kind":"test","data":null,"fetchRequest":null},{"title":"","displayName":"doc2","description":"","format":"","kind":"test","data":null,"fetchRequest":null}],
		"integrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	}
`
}

func fixModelPlaceholders() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := "Foo bar"
	return []model.ApplicationTemplatePlaceholder{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixGQLPlaceholderDefinitionInput() []*graphql.PlaceholderDefinitionInput {
	placeholderDesc := "Foo bar"
	return []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixGQLPlaceholders() []*graphql.PlaceholderDefinition {
	placeholderDesc := "Foo bar"
	return []*graphql.PlaceholderDefinition{
		{
			Name:        "test",
			Description: &placeholderDesc,
		},
	}
}

func fixAppTemplateCreateArgs(entity apptemplate.Entity) []driver.Value {
	return []driver.Value{entity.ID, entity.Name, entity.Description, entity.ApplicationInput, entity.Placeholders, entity.AccessLevel}
}

func fixSQLRows(entities []apptemplate.Entity) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, entity := range entities {
		out.AddRow(entity.ID, entity.Name, entity.Description, entity.ApplicationInput, entity.Placeholders, entity.AccessLevel)
	}
	return out
}
