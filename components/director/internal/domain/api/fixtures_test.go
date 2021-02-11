package api_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const (
	apiDefID         = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	specID           = "sssssssss-ssss-ssss-ssss-ssssssssssss"
	tenantID         = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
	externalTenantID = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	bundleID         = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

var fixedTimestamp = time.Now()

func fixAPIDefinitionModel(id string, bndlID string, name, targetURL string) *model.APIDefinition {
	return &model.APIDefinition{
		BundleID:   bndlID,
		Name:       name,
		TargetURL:  targetURL,
		BaseEntity: &model.BaseEntity{ID: id},
	}
}

func fixFullAPIDefinitionModel(placeholder string) (model.APIDefinition, model.Spec) {
	apiType := model.APISpecTypeOpenAPI
	spec := model.Spec{
		ID:         specID,
		Data:       str.Ptr("spec_data_" + placeholder),
		Format:     model.SpecFormatYaml,
		ObjectType: model.APISpecReference,
		ObjectID:   apiDefID,
		APIType:    &apiType,
	}

	deprecated := false
	forRemoval := false

	v := &model.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return model.APIDefinition{
		Tenant:      tenantID,
		BundleID:    bundleID,
		Name:        placeholder,
		Description: str.Ptr("desc_" + placeholder),
		TargetURL:   fmt.Sprintf("https://%s.com", placeholder),
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
		BaseEntity: &model.BaseEntity{
			ID:        apiDefID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}, spec
}

func fixFullGQLAPIDefinition(placeholder string) *graphql.APIDefinition {
	data := graphql.CLOB("spec_data_" + placeholder)

	spec := &graphql.APISpec{
		Data:         &data,
		Format:       graphql.SpecFormatYaml,
		Type:         graphql.APISpecTypeOpenAPI,
		DefinitionID: apiDefID,
	}

	deprecated := false
	forRemoval := false

	v := &graphql.Version{
		Value:           "v1.1",
		Deprecated:      &deprecated,
		DeprecatedSince: str.Ptr("v1.0"),
		ForRemoval:      &forRemoval,
	}

	return &graphql.APIDefinition{
		BundleID:    bundleID,
		Name:        placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Spec:        spec,
		TargetURL:   fmt.Sprintf("https://%s.com", placeholder),
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
		BaseEntity: &graphql.BaseEntity{
			ID:        apiDefID,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
		},
	}
}

func fixModelAPIDefinitionInput(name, description string, group string) (*model.APIDefinitionInput, *model.SpecInput) {
	data := "data"
	apiType := model.APISpecTypeOpenAPI

	spec := &model.SpecInput{
		Data:         &data,
		APIType:      &apiType,
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

	return &model.APIDefinitionInput{
		Name:        name,
		Description: &description,
		TargetURL:   "https://test-url.com",
		Group:       &group,
		Version:     v,
	}, spec
}

func fixGQLAPIDefinitionInput(name, description string, group string) *graphql.APIDefinitionInput {
	data := graphql.CLOB("data")

	spec := &graphql.APISpecInput{
		Data:         &data,
		Type:         graphql.APISpecTypeOpenAPI,
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

	return &graphql.APIDefinitionInput{
		Name:        name,
		Description: &description,
		TargetURL:   "https://test-url.com",
		Group:       &group,
		Spec:        spec,
		Version:     v,
	}
}

func fixEntityAPIDefinition(id string, bndlID string, name, targetUrl string) *api.Entity {
	return &api.Entity{
		BndlID:     bndlID,
		Name:       name,
		TargetURL:  targetUrl,
		BaseEntity: &repo.BaseEntity{ID: id},
	}
}

func fixFullEntityAPIDefinition(apiDefID, placeholder string) *api.Entity {
	boolPlaceholder := false

	return &api.Entity{
		TenantID:    tenantID,
		BndlID:      bundleID,
		Name:        placeholder,
		Description: repo.NewValidNullableString("desc_" + placeholder),
		Group:       repo.NewValidNullableString("group_" + placeholder),
		TargetURL:   fmt.Sprintf("https://%s.com", placeholder),
		Version: version.Version{
			VersionValue:           repo.NewNullableString(str.Ptr("v1.1")),
			VersionDepracated:      repo.NewNullableBool(&boolPlaceholder),
			VersionDepracatedSince: repo.NewNullableString(str.Ptr("v1.0")),
			VersionForRemoval:      repo.NewNullableBool(&boolPlaceholder),
		},
		BaseEntity: &repo.BaseEntity{
			ID:        apiDefID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
	}
}

func fixAPIDefinitionColumns() []string {
	return []string{"id", "tenant_id", "bundle_id", "name", "description", "group_name", "target_url", "version_value", "version_deprecated",
		"version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error"}
}

func fixAPIDefinitionRow(id, placeholder string) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, placeholder, "desc_" + placeholder, "group_" + placeholder,
		fmt.Sprintf("https://%s.com", placeholder), "v1.1", false, "v1.0", false, true, fixedTimestamp, time.Time{}, time.Time{}, nil}
}

func fixAPICreateArgs(id string, api *model.APIDefinition) []driver.Value {
	return []driver.Value{id, tenantID, bundleID, api.Name, api.Description, api.Group,
		api.TargetURL, api.Version.Value, api.Version.Deprecated, api.Version.DeprecatedSince,
		api.Version.ForRemoval, api.Ready, api.CreatedAt, api.UpdatedAt, api.DeletedAt, api.Error}
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

func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}
