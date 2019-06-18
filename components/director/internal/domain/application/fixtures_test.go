package application_test

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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

func fixModelApplicationWithLabels(id, name string, labels map[string][]string) *model.Application {
	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: nil,
		Labels:      labels,
	}
}

func fixModelApplicationWithAnnotations(id, name string, annotations map[string]interface{}) *model.Application {
	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: nil,
		Annotations: annotations,
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
		Name:        name,
		Description: &description,
		Tenant:      "tenant",
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
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
		Name:        name,
		Description: &description,
		Tenant:      graphql.Tenant("tenant"),
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
		HealthCheckURL: &url,
	}
}

func fixModelApplicationInput(name, description string) model.ApplicationInput {
	url := "https://foo.bar"

	kind := "test"
	desc := "Sample"
	return model.ApplicationInput{
		Name:        name,
		Description: &description,
		Annotations: map[string]interface{}{
			"key": "value",
		},
		Labels: map[string][]string{
			"test": {"val", "val2"},
		},
		HealthCheckURL: &url,
		Webhooks: []*model.ApplicationWebhookInput{
			{
				URL: "webhook1.foo.bar",
			},
			{
				URL: "webhook2.foo.bar",
			},
		},
		Apis: []*model.APIDefinitionInput{
			{
				Name:      "api1",
				TargetURL: "foo.bar",
			},
			{
				Name:      "api2",
				TargetURL: "foo.bar2",
			},
		},
		EventAPIs: []*model.EventAPIDefinitionInput{
			{
				Name:        "event1",
				Description: &desc,
			},
			{
				Name:        "event2",
				Description: &desc,
			},
		},
		Documents: []*model.DocumentInput{
			{
				DisplayName: "doc1",
				Kind:        &kind,
			},
			{
				DisplayName: "doc2",
				Kind:        &kind,
			},
		},
	}
}

func fixGQLApplicationInput(name, description string) graphql.ApplicationInput {
	labels := graphql.Labels{
		"test": {"val", "val2"},
	}
	annotations := graphql.Annotations{
		"key": "value",
	}
	url := "https://foo.bar"
	kind := "test"
	desc := "Sample"
	return graphql.ApplicationInput{
		Name:           name,
		Description:    &description,
		Annotations:    &annotations,
		Labels:         &labels,
		HealthCheckURL: &url,
		Webhooks: []*graphql.ApplicationWebhookInput{
			{
				URL: "webhook1.foo.bar",
			},
			{
				URL: "webhook2.foo.bar",
			},
		},
		Apis: []*graphql.APIDefinitionInput{
			{
				Name:      "api1",
				TargetURL: "foo.bar",
			},
			{
				Name:      "api2",
				TargetURL: "foo.bar2",
			},
		},
		EventAPIs: []*graphql.EventAPIDefinitionInput{
			{
				Name:        "event1",
				Description: &desc,
			},
			{
				Name:        "event2",
				Description: &desc,
			},
		},
		Documents: []*graphql.DocumentInput{
			{
				DisplayName: "doc1",
				Kind:        &kind,
			},
			{
				DisplayName: "doc2",
				Kind:        &kind,
			},
		},
	}
}
