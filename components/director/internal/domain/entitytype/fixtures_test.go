package entitytype_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	ID                 = "ID"
	ready              = true
	tenantID           = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	ordID              = "com.compass.v1"
	localID            = "BusinessPartner"
	correlationIDs     = `["sap.s4:sot:BusinessPartner", "sap.s4:sot:CostCenter", "sap.s4:sot:WorkforcePerson"]`
	level              = "aggregate"
	title              = "BusinessPartner"
	packageID          = "sap.xref:package:SomePackage:v1"
	publicVisibility   = "public"
	product            = `["sap:product:S4HANA_OD:"]`
	releaseStatus      = "active"
	testTenant         = "tnt"
	testExternalTenant = "external-tnt"
	entityTypeID       = "entity-type-id"
)

var (
	testError               = errors.New("test error")
	fixedTimestamp          = time.Now()
	appID                   = "appID"
	appTemplateVersionID    = "appTemplateVersionID"
	shortDescription        = "A business partner is a person, an organization, or a group of persons or organizations in which a company has a business interest."
	description             = "A workforce person is a natural person with a work agreement or relationship in form of a work assignment; it can be an employee or a contingent worker.\n"
	systemInstanceAware     = false
	policyLevel             = "custom"
	customPolicyLevel       = "sap:core:v1"
	sunsetDate              = "2022-01-08T15:47:04+00:00"
	successors              = `["sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1"]`
	extensible              = `{"supported":"automatic","description":"Please find the extensibility documentation"}`
	tags                    = `["storage","high-availability"]`
	resourceHash            = "123456"
	version_value           = "v1.1"
	version_deprecated      = false
	version_deprecatedSince = "v1.0"
	version_forRemoval      = false
	changeLogEntries        = removeWhitespace(`[
        {
		  "date": "2020-04-29",
		  "description": "lorem ipsum dolor sit amet",
		  "releaseStatus": "active",
		  "url": "https://example.com/changelog/v1",
          "version": "1.0.0"
        }
      ]`)
	links = removeWhitespace(`[
        {
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 1",
          "url": "https://example.com/2018/04/11/testing/"
        },
		{
		  "description": "lorem ipsum dolor nem",
          "title": "Link Title 2",
          "url": "http://example.com/testing/"
        }
      ]`)
	labels = removeWhitespace(`{
        "label-key-1": [
          "label-value-1",
          "label-value-2"
        ]
      }`)
	documentLabels = removeWhitespace(`{
        "Some Aspect": ["Markdown Documentation [with links](#)", "With multiple values"]
      }`)
)

func removeWhitespace(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, " ", ""), "\n", ""), "\t", "")
}

func fixVersionEntity(value string, deprecated bool, deprecatedSince string, forRemoval bool) version.Version {
	return version.Version{
		Value:           repo.NewNullableString(&value),
		Deprecated:      repo.NewNullableBool(&deprecated),
		DeprecatedSince: repo.NewNullableString(&deprecatedSince),
		ForRemoval:      repo.NewNullableBool(&forRemoval),
	}
}

func fixVersionModel(value string, deprecated bool, deprecatedSince string, forRemoval bool) *model.Version {
	return &model.Version{
		Value:           value,
		Deprecated:      &deprecated,
		DeprecatedSince: &deprecatedSince,
		ForRemoval:      &forRemoval,
	}
}

func fixEntityTypeEntity(entityTypeID string) *entitytype.Entity {
	return &entitytype.Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        entityTypeID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
		ApplicationID:                repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID: repo.NewValidNullableString(appTemplateVersionID),
		OrdID:                        ordID,
		LocalID:                      localID,
		CorrelationIDs:               repo.NewValidNullableString(correlationIDs),
		Level:                        level,
		Title:                        title,
		ShortDescription:             repo.NewNullableString(&shortDescription),
		Description:                  repo.NewNullableString(&description),
		SystemInstanceAware:          repo.NewNullableBool(&systemInstanceAware),
		ChangeLogEntries:             repo.NewNullableStringFromJSONRawMessage(json.RawMessage(changeLogEntries)),
		PackageID:                    packageID,
		Visibility:                   publicVisibility,
		Links:                        repo.NewNullableStringFromJSONRawMessage(json.RawMessage(links)),
		PartOfProducts:               repo.NewNullableStringFromJSONRawMessage(json.RawMessage(product)),
		PolicyLevel:                  repo.NewNullableString(&policyLevel),
		CustomPolicyLevel:            repo.NewNullableString(&customPolicyLevel),
		ReleaseStatus:                releaseStatus,
		SunsetDate:                   repo.NewNullableString(&sunsetDate),
		Successors:                   repo.NewNullableStringFromJSONRawMessage(json.RawMessage(successors)),
		Extensible:                   repo.NewNullableStringFromJSONRawMessage(json.RawMessage(extensible)),
		Tags:                         repo.NewNullableStringFromJSONRawMessage(json.RawMessage(tags)),
		Labels:                       repo.NewNullableStringFromJSONRawMessage(json.RawMessage(labels)),
		DocumentationLabels:          repo.NewNullableStringFromJSONRawMessage(json.RawMessage(documentLabels)),
		Version:                      fixVersionEntity(version_value, version_deprecated, version_deprecatedSince, version_forRemoval),
		ResourceHash:                 repo.NewNullableString(&resourceHash),
	}
}

