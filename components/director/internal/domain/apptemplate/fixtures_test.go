package apptemplate_test

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testID       = "foo"
	testName     = "bar"
	testDescription = "Lorem ipsum"
	testPageSize = 3
	testCursor   = ""
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
		Placeholders:     fixPlaceholders(),
		AccessLevel:      model.GlobalApplicationTemplateAccessLevel,
	}
}

func fixEntityAppTemplate(t *testing.T, id, name string) *apptemplate.Entity {
	appInput := fixModelApplicationCreateInput()
	marshalledAppInput, err := json.Marshal(appInput)
	require.NoError(t, err)

	placeholders := fixPlaceholders()
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
	testURL  := "https://foo.bar"
	intSysID := "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	return &model.ApplicationCreateInput{
		Name:        "foo",
		Description: &desc,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
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

func fixPlaceholders() []model.ApplicationTemplatePlaceholder {
	placeholderDesc := "Foo bar"
	return []model.ApplicationTemplatePlaceholder{
		{
			Name: "test",
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
