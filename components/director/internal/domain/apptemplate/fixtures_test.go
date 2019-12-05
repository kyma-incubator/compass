package apptemplate_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
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
	testTenant         = "tnt"
	testID             = "foo"
	testName           = "bar"
	testPageSize       = 3
	testCursor         = ""
	appInputJSONString = `{"Name":"foo","Description":"Lorem ipsum","Labels":{"test":["val","val2"]},"HealthCheckURL":"https://foo.bar","Webhooks":[{"Type":"","URL":"webhook1.foo.bar","Auth":null},{"Type":"","URL":"webhook2.foo.bar","Auth":null}],"Apis":[{"Name":"api1","Description":null,"TargetURL":"foo.bar","Group":null,"Spec":null,"Version":null,"DefaultAuth":null},{"Name":"api2","Description":null,"TargetURL":"foo.bar2","Group":null,"Spec":null,"Version":null,"DefaultAuth":null}],"EventAPIs":[{"Name":"event1","Description":"Sample","Spec":{"Data":"data","Type":"ASYNC_API","Format":"JSON"},"Group":null,"Version":null},{"Name":"event2","Description":"Sample","Spec":{"Data":"data2","Type":"ASYNC_API","Format":"JSON"},"Group":null,"Version":null}],"Documents":[{"Title":"","DisplayName":"doc1","Description":"","Format":"","Kind":"test","Data":null,"FetchRequest":null},{"Title":"","DisplayName":"doc2","Description":"","Format":"","Kind":"test","Data":null,"FetchRequest":null}],"IntegrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`
	appInputGQLString  = `{name: "foo",description: "Lorem ipsum",labels: {test: ["val","val2" ],},webhooks: [ {type: ,url: "webhook1.foo.bar",}, {type: ,url: "webhook2.foo.bar",} ],healthCheckURL: "https://foo.bar",apis: [ {name: "api1",targetURL: "foo.bar",}, {name: "api2",targetURL: "foo.bar2",}],eventAPIs: [ {name: "event1",description: "Sample",spec: {data: "data",eventSpecType: ,format: JSON,},}, {name: "event2",description: "Sample",spec: {data: "data2",eventSpecType: ,format: JSON,},}], documents: [{title: "",displayName: "doc1",description: "",format: ,kind: "test",},{title: "",displayName: "doc2",description: "",format: ,kind: "test",} ],integrationSystemID: "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii",}`
)

var (
	testDescription  = "Lorem ipsum"
	testURL          = "http://valid.url"
	testError        = errors.New("test error")
	testTableColumns = []string{"id", "name", "description", "application_input", "placeholders", "access_level"}
)

func fixModelAppTemplate(id, name string) *model.ApplicationTemplate {
	desc := testDescription

	return &model.ApplicationTemplate{
		ID:                   id,
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputJSONString,
		Placeholders:         fixModelPlaceholders(),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixGQLAppTemplate(id, name string) *graphql.ApplicationTemplate {
	desc := testDescription

	return &graphql.ApplicationTemplate{
		ID:               id,
		Name:             name,
		Description:      &desc,
		ApplicationInput: appInputGQLString,
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

func fixModelAppTemplateInput(name string, appInputString string) *model.ApplicationTemplateInput {
	desc := testDescription

	return &model.ApplicationTemplateInput{
		Name:                 name,
		Description:          &desc,
		ApplicationInputJSON: appInputString,
		Placeholders:         fixModelPlaceholders(),
		AccessLevel:          model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixGQLAppTemplateInput(name string) *graphql.ApplicationTemplateInput {
	desc := testDescription

	return &graphql.ApplicationTemplateInput{
		Name:        name,
		Description: &desc,
		ApplicationInput: &graphql.ApplicationCreateInput{
			Name:        "foo",
			Description: &desc,
		},
		Placeholders: fixGQLPlaceholderDefinitionInput(),
		AccessLevel:  graphql.ApplicationTemplateAccessLevelGlobal,
	}
}

func fixEntityAppTemplate(t *testing.T, id, name string) *apptemplate.Entity {
	marshalledAppInput := `{"Name":"foo","Description":"Lorem ipsum","Labels":{"test":["val","val2"]},"HealthCheckURL":"https://foo.bar","Webhooks":[{"Type":"","URL":"webhook1.foo.bar","Auth":null},{"Type":"","URL":"webhook2.foo.bar","Auth":null}],"Apis":[{"Name":"api1","Description":null,"TargetURL":"foo.bar","Group":null,"Spec":null,"Version":null,"DefaultAuth":null},{"Name":"api2","Description":null,"TargetURL":"foo.bar2","Group":null,"Spec":null,"Version":null,"DefaultAuth":null}],"EventAPIs":[{"Name":"event1","Description":"Sample","Spec":{"Data":"data","Type":"ASYNC_API","Format":"JSON"},"Group":null,"Version":null},{"Name":"event2","Description":"Sample","Spec":{"Data":"data2","Type":"ASYNC_API","Format":"JSON"},"Group":null,"Version":null}],"Documents":[{"Title":"","DisplayName":"doc1","Description":"","Format":"","Kind":"test","Data":null,"FetchRequest":null},{"Title":"","DisplayName":"doc2","Description":"","Format":"","Kind":"test","Data":null,"FetchRequest":null}],"IntegrationSystemID":"iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"}`

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
	return fmt.Sprintf(`{"Name": "%s", "Description": "%s", "HealthCheckURL": "%s"}`, name, testDescription, testURL)
}

func fixModelApplicationCreateInput(name string) model.ApplicationCreateInput {
	return model.ApplicationCreateInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixGQLApplicationCreateInput(name string) graphql.ApplicationCreateInput {
	return graphql.ApplicationCreateInput{
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixModelApplication(id, name string) model.Application {
	return model.Application{
		ID:             id,
		Tenant:         testTenant,
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}

func fixGQLApplication(id, name string) graphql.Application {
	return graphql.Application{
		ID:             id,
		Name:           name,
		Description:    &testDescription,
		HealthCheckURL: &testURL,
	}
}
