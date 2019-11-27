package eventapi_test

import (
	"database/sql/driver"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	eventAPIID = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	appID      = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantID   = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
)

func fixMinModelEventAPIDefinition(id, placeholder string) *model.EventAPIDefinition {
	return &model.EventAPIDefinition{ID: id, Tenant: tenantID, ApplicationID: appID, Name: placeholder}
}

func fixGQLEventAPIDefinition(id, placeholder string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          placeholder,
	}
}

func fixFullModelEventAPIDefinition(id, placeholder string) model.EventAPIDefinition {
	spec := &model.EventAPISpec{
		Data:   str.Ptr("data"),
		Format: model.SpecFormatJSON,
		Type:   model.EventAPISpecTypeAsyncAPI,
	}
	v := fixVersionModel()

	return model.EventAPIDefinition{
		ID:            id,
		ApplicationID: appID,
		Tenant:        tenantID,
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Group:         str.Ptr("group_" + placeholder),
		Spec:          spec,
		Version:       &v,
	}
}

func fixDetailedGQLEventAPIDefinition(id, placeholder string) *graphql.EventDefinition {
	data := graphql.CLOB("data")
	format := graphql.SpecFormatJSON

	spec := &graphql.EventSpec{
		Data:         &data,
		Format:       format,
		Type:         graphql.EventSpecTypeAsyncAPI,
		DefinitionID: id,
	}

	deprecated := false
	forRemoval := false

	v := &graphql.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return &graphql.EventDefinition{
		ID:            id,
		ApplicationID: appID,
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Spec:          spec,
		Group:         str.Ptr("group_" + placeholder),
		Version:       v,
	}
}

func fixModelEventAPIDefinitionInput() *model.EventAPIDefinitionInput {
	data := "data"
	format := model.SpecFormatYaml

	spec := &model.EventAPISpecInput{
		Data:          &data,
		EventSpecType: model.EventAPISpecTypeAsyncAPI,
		Format:        format,
		FetchRequest:  &model.FetchRequestInput{},
	}

	deprecated := false
	forRemoval := false

	v := &model.VersionInput{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return &model.EventAPIDefinitionInput{
		Name:        "name",
		Description: str.Ptr("description"),
		Group:       str.Ptr("group"),
		Spec:        spec,
		Version:     v,
	}
}

func fixGQLEventAPIDefinitionInput() *graphql.EventDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.EventSpecInput{
		Data:         &data,
		Type:         graphql.EventSpecTypeAsyncAPI,
		Format:       graphql.SpecFormatYaml,
		FetchRequest: &graphql.FetchRequestInput{},
	}

	deprecated := false
	forRemoval := false

	v := &graphql.VersionInput{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return &graphql.EventDefinitionInput{
		Name:        "name",
		Description: str.Ptr("description"),
		Group:       str.Ptr("group"),
		Spec:        spec,
		Version:     v,
	}
}

func fixFullEventAPIDef(id, placeholder string) eventapi.Entity {
	v := fixVersionEntity()
	return eventapi.Entity{
		ID:          id,
		AppID:       appID,
		TenantID:    tenantID,
		Name:        placeholder,
		GroupName:   repo.NewValidNullableString("group_" + placeholder),
		Description: repo.NewValidNullableString("desc_" + placeholder),
		EntitySpec: eventapi.EntitySpec{
			SpecData:   repo.NewValidNullableString("data"),
			SpecType:   repo.NewValidNullableString(string(model.EventAPISpecTypeAsyncAPI)),
			SpecFormat: repo.NewValidNullableString(string(model.SpecFormatJSON)),
		},
		Version: v,
	}
}

func fixMinEntityEventAPIDef(id, placeholder string) eventapi.Entity {
	return eventapi.Entity{ID: id, TenantID: tenantID, AppID: appID, Name: placeholder}
}

func fixVersionModel() model.Version {
	deprecated := false
	forRemoval := false
	return model.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}
}

func fixVersionEntity() version.Version {
	return version.Version{
		VersionDepracated:      repo.NewValidNullableBool(false),
		VersionValue:           repo.NewValidNullableString("v1.1"),
		VersionDepracatedSince: repo.NewValidNullableString("v1.0"),
		VersionForRemoval:      repo.NewValidNullableBool(false),
	}
}

func fixEventAPIDefinitionColumns() []string {
	return []string{"id", "tenant_id", "app_id", "name", "description", "group_name", "spec_data",
		"spec_format", "spec_type", "version_value", "version_deprecated",
		"version_deprecated_since", "version_for_removal"}
}

func fixEventAPIDefinitionRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, appID, placeholder, "desc_" + placeholder, "group_" + placeholder,
		"data", "JSON", "ASYNC_API", "v1.1", false, "v1.0", false}
}

func fixEventAPICreateArgs(id string, api model.EventAPIDefinition) []driver.Value {
	return []driver.Value{id, tenantID, appID, api.Name, api.Description, api.Group,
		api.Spec.Data, string(api.Spec.Format), string(api.Spec.Type), api.Version.Value, api.Version.Deprecated,
		api.Version.DeprecatedSince, api.Version.ForRemoval}
}

func fixModelFetchRequest(id, url string, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		Tenant: tenantID,
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.EventAPIFetchRequestReference,
		ObjectID:   "foo",
	}
}

func fixGQLFetchRequest(url string, timestamp time.Time) *graphql.FetchRequest {
	return &graphql.FetchRequest{
		Filter: nil,
		Mode:   graphql.FetchModeSingle,
		Auth:   nil,
		URL:    url,
		Status: &graphql.FetchRequestStatus{
			Timestamp: graphql.Timestamp(timestamp),
			Condition: graphql.FetchRequestStatusConditionInitial,
		},
	}
}
