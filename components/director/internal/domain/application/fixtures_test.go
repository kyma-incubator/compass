package application_test

import (
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

func fixModelApplicationWithAllUpdatableFields(id, tenant, name, description, url string) *model.Application {
	return &model.Application{
		ID:     id,
		Tenant: tenant,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
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
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &model.Application{
		ID: id,
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: time,
		},
		Name:                name,
		Description:         &description,
		Tenant:              tenant,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
	}
}

func fixDetailedGQLApplication(t *testing.T, id, name, description string) *graphql.Application {
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	return &graphql.Application{
		ID: id,
		Status: &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
			Timestamp: graphql.Timestamp(time),
		},
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &testURL,
		IntegrationSystemID: &intSysID,
		ProviderName:        str.Ptr("provider name"),
	}
}

func fixDetailedEntityApplication(t *testing.T, id, tenant, name, description string) *application.Entity {
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
	}
}

func fixModelApplicationUpdateInput(name, description, url string) model.ApplicationUpdateInput {
	return model.ApplicationUpdateInput{
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &url,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
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
	}
}

func fixGQLApplicationUpdateInput(name, description, url string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name:                name,
		Description:         &description,
		HealthCheckURL:      &url,
		IntegrationSystemID: &intSysID,
		ProviderName:        &providerName,
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

func fixModelAPIDefinition(id string, appId *string, name, description string, group string) *model.APIDefinition {
	return &model.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}

func fixGQLAPIDefinition(id string, appId *string, name, description string, group string) *graphql.APIDefinition {
	return &graphql.APIDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
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

func fixModelEventAPIDefinition(id, appId, name, description string, group string) *model.EventDefinition {
	return &model.EventDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
	}
}
func fixMinModelEventAPIDefinition(id, placeholder string) *model.EventDefinition {
	return &model.EventDefinition{ID: id, Tenant: "ttttttttt-tttt-tttt-tttt-tttttttttttt",
		ApplicationID: "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: placeholder}
}
func fixGQLEventDefinition(id, appId, name, description string, group string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		ID:            id,
		ApplicationID: appId,
		Name:          name,
		Description:   &description,
		Group:         &group,
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

func fixModelPackage(id, tenantID, appId, name, description string) *model.Package {
	return &model.Package{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  appId,
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: nil,
		DefaultInstanceAuth:            nil,
	}
}

func fixGQLPackage(id, appId, name, description string) *graphql.Package {
	return &graphql.Package{
		ID:                             id,
		Name:                           name,
		Description:                    &description,
		InstanceAuthRequestInputSchema: nil,
		DefaultInstanceAuth:            nil,
	}
}

func fixGQLPackagePage(packages []*graphql.Package) *graphql.PackagePage {
	return &graphql.PackagePage{
		Data: packages,
		PageInfo: &graphql.PageInfo{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(packages),
	}
}

func fixPackagePage(packages []*model.Package) *model.PackagePage {
	return &model.PackagePage{
		Data: packages,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(packages),
	}
}
