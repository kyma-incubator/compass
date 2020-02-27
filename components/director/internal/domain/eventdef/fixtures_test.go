package eventdef_test

import (
	"database/sql/driver"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
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
	packageID  = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
)

func fixMinModelEventAPIDefinition(id, placeholder string) *model.EventDefinition {
	return &model.EventDefinition{ID: id, Tenant: tenantID, ApplicationID: str.Ptr(appID), PackageID: str.Ptr(packageID), Name: placeholder}
}

func fixGQLEventDefinition(id, placeholder string) *graphql.EventDefinition {
	return &graphql.EventDefinition{
		ID:            id,
		ApplicationID: str.Ptr(appID),
		PackageID:     str.Ptr(packageID),
		Name:          placeholder,
	}
}

func fixFullModelEventDefinition(id, placeholder string) model.EventDefinition {
	spec := &model.EventSpec{
		Data:   str.Ptr("data"),
		Format: model.SpecFormatJSON,
		Type:   model.EventSpecTypeAsyncAPI,
	}
	v := fixVersionModel()

	return model.EventDefinition{
		ID:            id,
		ApplicationID: str.Ptr(appID),
		Tenant:        tenantID,
		PackageID:     str.Ptr(packageID),
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Group:         str.Ptr("group_" + placeholder),
		Spec:          spec,
		Version:       &v,
	}
}

func fixDetailedGQLEventDefinition(id, placeholder string) *graphql.EventDefinition {
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
		ApplicationID: str.Ptr(appID),
		PackageID:     str.Ptr(packageID),
		Name:          placeholder,
		Description:   str.Ptr("desc_" + placeholder),
		Spec:          spec,
		Group:         str.Ptr("group_" + placeholder),
		Version:       v,
	}
}

func fixModelEventDefinitionInput() *model.EventDefinitionInput {
	data := "data"
	format := model.SpecFormatYaml

	spec := &model.EventSpecInput{
		Data:          &data,
		EventSpecType: model.EventSpecTypeAsyncAPI,
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

	return &model.EventDefinitionInput{
		Name:        "name",
		Description: str.Ptr("description"),
		Group:       str.Ptr("group"),
		Spec:        spec,
		Version:     v,
	}
}

func fixGQLEventDefinitionInput() *graphql.EventDefinitionInput {
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

func fixFullEventDef(id, placeholder string) eventdef.Entity {
	v := fixVersionEntity()
	return eventdef.Entity{
		ID:          id,
		AppID:       repo.NewNullableString(str.Ptr(appID)),
		PkgID:       repo.NewNullableString(str.Ptr(packageID)),
		TenantID:    tenantID,
		Name:        placeholder,
		GroupName:   repo.NewValidNullableString("group_" + placeholder),
		Description: repo.NewValidNullableString("desc_" + placeholder),
		EntitySpec: eventdef.EntitySpec{
			SpecData:   repo.NewValidNullableString("data"),
			SpecType:   repo.NewValidNullableString(string(model.EventSpecTypeAsyncAPI)),
			SpecFormat: repo.NewValidNullableString(string(model.SpecFormatJSON)),
		},
		Version: v,
	}
}

func fixMinEntityEventDef(id, placeholder string) eventdef.Entity {
	return eventdef.Entity{ID: id, TenantID: tenantID, AppID: repo.NewNullableString(str.Ptr(appID)),
		PkgID: repo.NewNullableString(str.Ptr(packageID)), Name: placeholder}
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

func fixEventDefinitionColumns() []string {
	return []string{"id", "tenant_id", "app_id", "package_id", "name", "description", "group_name", "spec_data",
		"spec_format", "spec_type", "version_value", "version_deprecated",
		"version_deprecated_since", "version_for_removal"}
}

func fixEventDefinitionRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, appID, packageID, placeholder, "desc_" + placeholder, "group_" + placeholder,
		"data", "JSON", "ASYNC_API", "v1.1", false, "v1.0", false}
}

func fixEventCreateArgs(id string, api model.EventDefinition) []driver.Value {
	return []driver.Value{id, tenantID, appID, packageID, api.Name, api.Description, api.Group,
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
