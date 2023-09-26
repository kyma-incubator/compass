package eventdef_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	bundleID         = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	packageID        = "ppppppppp-pppp-pppp-pppp-pppppppppppp"
	ordID            = "com.compass.ord.v1"
	localTenantID    = "localTenantID"
	extensible       = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	successors       = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1"]`
	resourceHash     = "123456"
	publicVisibility = "public"
)

var (
	fixedTimestamp       = time.Now()
	appID                = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID = "aaaaaaaaa-bbbb-cccc-dddd-aaaaaaaaaaaa"
)

func fixEventDefinitionModel(id string, name string) *model.EventDefinition {
	return &model.EventDefinition{
		Name:       name,
		BaseEntity: &model.BaseEntity{ID: id},
		Visibility: str.Ptr(publicVisibility),
	}
}

func fixFullEventDefinitionModel(placeholder string) (model.EventDefinition, model.Spec, model.BundleReference) {
	return fixFullEventDefinitionModelWithID(eventID, placeholder)
}

func fixFullEventDefinitionModelWithID(id, placeholder string) (model.EventDefinition, model.Spec, model.BundleReference) {
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

	eventBundleReference := model.BundleReference{
		BundleID:   str.Ptr(bundleID),
		ObjectType: model.BundleEventReference,
		ObjectID:   str.Ptr(eventID),
	}

	boolVar := false
	return model.EventDefinition{
		ApplicationID:                           &appID,
		PackageID:                               str.Ptr(packageID),
		Name:                                    placeholder,
		Description:                             str.Ptr("desc_" + placeholder),
		Group:                                   str.Ptr("group_" + placeholder),
		OrdID:                                   str.Ptr(ordID),
		LocalTenantID:                           str.Ptr(localTenantID),
		ShortDescription:                        str.Ptr("shortDescription"),
		SystemInstanceAware:                     &boolVar,
		PolicyLevel:                             nil,
		CustomPolicyLevel:                       nil,
		Tags:                                    json.RawMessage("[]"),
		Countries:                               json.RawMessage("[]"),
		Links:                                   json.RawMessage("[]"),
		ReleaseStatus:                           str.Ptr("releaseStatus"),
		SunsetDate:                              str.Ptr("sunsetDate"),
		Successors:                              json.RawMessage(successors),
		ChangeLogEntries:                        json.RawMessage("[]"),
		Labels:                                  json.RawMessage("[]"),
		Visibility:                              str.Ptr(publicVisibility),
		Disabled:                                &boolVar,
		PartOfProducts:                          json.RawMessage("[]"),
		LineOfBusiness:                          json.RawMessage("[]"),
		Industry:                                json.RawMessage("[]"),
		ImplementationStandard:                  str.Ptr("implementationStandard"),
		CustomImplementationStandard:            str.Ptr("customImplementationStandard"),
		CustomImplementationStandardDescription: str.Ptr("customImplementationStandardDescription"),
		Extensible:                              json.RawMessage(extensible),
		Version:                                 v,
		Hierarchy:                               json.RawMessage("[]"),
		DocumentationLabels:                     json.RawMessage("[]"),
		CorrelationIDs:                          json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			ID:        id,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		ResourceHash: str.Ptr(resourceHash),
	}, spec, eventBundleReference
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
		BundleID:    bundleID,
		Name:        placeholder,
		Description: str.Ptr("desc_" + placeholder),
		Spec:        spec,
		Group:       str.Ptr("group_" + placeholder),
		Version:     v,
		BaseEntity: &graphql.BaseEntity{
			ID:        eventID,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
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
		Name:         name,
		Description:  &description,
		Group:        &group,
		VersionInput: v,
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

func fixEntityEventDefinition(id string, name string) *event.Entity {
	return &event.Entity{
		Name:       name,
		BaseEntity: &repo.BaseEntity{ID: id},
		Visibility: publicVisibility,
	}
}

func fixFullEntityEventDefinition(eventID, placeholder string) *event.Entity {
	return &event.Entity{
		ApplicationID:                           repo.NewValidNullableString(appID),
		PackageID:                               repo.NewValidNullableString(packageID),
		Name:                                    placeholder,
		Description:                             repo.NewValidNullableString("desc_" + placeholder),
		GroupName:                               repo.NewValidNullableString("group_" + placeholder),
		OrdID:                                   repo.NewValidNullableString(ordID),
		LocalTenantID:                           repo.NewValidNullableString(localTenantID),
		ShortDescription:                        repo.NewValidNullableString("shortDescription"),
		SystemInstanceAware:                     repo.NewValidNullableBool(false),
		PolicyLevel:                             sql.NullString{},
		CustomPolicyLevel:                       sql.NullString{},
		ChangeLogEntries:                        repo.NewValidNullableString("[]"),
		Links:                                   repo.NewValidNullableString("[]"),
		Tags:                                    repo.NewValidNullableString("[]"),
		Countries:                               repo.NewValidNullableString("[]"),
		ReleaseStatus:                           repo.NewValidNullableString("releaseStatus"),
		SunsetDate:                              repo.NewValidNullableString("sunsetDate"),
		Successors:                              repo.NewValidNullableString(successors),
		Labels:                                  repo.NewValidNullableString("[]"),
		Visibility:                              publicVisibility,
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
		ResourceHash:        repo.NewValidNullableString(resourceHash),
		Hierarchy:           repo.NewValidNullableString("[]"),
		DocumentationLabels: repo.NewValidNullableString("[]"),
		CorrelationIDs:      repo.NewValidNullableString("[]"),
		BaseEntity: &repo.BaseEntity{
			ID:        eventID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
	}
}

func fixEventDefinitionColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "package_id", "name", "description", "group_name", "ord_id", "local_tenant_id",
		"short_description", "system_instance_aware", "policy_level", "custom_policy_level",
		"changelog_entries", "links", "tags", "countries", "release_status",
		"sunset_date", "labels", "visibility", "disabled", "part_of_products", "line_of_business", "industry", "version_value", "version_deprecated", "version_deprecated_since",
		"version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "implementation_standard", "custom_implementation_standard",
		"custom_implementation_standard_description", "extensible", "successors", "resource_hash", "hierarchy", "documentation_labels", "correlation_ids"}
}

func fixEventDefinitionRow(id, placeholder string) []driver.Value {
	boolVar := false
	return []driver.Value{id, appID, repo.NewValidNullableString(""), packageID, placeholder, "desc_" + placeholder, "group_" + placeholder, ordID, localTenantID, "shortDescription", &boolVar, nil, nil,
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "releaseStatus", "sunsetDate", repo.NewValidNullableString("[]"), publicVisibility, &boolVar,
		repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), "v1.1", false, "v1.0", false, true, fixedTimestamp, time.Time{}, time.Time{}, nil, "implementationStandard", "customImplementationStandard", "customImplementationStandardDescription", repo.NewValidNullableString(extensible),
		repo.NewValidNullableString(successors), repo.NewValidNullableString(resourceHash), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]"), repo.NewValidNullableString("[]")}
}