func fixEntityTypeModel(entityTypeID string) *model.EntityType {
	return &model.EntityType{
		BaseEntity: &model.BaseEntity{
			ID:        entityTypeID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		OrdID:                        ordID,
		LocalID:                      localID,
		CorrelationIDs:               json.RawMessage(correlationIDs),
		Level:                        level,
		Title:                        title,
		ShortDescription:             &shortDescription,
		Description:                  &description,
		SystemInstanceAware:          &systemInstanceAware,
		ChangeLogEntries:             json.RawMessage(changeLogEntries),
		OrdPackageID:                 packageID,
		Visibility:                   publicVisibility,
		Links:                        json.RawMessage(links),
		PartOfProducts:               json.RawMessage(product),
		PolicyLevel:                  &policyLevel,
		CustomPolicyLevel:            &customPolicyLevel,
		ReleaseStatus:                releaseStatus,
		SunsetDate:                   &sunsetDate,
		Successors:                   json.RawMessage(successors),
		Extensible:                   json.RawMessage(extensible),
		Tags:                         json.RawMessage(tags),
		Labels:                       json.RawMessage(labels),
		DocumentationLabels:          json.RawMessage(documentLabels),
		Version:                      fixVersionModel(version_value, version_deprecated, version_deprecatedSince, version_forRemoval),
		ResourceHash:                 &resourceHash,
	}
}

func fixEntityTypeInputModel() model.EntityTypeInput {
	return model.EntityTypeInput{
		OrdID:               ordID,
		LocalID:             localID,
		CorrelationIDs:      json.RawMessage(correlationIDs),
		Level:               level,
		Title:               title,
		ShortDescription:    &shortDescription,
		Description:         &description,
		SystemInstanceAware: &systemInstanceAware,
		ChangeLogEntries:    json.RawMessage(changeLogEntries),
		OrdPackageID:        packageID,
		Visibility:          publicVisibility,
		Links:               json.RawMessage(links),
		PartOfProducts:      json.RawMessage(product),
		PolicyLevel:         &policyLevel,
		CustomPolicyLevel:   &customPolicyLevel,
		ReleaseStatus:       releaseStatus,
		SunsetDate:          &sunsetDate,
		Successors:          json.RawMessage(successors),
		Extensible:          json.RawMessage(extensible),
		Tags:                json.RawMessage(tags),
		Labels:              json.RawMessage(labels),
		DocumentationLabels: json.RawMessage(documentLabels),
	}
}

func fixEntityTypeColumns() []string {
	return []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "app_id", "app_template_version_id", "ord_id", "local_id",
		"correlation_ids", "level", "title", "short_description", "description", "system_instance_aware", "changelog_entries", "package_id", "visibility",
		"links", "part_of_products", "policy_level", "custom_policy_level", "release_status", "sunset_date", "successors", "extensible", "tags", "labels",
		"documentation_labels", "resource_hash", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
}

func fixEntityTypeRow(id string) []driver.Value {
	return []driver.Value{id, ready, fixedTimestamp, time.Time{}, time.Time{}, nil, appID, repo.NewValidNullableString(""), ordID, localID,
		repo.NewValidNullableString(correlationIDs), level, title, repo.NewNullableString(&shortDescription), repo.NewNullableString(&description), repo.NewNullableBool(&systemInstanceAware), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(changeLogEntries)), packageID, publicVisibility,
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage(links)), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(product)), repo.NewNullableString(&policyLevel), repo.NewNullableString(&customPolicyLevel), releaseStatus, repo.NewNullableString(&sunsetDate), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(successors)), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(extensible)), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(tags)), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(labels)),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage(documentLabels)), repo.NewNullableString(&resourceHash), repo.NewNullableString(&version_value), repo.NewNullableBool(&version_deprecated), repo.NewNullableString(&version_deprecatedSince), repo.NewNullableBool(&version_forRemoval)}
}
