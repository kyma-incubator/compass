package application_test

import (
	"database/sql"
	"encoding/json"
	"net/url"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

var (
	testURL            = "https://foo.bar"
	intSysID           = "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	providerName       = "provider name"
	fixedTimestamp     = time.Now()
	legacyConnectorURL = "url.com"
	systemNumber       = "123"
	appName            = "appName"
)

func stringPtr(s string) *string {
	return &s
}

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

func fixModelApplication(id, tenant, name, description string) *model.Application {
	return &model.Application{
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
		BaseEntity:  &model.BaseEntity{ID: id},
	}
}

func fixModelApplicationWithAllUpdatableFields(id, name, description, url string, baseURL *string, conditionStatus model.ApplicationStatusCondition, conditionTimestamp time.Time) *model.Application {
	return &model.Application{
		Status: &model.ApplicationStatus{
			Condition: conditionStatus,
			Timestamp: conditionTimestamp,
		},
		IntegrationSystemID: &intSysID,
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &url,
		ProviderName:        &providerName,
		BaseEntity:          &model.BaseEntity{ID: id},
		BaseURL:             baseURL,
	}
}

func fixGQLApplication(id, name, description string) *graphql.Application {
	return &graphql.Application{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

func fixDetailedModelApplication(t *testing.T, id, tenant, name, description string) *model.Application {
	appStatusTimestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.Application{
		ProviderName: &providerName,
		Name:         name,
		Description:  &description,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: appStatusTimestamp,
		},
		HealthCheckURL:      &testURL,
		SystemNumber:        &systemNumber,
		IntegrationSystemID: &intSysID,
		BaseURL:             str.Ptr("base_url"),
		Labels:              json.RawMessage("[]"),
		CorrelationIDs:      json.RawMessage("[]"),
		SystemStatus:        str.Ptr("reachable"),
		DocumentationLabels: json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixDetailedGQLApplication(t *testing.T, id, name, description string) *graphql.Application {
	appStatusTimestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.Application{
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
			Timestamp: graphql.Timestamp(appStatusTimestamp),
		},
		Name:                name,
		SystemNumber:        &systemNumber,
		Description:         &description,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        str.Ptr("provider name"),
		BaseURL:             str.Ptr("base_url"),
		SystemStatus:        str.Ptr("reachable"),
		BaseEntity: &graphql.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
		},
	}
}

func fixDetailedEntityApplication(t *testing.T, id, tenant, name, description string) *application.Entity {
	ts, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &application.Entity{
		Name:                name,
		ProviderName:        repo.NewNullableString(&providerName),
		Description:         repo.NewValidNullableString(description),
		StatusCondition:     string(model.ApplicationStatusConditionInitial),
		StatusTimestamp:     ts,
		SystemNumber:        repo.NewValidNullableString(systemNumber),
		HealthCheckURL:      repo.NewValidNullableString(testURL),
		IntegrationSystemID: repo.NewNullableString(&intSysID),
		BaseURL:             repo.NewValidNullableString("base_url"),
		Labels:              repo.NewValidNullableString("[]"),
		CorrelationIDs:      repo.NewValidNullableString("[]"),
		SystemStatus:        repo.NewValidNullableString("reachable"),
		DocumentationLabels: repo.NewValidNullableString("[]"),
		BaseEntity: &repo.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     sql.NullString{},
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixModelApplicationRegisterInput(name, description string) model.ApplicationRegisterInput {
	desc := "Sample"
	kind := "test"
	return model.ApplicationRegisterInput{
		Name:        name,
		Description: &description,
		Labels: map[string]interface{}{
			"test": []string{"val", "val2"},
		},
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		Webhooks: []*model.WebhookInput{
			{URL: stringPtr("webhook1.foo.bar")},
			{URL: stringPtr("webhook2.foo.bar")},
		},
		Bundles: []*model.BundleCreateInput{
			{
				Name: "foo",
				APIDefinitions: []*model.APIDefinitionInput{
					{Name: "api1", TargetURLs: api.ConvertTargetURLToJSONArray("foo.bar")},
					{Name: "api2", TargetURLs: api.ConvertTargetURLToJSONArray("foo.bar2")},
				},
				EventDefinitions: []*model.EventDefinitionInput{
					{Name: "event1", Description: &desc},
					{Name: "event2", Description: &desc},
				},
				Documents: []*model.DocumentInput{
					{DisplayName: "doc1", Kind: &kind},
					{DisplayName: "doc2", Kind: &kind},
				},
			},
		},
	}
}

func fixModelApplicationUpdateInput(name, description, healthCheckUrl, baseUrl string, statusCondition model.ApplicationStatusCondition) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Description:         &description,
		HealthCheckURL:      &healthCheckUrl,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		StatusCondition:     &statusCondition,
		BaseURL: 		     &baseUrl,
	}
}

func fixModelApplicationUpdateInputStatus(statusCondition model.ApplicationStatusCondition) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		StatusCondition: &statusCondition,
	}
}

