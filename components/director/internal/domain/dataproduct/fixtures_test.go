package dataproduct_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID = "external-tnt"
	dataProductID    = "data-product-id"
	dataProductOrdID = "sap.foo.bar:dataProduct:CustomerOrder:v1"
	localTenantID    = "localTenantID"
	title            = "Data Product title"
	shortDescription = "Short description for Data Product"
	description      = "Description for Data Product"
	packageID        = "packageID"
	lastUpdate       = "2023-12-20T10:46:04+00:00"
	publicVisibility = "public"
	releaseStatus    = "active"
	deprecationDate  = "2023-12-27T10:46:04+00:00"
	sunsetDate       = "2023-12-30T10:46:04+00:00"
	resourceHash     = "12234567890"
	dataProductType  = "base"
	category         = "other"
	outputPorts      = `[{"ordId":"sap.cic:apiResource:RetailTransactionOData:v1"}]`
)

var (
	fixedTimestamp         = time.Now()
	appID                  = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID   = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
	versionValue           = "v1.1"
	versionDeprecated      = false
	versionDeprecatedSince = "v1.0"
	versionForRemoval      = false
	disabled               = false
	systemInstanceAware    = false
	responsible            = "sap:foo:TEST"
	ready                  = true
	testErr                = errors.New("test error")
)

func fixDataProductInputModelWithPackageOrdID(packageOrdID string) model.DataProductInput {
	return model.DataProductInput{
		OrdID:               str.Ptr(dataProductOrdID),
		LocalTenantID:       str.Ptr(localTenantID),
		CorrelationIDs:      json.RawMessage("[]"),
		Title:               title,
		ShortDescription:    str.Ptr(shortDescription),
		Description:         str.Ptr(description),
		OrdPackageID:        str.Ptr(packageOrdID),
		Visibility:          str.Ptr(publicVisibility),
		LastUpdate:          str.Ptr(lastUpdate),
		ReleaseStatus:       str.Ptr(releaseStatus),
		Disabled:            &disabled,
		DeprecationDate:     str.Ptr(deprecationDate),
		SunsetDate:          str.Ptr(sunsetDate),
		Successors:          json.RawMessage("[]"),
		ChangeLogEntries:    json.RawMessage("[]"),
		Type:                dataProductType,
		Category:            category,
		EntityTypes:         json.RawMessage("[]"),
		InputPorts:          json.RawMessage("[]"),
		OutputPorts:         json.RawMessage(outputPorts),
		Responsible:         str.Ptr(responsible),
		DataProductLinks:    json.RawMessage("[]"),
		Links:               json.RawMessage("[]"),
		Industry:            json.RawMessage("[]"),
		LineOfBusiness:      json.RawMessage("[]"),
		Tags:                json.RawMessage("[]"),
		Labels:              json.RawMessage("[]"),
		DocumentationLabels: json.RawMessage("[]"),
		PolicyLevel:         nil,
		CustomPolicyLevel:   nil,
		SystemInstanceAware: &systemInstanceAware,
	}
}

