package eventdef_test

import (
	"database/sql"
	"database/sql/driver"
	"time"

	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	eventID          = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	specID           = "sssssssss-ssss-ssss-ssss-ssssssssssss"
	tenantID         = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
	externalTenantID = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	bundleID         = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

var fixedTimestamp = time.Now()

func fixEventDefinitionModel(id string, bndlID string, name string) *model.EventDefinition {
	return &model.EventDefinition{
		ID:       id,
		BundleID: bndlID,
		Name:     name,
	}
}

func fixFullEventDefinitionModel(placeholder string) (model.EventDefinition, model.Spec) {
	eventType := model.EventSpecTypeAsyncAPI
	spec := model.Spec{
		ID:         specID,
		Data:       str.Ptr("spec_data_" + placeholder),
		Format:     model.SpecFormatYaml,
		ObjectType: model.EventSpecReference,
		ObjectID:   eventID,
		EventType:  &eventType,
	}

	deprecated := false
	forRemoval := false

	v := &model.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return model.EventDefinition{
		ID:          eventID,
		Tenant:      tenantID,
		BundleID:    bundleID,
		Name:        placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
		BaseEntity: &model.BaseEntity{
			Ready:     true,
			CreatedAt: fixedTimestamp,
			UpdatedAt: fixedTimestamp,
			DeletedAt: time.Time{},
			Error:     nil,
		},
	}, spec
}

func fixFullGQLEventDefinition(placeholder string) *graphql.EventDefinition {
	data := graphql.CLOB("spec_data_" + placeholder)

	spec := &graphql.EventSpec{
		Data:         &data,
		Format:       graphql.SpecFormatYaml,
		Type:         graphql.EventSpecTypeAsyncAPI,
		DefinitionID: eventID,
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
		ID:          eventID,
		BundleID:    bundleID,
		Name:        placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Spec:        spec,
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
		BaseEntity: &graphql.BaseEntity{
			Ready:     true,
			Error:     nil,
			CreatedAt: graphql.Timestamp(fixedTimestamp),
			UpdatedAt: graphql.Timestamp(fixedTimestamp),
			DeletedAt: graphql.Timestamp(time.Time{}),
		},
	}
}

func fixModelEventDefinitionInput(name, description string, group string) (*model.EventDefinitionInput, *model.SpecInput) {
	data := "data"
	eventType := model.EventSpecTypeAsyncAPI

	spec := &model.SpecInput{
		Data:         &data,
		EventType:    &eventType,
		Format:       model.SpecFormatYaml,
		FetchRequest: &model.FetchRequestInput{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	v := &model.VersionInput{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	return &model.EventDefinitionInput{
		Name:        name,
		Description: &description,
		Group:       &group,
		Version:     v,
	}, spec
}

func fixGQLEventDefinitionInput(name, description string, group string) *graphql.EventDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.EventSpecInput{
		Data:         &data,
		Type:         graphql.EventSpecTypeAsyncAPI,
		Format:       graphql.SpecFormatYaml,
		FetchRequest: &graphql.FetchRequestInput{},
	}

	deprecated := false
	deprecatedSince := ""
	forRemoval := false

	v := &graphql.VersionInput{
		Value:           "1.0.0",
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}

	return &graphql.EventDefinitionInput{
		Name:        name,
		Description: &description,
		Group:       &group,
		Spec:        spec,
		Version:     v,
	}
}

func fixEntityEventDefinition(id string, bndlID string, name string) *event.Entity {
	return &event.Entity{
		ID:     id,
		BndlID: bndlID,
		Name:   name,
	}
}

func fixFullEntityEventDefinition(eventID, placeholder string) *event.Entity {
	boolPlaceholder := false

	return &event.Entity{
		ID:          eventID,
		TenantID:    tenantID,
		BndlID:      bundleID,
		Name:        placeholder,
		Description: repo.NewValidNullableString("desc_" + placeholder),
		GroupName:   repo.NewValidNullableString("group_" + placeholder),
		Version: version.Version{
			VersionValue:           repo.NewNullableString(str.Ptr("v1.1")),
			VersionDepracated:      repo.NewNullableBool(&boolPlaceholder),
			VersionDepracatedSince: repo.NewNullableString(str.Ptr("v1.0")),
			VersionForRemoval:      repo.NewNullableBool(&boolPlaceholder),
		},
		BaseEntity: &repo.BaseEntity{
			Ready:     true,
			CreatedAt: fixedTimestamp,
			UpdatedAt: fixedTimestamp,
			DeletedAt: time.Time{},
			Error:     sql.NullString{},
		},
	}
}

func fixEventDefinitionColumns() []string {
	return []string{"id", "tenant_id", "bundle_id", "name", "description", "group_name", "version_value", "version_deprecated",
		"version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error"}
}

func fixEventDefinitionRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, placeholder, "desc_" + placeholder, "group_" + placeholder, "v1.1", false, "v1.0", false, true, fixedTimestamp, fixedTimestamp, time.Time{}, nil}
}

func fixEventCreateArgs(id string, event *model.EventDefinition) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, event.Name, event.Description, event.Group, event.Version.Value, event.Version.Deprecated, event.Version.DeprecatedSince,
		event.Version.ForRemoval, event.Ready, event.CreatedAt, event.UpdatedAt, event.DeletedAt, event.Error}
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
		ObjectType: model.SpecFetchRequestReference,
		ObjectID:   specID,
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
