package application_test

import (
	"database/sql"
	"net/url"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/stretchr/testify/require"
)

var (
	testURL      = "https://foo.bar"
	intSysID     = "iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii"
	providerName = "provider name"
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

func fixModelApplication(id, tenant, name, description string) *model.Application {
	return &model.Application{
		ID:     id,
		Tenant: tenant,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
		},
		Name:        name,
		Description: &description,
	}
}

func fixModelApplicationWithAllUpdatableFields(id, tenant, name, description, url string, conditionStatus model.ApplicationStatusCondition, conditionTimestamp time.Time) *model.Application {
	return &model.Application{
		ID:     id,
		Tenant: tenant,
		Status: &model.ApplicationStatus{
			Condition: conditionStatus,
			Timestamp: conditionTimestamp,
		},
		IntegrationSystemID: &intSysID,
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &url,
		ProviderName:        &providerName,
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

func fixDetailedModelApplication(t *testing.T, id, tenant, name, description string) *model.Application {
	return fixDetailedModelApplicationWithTimestamp(t, id, tenant, name, description, time.Now())
}

func fixDetailedModelApplicationWithTimestamp(t *testing.T, id, tenant, name, description string, createdAt time.Time) *model.Application {
	appStatusTimestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: appStatusTimestamp,
		},
		Name:                name,
		Description:         &description,
		Tenant:              tenant,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		Ready:               true,
		Error:               nil,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		DeletedAt:           time.Time{},
	}
}

func fixDetailedGQLApplication(t *testing.T, id, name, description string) *graphql.Application {
	return fixDetailedGQLApplicationWithTimestamp(t, id, name, description, createdAt)
}

func fixDetailedGQLApplicationWithTimestamp(t *testing.T, id, name, description string, createdAt time.Time) *graphql.Application {
	appStatusTimestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.Application{
		ID: id,
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
			Timestamp: graphql.Timestamp(appStatusTimestamp),
		},
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        str.Ptr("provider name"),
		Ready:               true,
		Error:               nil,
		CreatedAt:           graphql.Timestamp(createdAt),
		UpdatedAt:           graphql.Timestamp(createdAt),
		DeletedAt:           graphql.Timestamp(time.Time{}),
	}
}

func fixDetailedEntityApplication(t *testing.T, id, tenant, name, description string) *application.Entity {
	return fixDetailedEntityApplicationWithTimestamp(t, id, tenant, name, description, time.Now())
}

func fixDetailedEntityApplicationWithTimestamp(t *testing.T, id, tenant, name, description string, createdAt time.Time) *application.Entity {
	ts, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &application.Entity{
		ID:                  id,
		TenantID:            tenant,
		Name:                name,
		Description:         repo.NewValidNullableString(description),
		StatusCondition:     string(model.ApplicationStatusConditionInitial),
		StatusTimestamp:     ts,
		HealthCheckURL:      repo.NewValidNullableString(testURL),
		IntegrationSystemID: repo.NewNullableString(&intSysID),
		ProviderName:        repo.NewNullableString(&providerName),
		Ready:               true,
		Error:               sql.NullString{},
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		DeletedAt:           time.Time{},
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
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
		},
		Bundles: []*model.BundleCreateInput{
			{
				Name: "foo",
				APIDefinitions: []*model.APIDefinitionInput{
					{Name: "api1", TargetURL: "foo.bar"},
					{Name: "api2", TargetURL: "foo.bar2"},
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

func fixModelApplicationUpdateInput(name, description, url string, statusCondition model.ApplicationStatusCondition) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Description:         &description,
		HealthCheckURL:      &url,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		StatusCondition:     &statusCondition,
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
		Labels:              &labels,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		Webhooks: []*graphql.WebhookInput{
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
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

func fixGQLApplicationUpdateInput(name, description, url string, statusCondition graphql.ApplicationStatusCondition) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:         &description,
		HealthCheckURL:      &url,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
		StatusCondition:     &statusCondition,
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
		BundleID: bundleID,
		ID:       id,
		Title:    docTitle,
		Format:   model.DocumentFormatMarkdown,
		Kind:     &docKind,
		Data:     &docData,
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
		ApplicationID: &appID,
		ID:            id,
		Type:          model.WebhookTypeConfigurationChanged,
		URL:           "foourl",
		Auth:          &model.Auth{},
	}
}

func fixGQLWebhook(id string) *graphql.Webhook {
	return &graphql.Webhook{
		ID:   id,
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  "foourl",
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

func fixModelEventAPIDefinition(id string, appId, bundleID string, name, description string, group string) *model.EventDefinition {
	return &model.EventDefinition{
		ID:          id,
		BundleID:    bundleID,
		Name:        name,
		Description: &description,
		Group:       &group,
	}
}
func fixMinModelEventAPIDefinition(id, placeholder string) *model.EventDefinition {
	return &model.EventDefinition{ID: id, Tenant: "ttttttttt-tttt-tttt-tttt-tttttttttttt",
		BundleID: "ppppppppp-pppp-pppp-pppp-pppppppppppp", Name: placeholder}
}
func fixGQLEventDefinition(id string, appId, bundleID string, name, description string, group string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		ID:          id,
		BundleID:    bundleID,
		Name:        name,
		Description: &description,
		Group:       &group,
	}
}

func fixFetchRequest(url string, objectType model.FetchRequestReferenceObjectType, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     "foo",
		Tenant: "tenant",
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

func fixModelBundle(id, tenantID, appId, name, description string) *model.Bundle {
	return &model.Bundle{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  appId,
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: nil,
		DefaultInstanceAuth:            nil,
	}
}

func fixGQLBundle(id, appId, name, description string) *graphql.Bundle {
	return &graphql.Bundle{
		ID:                             id,
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