func fixDataProductModel(dataProductID string) *model.DataProduct {
	return &model.DataProduct{
		BaseEntity: &model.BaseEntity{
			ID:        dataProductID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		ApplicationID:                &appID,
		ApplicationTemplateVersionID: &appTemplateVersionID,
		OrdID:                        str.Ptr(dataProductOrdID),
		LocalTenantID:                str.Ptr(localTenantID),
		CorrelationIDs:               json.RawMessage("[]"),
		Title:                        title,
		ShortDescription:             str.Ptr(shortDescription),
		Description:                  str.Ptr(description),
		PackageID:                    str.Ptr(packageID),
		LastUpdate:                   str.Ptr(lastUpdate),
		Visibility:                   str.Ptr(publicVisibility),
		ReleaseStatus:                str.Ptr(releaseStatus),
		Disabled:                     &disabled,
		DeprecationDate:              str.Ptr(deprecationDate),
		SunsetDate:                   str.Ptr(sunsetDate),
		Successors:                   json.RawMessage("[]"),
		ChangeLogEntries:             json.RawMessage("[]"),
		Type:                         dataProductType,
		Category:                     category,
		EntityTypes:                  json.RawMessage("[]"),
		InputPorts:                   json.RawMessage("[]"),
		OutputPorts:                  json.RawMessage(outputPorts),
		Responsible:                  &responsible,
		DataProductLinks:             json.RawMessage("[]"),
		Links:                        json.RawMessage("[]"),
		Industry:                     json.RawMessage("[]"),
		LineOfBusiness:               json.RawMessage("[]"),
		Tags:                         json.RawMessage("[]"),
		Labels:                       json.RawMessage("[]"),
		DocumentationLabels:          json.RawMessage("[]"),
		PolicyLevel:                  nil,
		CustomPolicyLevel:            nil,
		SystemInstanceAware:          &systemInstanceAware,
		ResourceHash:                 str.Ptr(resourceHash),
		Version: &model.Version{
			Value:           versionValue,
			Deprecated:      &versionDeprecated,
			DeprecatedSince: &versionDeprecatedSince,
			ForRemoval:      &versionForRemoval,
		},
	}
}

func fixDataProductEntity(dataProductID, appID string) *dataproduct.Entity {
	return &dataproduct.Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        dataProductID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
		ApplicationID:                repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID: repo.NewValidNullableString(appTemplateVersionID),
		OrdID:                        repo.NewValidNullableString(dataProductOrdID),
		LocalTenantID:                repo.NewValidNullableString(localTenantID),
		CorrelationIDs:               repo.NewValidNullableString("[]"),
		Title:                        title,
		ShortDescription:             repo.NewValidNullableString(shortDescription),
		Description:                  repo.NewValidNullableString(description),
		PackageID:                    repo.NewValidNullableString(packageID),
		LastUpdate:                   repo.NewValidNullableString(lastUpdate),
		Visibility:                   publicVisibility,
		ReleaseStatus:                repo.NewValidNullableString(releaseStatus),
		Disabled:                     repo.NewValidNullableBool(disabled),
		DeprecationDate:              repo.NewValidNullableString(deprecationDate),
		SunsetDate:                   repo.NewValidNullableString(sunsetDate),
		Successors:                   repo.NewValidNullableString("[]"),
		ChangeLogEntries:             repo.NewValidNullableString("[]"),
		Type:                         dataProductType,
		Category:                     category,
		EntityTypes:                  repo.NewValidNullableString("[]"),
		InputPorts:                   repo.NewValidNullableString("[]"),
		OutputPorts:                  repo.NewValidNullableString(outputPorts),
		Responsible:                  repo.NewValidNullableString(responsible),
		DataProductLinks:             repo.NewValidNullableString("[]"),
		Links:                        repo.NewValidNullableString("[]"),
		Industry:                     repo.NewValidNullableString("[]"),
		LineOfBusiness:               repo.NewValidNullableString("[]"),
		Tags:                         repo.NewValidNullableString("[]"),
		Labels:                       repo.NewValidNullableString("[]"),
		DocumentationLabels:          repo.NewValidNullableString("[]"),
		SystemInstanceAware:          repo.NewValidNullableBool(systemInstanceAware),
		ResourceHash:                 repo.NewValidNullableString(resourceHash),
		Version: version.Version{
			Value:           repo.NewValidNullableString(versionValue),
			Deprecated:      repo.NewValidNullableBool(versionDeprecated),
			DeprecatedSince: repo.NewValidNullableString(versionDeprecatedSince),
			ForRemoval:      repo.NewValidNullableBool(versionForRemoval),
		},
	}
}

func fixDataProductColumns() []string {
	return []string{"id", "app_id", "app_template_version_id", "ord_id", "local_tenant_id", "correlation_ids", "title", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "disabled", "deprecation_date", "sunset_date", "successors", "changelog_entries", "type", "category", "entity_types", "input_ports", "output_ports", "responsible", "data_product_links",
		"links", "industry", "line_of_business", "tags", "labels", "documentation_labels", "policy_level", "custom_policy_level", "system_instance_aware",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
}

func fixDataProductRow(id, appID string) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(appTemplateVersionID), dataProductOrdID, localTenantID, repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		title, repo.NewValidNullableString(shortDescription), repo.NewValidNullableString(description), packageID, repo.NewValidNullableString(lastUpdate), publicVisibility, releaseStatus,
		repo.NewValidNullableBool(disabled), repo.NewValidNullableString(deprecationDate), repo.NewValidNullableString(sunsetDate), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), dataProductType, category, repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage(outputPorts)), repo.NewValidNullableString(responsible), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), nil, nil, &systemInstanceAware,
		repo.NewNullableString(&versionValue), repo.NewNullableBool(&versionDeprecated), repo.NewNullableString(&versionDeprecatedSince), repo.NewNullableBool(&versionForRemoval),
		ready, fixedTimestamp, time.Time{}, time.Time{}, nil, resourceHash,
	}
}

