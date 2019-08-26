package application_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func fixApplicationPage(applications []*model.Application) *model.ApplicationPage {
	return &model.ApplicationPage{
		Data: applications,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(applications),
	}
}

func fixGQLApplicationPage(applications []*graphql.Application) *graphql.ApplicationPage {
	return &graphql.ApplicationPage{
		Data: applications,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(applications),
	}
}

func fixModelApplication(id, name, description string) *model.Application {
	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

func fixGQLApplication(id, name, description string) *graphql.Application {
	return &graphql.Application{
		ID: id,
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

func fixDetailedModelApplication(t *testing.T, id, name, description string) *model.Application {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	url := "https://foo.bar"
	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: time,
		},
		Name:           name,
		Description:    &description,
		Tenant:         "tenant",
		HealthCheckURL: &url,
	}
}

func fixDetailedGQLApplication(t *testing.T, id, name, description string) *graphql.Application {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)
	url := "https://foo.bar"

	return &graphql.Application{
		ID: id,
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
			Timestamp: graphql.Timestamp(time),
		},
		Name:           name,
		Description:    &description,
		HealthCheckURL: &url,
	}
}

func fixModelApplicationInput(name, description string) model.ApplicationInput {
	url := "https://foo.bar"

	desc := "Sample"
	kind := "test"
	return model.ApplicationInput{
		Name:        name,
		Description: &description,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
		},
		HealthCheckURL: &url,
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

func fixGQLApplicationInput(name, description string) graphql.ApplicationInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}
	url := "https://foo.bar"
	kind := "test"
	desc := "Sample"
	return graphql.ApplicationInput{
		Name:           name,
		Description:    &description,
		Labels:         &labels,
		HealthCheckURL: &url,
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

var (
	docKind  = "fookind"
	docTitle = "footitle"
	docData  = "foodata"
	docCLOB  = graphql.CLOB(docData)
)

func fixModelDocument(applicationID, id string) *model.Document {
	return &model.Document{
		ApplicationID: applicationID,
		ID:            id,
		Title:         docTitle,
		Format:        model.DocumentFormatMarkdown,
		Kind:          &docKind,
		Data:          &docData,
	}
}

func fixModelDocumentPage(documents []*model.Document) *model.DocumentPage {
	return &model.DocumentPage{
		Data: documents,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(documents),
	}
}

func fixGQLDocument(id string) *graphql.Document {
	return &graphql.Document{
		ID:     id,
		Title:  docTitle,
		Format: graphql.DocumentFormatMarkdown,
		Kind:   &docKind,
		Data:   &docCLOB,
	}
}

func fixGQLDocumentPage(documents []*graphql.Document) *graphql.DocumentPage {
	return &graphql.DocumentPage{
		Data: documents,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(documents),
	}
}

func fixModelWebhook(appID, id string) *model.Webhook {
	return &model.Webhook{
		ApplicationID: appID,
		ID:            id,
		Type:          model.WebhookTypeConfigurationChanged,
		URL:           "foourl",
		Auth:          &model.Auth{},
	}
}

func fixGQLWebhook(id string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:   id,
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  "foourl",
		Auth: &graphql.Auth{},
	}
}

func fixModelWebhookInput() *model.WebhookInput {
	return &model.WebhookInput{
		Type: model.WebhookTypeConfigurationChanged,
		URL:  "foourl",
		Auth: &model.AuthInput{},
	}
}

func fixGQLWebhookInput() *graphql.WebhookInput {
	return &graphql.WebhookInput{
		Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		URL:  "foourl",
		Auth: &graphql.AuthInput{},
	}
}

func fixAPIDefinitionPage(apiDefinitions []*model.APIDefinition) *model.APIDefinitionPage {
	return &model.APIDefinitionPage{
		Data: apiDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(apiDefinitions),
	}
}

func fixGQLAPIDefinitionPage(apiDefinitions []*graphql.APIDefinition) *graphql.APIDefinitionPage {
	return &graphql.APIDefinitionPage{
		Data: apiDefinitions,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(apiDefinitions),
	}
}

func fixModelAPIDefinition(id, appId, name, description string, group string) *model.APIDefinition {
	return &model.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}

func fixGQLAPIDefinition(id, appId, name, description string, group string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}
func fixEventAPIDefinitionPage(eventAPIDefinitions []*model.EventAPIDefinition) *model.EventAPIDefinitionPage {
	return &model.EventAPIDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

func fixGQLEventAPIDefinitionPage(eventAPIDefinitions []*graphql.EventAPIDefinition) *graphql.EventAPIDefinitionPage {
	return &graphql.EventAPIDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

func fixModelEventAPIDefinition(id, appId, name, description string, group string) *model.EventAPIDefinition {
	return &model.EventAPIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}

func fixGQLEventAPIDefinition(id, appId, name, description string, group string) *graphql.EventAPIDefinition {
	return &graphql.EventAPIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}
