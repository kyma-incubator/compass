package api_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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
	packageID        = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
	appID            = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	ordID            = "com.compass.ord.v1"
	extensible       = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	successors       = `["sap.s4:apiResource:API_BILL_OF_MATERIAL_SRV:v2"]`
	resourceHash     = "123456"
)

var fixedTimestamp = time.Now()

func fixAPIDefinitionModel(id string, name, targetURL string) *model.APIDefinition {
	return &model.APIDefinition{
		Name:       name,
		TargetURLs: api.ConvertTargetUrlToJsonArray(targetURL),
		BaseEntity: &model.BaseEntity{ID: id},
	}
}

func fixFullAPIDefinitionModel(placeholder string) (model.APIDefinition, model.Spec, model.BundleReference) {
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

	apiBundleReference := model.BundleReference{
		Tenant:              tenantID,
		BundleID:            str.Ptr(bundleID),
		ObjectType:          model.BundleAPIReference,
		ObjectID:            str.Ptr(apiDefID),
		APIDefaultTargetURL: str.Ptr(fmt.Sprintf("https://%s.com", placeholder)),
	}

	boolVar := false
	return model.APIDefinition{
		ApplicationID:                           appID,
		PackageID:                               str.Ptr(packageID),
		Tenant:                                  tenantID,
		Name:                                    placeholder,
		Description:                             str.Ptr("desc_" + placeholder),
		TargetURLs:                              api.ConvertTargetUrlToJsonArray(fmt.Sprintf("https://%s.com", placeholder)),
		Group:                                   str.Ptr("group_" + placeholder),
		OrdID:                                   str.Ptr(ordID),
		ShortDescription:                        str.Ptr("shortDescription"),
		SystemInstanceAware:                     &boolVar,
		ApiProtocol:                             str.Ptr("apiProtocol"),
		Tags:                                    json.RawMessage("[]"),
		Countries:                               json.RawMessage("[]"),
		Links:                                   json.RawMessage("[]"),
		APIResourceLinks:                        json.RawMessage("[]"),
		ReleaseStatus:                           str.Ptr("releaseStatus"),
		SunsetDate:                              str.Ptr("sunsetDate"),
		Successors:                              json.RawMessage(successors),
		ChangeLogEntries:                        json.RawMessage("[]"),
		Labels:                                  json.RawMessage("[]"),
		Visibility:                              str.Ptr("visibility"),
		Disabled:                                &boolVar,
		PartOfProducts:                          json.RawMessage("[]"),
		LineOfBusiness:                          json.RawMessage("[]"),
		Industry:                                json.RawMessage("[]"),
		ImplementationStandard:                  str.Ptr("implementationStandard"),
		CustomImplementationStandard:            str.Ptr("customImplementationStandard"),
		CustomImplementationStandardDescription: str.Ptr("customImplementationStandardDescription"),
		Version:                                 v,
		Extensible:                              json.RawMessage(extensible),
		ResourceHash:                            str.Ptr(resourceHash),
		BaseEntity: &model.BaseEntity{
			ID:        apiDefID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}, spec, apiBundleReference
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
		Name:         name,
		Description:  &description,
		TargetURLs:   api.ConvertTargetUrlToJsonArray("https://test-url.com"),
		Group:        &group,
		VersionInput: v,
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

func fixEntityAPIDefinition(id string, name, targetUrl string) *api.Entity {
	return &api.Entity{
		Name:       name,
		TargetURLs: repo.NewValidNullableString(`["` + targetUrl + `"]`),
		BaseEntity: &repo.BaseEntity{ID: id},
	}
}

func fixFullEntityAPIDefinition(apiDefID, placeholder string) api.Entity {
	return api.Entity{
		TenantID:                                tenantID,
		ApplicationID:                           appID,
		PackageID:                               repo.NewValidNullableString(packageID),
		Name:                                    placeholder,
		Description:                             repo.NewValidNullableString("desc_" + placeholder),
		Group:                                   repo.NewValidNullableString("group_" + placeholder),
		TargetURLs:                              repo.NewValidNullableString(`["` + fmt.Sprintf("https://%s.com", placeholder) + `"]`),
		OrdID:                                   repo.NewValidNullableString(ordID),
		ShortDescription:                        repo.NewValidNullableString("shortDescription"),
		SystemInstanceAware:                     repo.NewValidNullableBool(false),
		ApiProtocol:                             repo.NewValidNullableString("apiProtocol"),
		Tags:                                    repo.NewValidNullableString("[]"),
		Countries:                               repo.NewValidNullableString("[]"),
		Links:                                   repo.NewValidNullableString("[]"),
		APIResourceLinks:                        repo.NewValidNullableString("[]"),
		ReleaseStatus:                           repo.NewValidNullableString("releaseStatus"),
		SunsetDate:                              repo.NewValidNullableString("sunsetDate"),
		Successors:                              repo.NewValidNullableString(successors),
		ChangeLogEntries:                        repo.NewValidNullableString("[]"),
		Labels:                                  repo.NewValidNullableString("[]"),
		Visibility:                              repo.NewValidNullableString("visibility"),
		Disabled:                                repo.NewValidNullableBool(false),
		PartOfProducts:                          repo.NewValidNullableString("[]"),
		LineOfBusiness:                          repo.NewValidNullableString("[]"),
		Industry:                                repo.NewValidNullableString("[]"),
		ImplementationStandard:                  repo.NewValidNullableString("implementationStandard"),
		CustomImplementationStandard:            repo.NewValidNullableString("customImplementationStandard"),
		CustomImplementationStandardDescription: repo.NewValidNullableString("customImplementationStandardDescription"),
		Extensible:                              repo.NewValidNullableString(extensible),
		Version: version.Version{
			Value:           repo.NewNullableString(str.Ptr("v1.1")),
			Deprecated:      repo.NewValidNullableBool(false),
			DeprecatedSince: repo.NewNullableString(str.Ptr("v1.0")),
			ForRemoval:      repo.NewValidNullableBool(false),
		},
		ResourceHash: repo.NewValidNullableString(resourceHash),
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

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN \( with recursive children AS \(SELECT t1\.id, t1\.parent FROM business_tenant_mappings t1 WHERE id = \$%d UNION ALL SELECT t2\.id, t2\.parent FROM business_tenant_mappings t2 INNER JOIN children t on t\.id = t2\.parent\) SELECT id from children \)`, i)
}

func fixAPIDefinitionColumns() []string {
	return []string{"id", "tenant_id", "app_id", "package_id", "name", "description", "group_name", "ord_id",
		"short_description", "system_instance_aware", "api_protocol", "tags", "countries", "links", "api_resource_links", "release_status",
		"sunset_date", "changelog_entries", "labels", "visibility", "disabled", "part_of_products", "line_of_business",
		"industry", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard", "custom_implementation_standard_description", "target_urls", "extensible", "successors", "resource_hash"}
}

func fixAPIDefinitionRow(id, placeholder string) []driver.Value {
	boolVar := false
	return []driver.Value{id, tenantID, appID, packageID, placeholder, "desc_" + placeholder, "group_" + placeholder,
		ordID, "shortDescription", &boolVar, "apiProtocol", repo.NewValidNullableString("[]"),
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "releaseStatus", "sunsetDate", repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "visibility", &boolVar,
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "v1.1", false, "v1.0", false, true, fixedTimestamp, time.Time{}, time.Time{}, nil, "implementationStandard", "customImplementationStandard", "customImplementationStandardDescription", repo.NewValidNullableString(`["` + fmt.Sprintf("https://%s.com", placeholder) + `"]`), repo.NewValidNullableString(extensible), repo.NewValidNullableString(successors), repo.NewValidNullableString(resourceHash)}
}

func fixAPICreateArgs(id string, apiDef *model.APIDefinition) []driver.Value {
	return []driver.Value{id, tenantID, appID, packageID, apiDef.Name, apiDef.Description, apiDef.Group,
		apiDef.OrdID, apiDef.ShortDescription, apiDef.SystemInstanceAware, apiDef.ApiProtocol, repo.NewNullableStringFromJSONRawMessage(apiDef.Tags), repo.NewNullableStringFromJSONRawMessage(apiDef.Countries),
		repo.NewNullableStringFromJSONRawMessage(apiDef.Links), repo.NewNullableStringFromJSONRawMessage(apiDef.APIResourceLinks),
		apiDef.ReleaseStatus, apiDef.SunsetDate, repo.NewNullableStringFromJSONRawMessage(apiDef.ChangeLogEntries), repo.NewNullableStringFromJSONRawMessage(apiDef.Labels), apiDef.Visibility,
		apiDef.Disabled, repo.NewNullableStringFromJSONRawMessage(apiDef.PartOfProducts), repo.NewNullableStringFromJSONRawMessage(apiDef.LineOfBusiness), repo.NewNullableStringFromJSONRawMessage(apiDef.Industry),
		apiDef.Version.Value, apiDef.Version.Deprecated, apiDef.Version.DeprecatedSince, apiDef.Version.ForRemoval, apiDef.Ready, apiDef.CreatedAt, apiDef.UpdatedAt, apiDef.DeletedAt, apiDef.Error, apiDef.ImplementationStandard, apiDef.CustomImplementationStandard, apiDef.CustomImplementationStandardDescription, repo.NewNullableStringFromJSONRawMessage(apiDef.TargetURLs), extensible, repo.NewNullableStringFromJSONRawMessage(apiDef.Successors), apiDef.ResourceHash}
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