func fixGQLApplicationRegisterInput(name, description string) graphql.ApplicationRegisterInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}
	kind := "test"
	desc := "Sample"
	return graphql.ApplicationRegisterInput{
		Name:                name,
		Description:         &description,
		Labels:              labels,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		Webhooks: []*graphql.WebhookInput{
			{URL: stringPtr("webhook1.foo.bar")},
			{URL: stringPtr("webhook2.foo.bar")},
		},
		Bundles: []*graphql.BundleCreateInput{
			{
				Name: "foo",
				APIDefinitions: []*graphql.APIDefinitionInput{
					{Name: "api1", TargetURL: "foo.bar"},
					{Name: "api2", TargetURL: "foo.bar2"},
				},
				EventDefinitions: []*graphql.EventDefinitionInput{
					{Name: "event1", Description: &desc},
					{Name: "event2", Description: &desc},
				},
				Documents: []*graphql.DocumentInput{
					{DisplayName: "doc1", Kind: &kind},
					{DisplayName: "doc2", Kind: &kind},
				},
			},
		},
	}
}

func fixGQLApplicationUpdateInput(name, description, healthCheckUrl, baseUrl string, statusCondition graphql.ApplicationStatusCondition) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:         &description,
		HealthCheckURL:      &healthCheckUrl,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		StatusCondition:     &statusCondition,
		BaseURL: 			 &baseUrl,
	}
}

var (
	docKind  = "fookind"
	docTitle = "footitle"
	docData  = "foodata"
	docCLOB  = graphql.CLOB(docData)
)

func fixModelDocument(bundleID, id string) *model.Document {
	return &model.Document{
		BundleID:   bundleID,
		Title:      docTitle,
		Format:     model.DocumentFormatMarkdown,
		Kind:       &docKind,
		Data:       &docData,
		BaseEntity: &model.BaseEntity{ID: id},
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
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
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
		ObjectID:   appID,
		ObjectType: model.ApplicationWebhookReference,
		ID:         id,
		Type:       model.WebhookTypeConfigurationChanged,
		URL:        stringPtr("foourl"),
		Auth:       &model.Auth{},
	}
}

func fixGQLWebhook(id string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:   id,
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  stringPtr("foourl"),
		Auth: &graphql.Auth{},
	}
}

func fixEventAPIDefinitionPage(eventAPIDefinitions []*model.EventDefinition) *model.EventDefinitionPage {
	return &model.EventDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

func fixGQLEventDefinitionPage(eventAPIDefinitions []*graphql.EventDefinition) *graphql.EventDefinitionPage {
	return &graphql.EventDefinitionPage{
		Data: eventAPIDefinitions,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(eventAPIDefinitions),
	}
}

func fixModelEventAPIDefinition(id string, appID, name, description string, group string) *model.EventDefinition {
	return &model.EventDefinition{
		Name:        name,
		Description: &description,
		Group:       &group,
		BaseEntity:  &model.BaseEntity{ID: id},
	}
}
func fixMinModelEventAPIDefinition(id, placeholder string) *model.EventDefinition {
	return &model.EventDefinition{Name: placeholder, BaseEntity: &model.BaseEntity{ID: id}}
}
func fixGQLEventDefinition(id string, appID, bundleID string, name, description string, group string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		BundleID:    bundleID,
		Name:        name,
		Description: &description,
		Group:       &group,
	}
}

func fixFetchRequest(url string, objectType model.FetchRequestReferenceObjectType, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     "foo",
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: objectType,
		ObjectID:   "foo",
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

func fixModelApplicationEventingConfiguration(t *testing.T, rawURL string) *model.ApplicationEventingConfiguration {
	validURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	require.NotNil(t, validURL)
	return &model.ApplicationEventingConfiguration{
		EventingConfiguration: model.EventingConfiguration{
			DefaultURL: *validURL,
		},
	}
}

func fixGQLApplicationEventingConfiguration(url string) *graphql.ApplicationEventingConfiguration {
	return &graphql.ApplicationEventingConfiguration{
		DefaultURL: url,
	}
}

func fixModelBundle(id, tenantID, appID, name, description string) *model.Bundle {
	return &model.Bundle{
		ApplicationID:                  appID,
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: nil,
		DefaultInstanceAuth:            nil,
		BaseEntity:                     &model.BaseEntity{ID: id},
	}
}

func fixGQLBundle(id, appID, name, description string) *graphql.Bundle {
	return &graphql.Bundle{
		BaseEntity: &graphql.BaseEntity{
			ID: id,
		},
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: nil,
		DefaultInstanceAuth:            nil,
	}
}

func fixGQLBundlePage(bundles []*graphql.Bundle) *graphql.BundlePage {
	return &graphql.BundlePage{
		Data: bundles,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(bundles),
	}
}

func fixBundlePage(bundles []*model.Bundle) *model.BundlePage {
	return &model.BundlePage{
		Data: bundles,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(bundles),
	}
}

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}

func fixAppColumns() []string {
	return []string{"id", "app_template_id", "system_number", "name", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "documentation_labels"}
}