func fixEventCreateArgs(id string, event *model.EventDefinition) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(""), packageID, event.Name, event.Description, event.Group, event.OrdID, event.LocalTenantID, event.ShortDescription,
		event.SystemInstanceAware, event.PolicyLevel, event.CustomPolicyLevel, repo.NewNullableStringFromJSONRawMessage(event.ChangeLogEntries), repo.NewNullableStringFromJSONRawMessage(event.Links),
		repo.NewNullableStringFromJSONRawMessage(event.Tags), repo.NewNullableStringFromJSONRawMessage(event.Countries), event.ReleaseStatus, event.SunsetDate,
		repo.NewNullableStringFromJSONRawMessage(event.Labels), event.Visibility,
		event.Disabled, repo.NewNullableStringFromJSONRawMessage(event.PartOfProducts), repo.NewNullableStringFromJSONRawMessage(event.LineOfBusiness), repo.NewNullableStringFromJSONRawMessage(event.Industry),
		event.Version.Value, event.Version.Deprecated, event.Version.DeprecatedSince, event.Version.ForRemoval, event.Ready, event.CreatedAt, event.UpdatedAt, event.DeletedAt, event.Error, event.ImplementationStandard, event.CustomImplementationStandard, event.CustomImplementationStandardDescription, repo.NewNullableStringFromJSONRawMessage(event.Extensible),
		repo.NewNullableStringFromJSONRawMessage(event.Successors), resourceHash, repo.NewNullableStringFromJSONRawMessage(event.Hierarchy), repo.NewNullableStringFromJSONRawMessage(event.DocumentationLabels), repo.NewNullableStringFromJSONRawMessage(event.CorrelationIDs)}
}

func fixModelFetchRequest(id, url string, timestamp time.Time) *model.FetchRequest {
	return &model.FetchRequest{
		ID:     id,
		URL:    url,
		Auth:   nil,
		Mode:   "SINGLE",
		Filter: nil,
		Status: &model.FetchRequestStatus{
			Condition: model.FetchRequestStatusConditionInitial,
			Timestamp: timestamp,
		},
		ObjectType: model.EventSpecFetchRequestReference,
		ObjectID:   specID,
	}
}

func fixModelBundleReference(bundleID, eventID string) *model.BundleReference {
	return &model.BundleReference{
		BundleID:   str.Ptr(bundleID),
		ObjectType: model.BundleEventReference,
		ObjectID:   str.Ptr(eventID),
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