func fixDataProductCreateArgs(id string, dataProduct *model.DataProduct) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(*dataProduct.ApplicationTemplateVersionID), dataProduct.OrdID, dataProduct.LocalTenantID, repo.NewNullableStringFromJSONRawMessage(dataProduct.CorrelationIDs), dataProduct.Title,
		repo.NewValidNullableString(*dataProduct.ShortDescription), repo.NewValidNullableString(*dataProduct.Description), dataProduct.PackageID, repo.NewValidNullableString(*dataProduct.LastUpdate),
		dataProduct.Visibility, dataProduct.ReleaseStatus, dataProduct.Disabled, repo.NewValidNullableString(*dataProduct.DeprecationDate), repo.NewValidNullableString(*dataProduct.SunsetDate), repo.NewNullableStringFromJSONRawMessage(dataProduct.Successors),
		repo.NewNullableStringFromJSONRawMessage(dataProduct.ChangeLogEntries), dataProduct.Type, dataProduct.Category, repo.NewNullableStringFromJSONRawMessage(dataProduct.EntityTypes), repo.NewNullableStringFromJSONRawMessage(dataProduct.InputPorts),
		repo.NewNullableStringFromJSONRawMessage(dataProduct.OutputPorts), dataProduct.Responsible, repo.NewNullableStringFromJSONRawMessage(dataProduct.DataProductLinks), repo.NewNullableStringFromJSONRawMessage(dataProduct.Links), repo.NewNullableStringFromJSONRawMessage(dataProduct.Industry),
		repo.NewNullableStringFromJSONRawMessage(dataProduct.LineOfBusiness), repo.NewNullableStringFromJSONRawMessage(dataProduct.Tags), repo.NewNullableStringFromJSONRawMessage(dataProduct.Labels), repo.NewNullableStringFromJSONRawMessage(dataProduct.DocumentationLabels),
		nil, nil, dataProduct.SystemInstanceAware, repo.NewNullableString(&dataProduct.Version.Value), repo.NewNullableBool(dataProduct.Version.Deprecated), repo.NewNullableString(dataProduct.Version.DeprecatedSince), repo.NewNullableBool(dataProduct.Version.ForRemoval), ready, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(*dataProduct.ResourceHash),
	}
}

func fixDataProductUpdateArgs(dataProduct *dataproduct.Entity) []driver.Value {
	return []driver.Value{dataProduct.OrdID, dataProduct.LocalTenantID, dataProduct.CorrelationIDs, dataProduct.Title,
		dataProduct.ShortDescription, dataProduct.Description, dataProduct.PackageID, dataProduct.LastUpdate,
		dataProduct.Visibility, dataProduct.ReleaseStatus, dataProduct.Disabled, dataProduct.DeprecationDate, dataProduct.SunsetDate, dataProduct.Successors,
		dataProduct.ChangeLogEntries, dataProduct.Type, dataProduct.Category, dataProduct.EntityTypes, dataProduct.InputPorts,
		dataProduct.OutputPorts, dataProduct.Responsible, dataProduct.DataProductLinks, dataProduct.Links, dataProduct.Industry,
		dataProduct.LineOfBusiness, dataProduct.Tags, dataProduct.Labels, dataProduct.DocumentationLabels,
		dataProduct.PolicyLevel, dataProduct.CustomPolicyLevel, dataProduct.SystemInstanceAware, dataProduct.Version.Value, dataProduct.Version.Deprecated,
		dataProduct.Version.DeprecatedSince, dataProduct.Version.ForRemoval, ready, dataProduct.CreatedAt, dataProduct.UpdatedAt, dataProduct.DeletedAt, dataProduct.Error, dataProduct.ResourceHash, dataProduct.ID,
	}
}
