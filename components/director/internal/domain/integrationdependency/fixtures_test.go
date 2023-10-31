package integrationdependency_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"time"
)

const (
	integrationDependencyID = "integrationDependencyID"
	tenantID                = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	externalTenantID        = "external-tnt"
	description             = "description"
	shortDescription        = "short desc"
	title                   = "title"
	ordID                   = "ordID"
	localTenantID           = "localTenantID"
	packageID               = "packageID"
	publicVisibility        = "public"
	releaseStatus           = "active"
	sunsetDate              = "2022-01-08T15:47:04+00:00"
	lastUpdate              = "2022-01-08T15:47:04+00:00"
	resourceHash            = "123456"
)

var (
	fixedTimestamp           = time.Now()
	appID                    = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID     = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
	mandatory                = true
	ready                    = true
	supportMultipleProviders = true
	versionValue             = "v1.1"
	versionDeprecated        = false
	versionDeprecatedSince   = "v1.0"
	versionForRemoval        = false
	testErr                  = errors.New("test error")
)

func fixIntegrationDependencyModel(integrationDependencyID string) *model.IntegrationDependency {
	return &model.IntegrationDependency{
		BaseEntity: &model.BaseEntity{
			ID:        integrationDependencyID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
		ApplicationID:                  &appID,
		ApplicationTemplateVersionID:   &appTemplateVersionID,
		OrdID:                          str.Ptr(ordID),
		LocalTenantID:                  str.Ptr(localTenantID),
		CorrelationIDs:                 json.RawMessage("[]"),
		Title:                          title,
		ShortDescription:               str.Ptr(shortDescription),
		Description:                    str.Ptr(description),
		PackageID:                      str.Ptr(packageID),
		Visibility:                     publicVisibility,
		LastUpdate:                     str.Ptr(lastUpdate),
		ReleaseStatus:                  str.Ptr(releaseStatus),
		SunsetDate:                     str.Ptr(sunsetDate),
		Successors:                     json.RawMessage("[]"),
		RelatedIntegrationDependencies: json.RawMessage("[]"),
		Mandatory:                      &mandatory,
		Links:                          json.RawMessage("[]"),
		Tags:                           json.RawMessage("[]"),
		Labels:                         json.RawMessage("[]"),
		DocumentationLabels:            json.RawMessage("[]"),
		ResourceHash:                   str.Ptr(resourceHash),
		Version: &model.Version{
			Value:           versionValue,
			Deprecated:      &versionDeprecated,
			DeprecatedSince: &versionDeprecatedSince,
			ForRemoval:      &versionForRemoval,
		},
	}
}

func fixIntegrationDependencyEntity(integrationDependencyID string) *integrationdependency.Entity {
	return &integrationdependency.Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        integrationDependencyID,
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     sql.NullString{},
		},
		ApplicationID:                  repo.NewValidNullableString(appID),
		ApplicationTemplateVersionID:   repo.NewValidNullableString(appTemplateVersionID),
		OrdID:                          repo.NewValidNullableString(ordID),
		LocalTenantID:                  repo.NewValidNullableString(localTenantID),
		CorrelationIDs:                 repo.NewValidNullableString("[]"),
		Title:                          title,
		ShortDescription:               repo.NewValidNullableString(shortDescription),
		Description:                    repo.NewValidNullableString(description),
		PackageID:                      repo.NewValidNullableString(packageID),
		Visibility:                     publicVisibility,
		LastUpdate:                     repo.NewValidNullableString(lastUpdate),
		ReleaseStatus:                  repo.NewValidNullableString(releaseStatus),
		SunsetDate:                     repo.NewValidNullableString(sunsetDate),
		Successors:                     repo.NewValidNullableString("[]"),
		RelatedIntegrationDependencies: repo.NewValidNullableString("[]"),
		Mandatory:                      repo.NewValidNullableBool(mandatory),
		Links:                          repo.NewValidNullableString("[]"),
		Tags:                           repo.NewValidNullableString("[]"),
		Labels:                         repo.NewValidNullableString("[]"),
		DocumentationLabels:            repo.NewValidNullableString("[]"),
		ResourceHash:                   repo.NewValidNullableString(resourceHash),
		Version: version.Version{
			Value:           repo.NewValidNullableString(versionValue),
			Deprecated:      repo.NewValidNullableBool(versionDeprecated),
			DeprecatedSince: repo.NewValidNullableString(versionDeprecatedSince),
			ForRemoval:      repo.NewValidNullableBool(versionForRemoval),
		},
	}
}

func fixIntegrationDependencyInputModel() model.IntegrationDependencyInput {
	return model.IntegrationDependencyInput{
		OrdID:            str.Ptr(ordID),
		LocalTenantID:    str.Ptr(localTenantID),
		CorrelationIDs:   json.RawMessage("[]"),
		Title:            title,
		ShortDescription: str.Ptr(shortDescription),
		Description:      str.Ptr(description),
		OrdPackageID:     str.Ptr(packageID),
		Visibility:       publicVisibility,
		LastUpdate:       str.Ptr(lastUpdate),
		ReleaseStatus:    str.Ptr(releaseStatus),
		SunsetDate:       str.Ptr(sunsetDate),
		Mandatory:        &mandatory,
		Aspects: []*model.AspectInput{
			{
				Title:                    title,
				Description:              str.Ptr(description),
				Mandatory:                &mandatory,
				SupportMultipleProviders: &supportMultipleProviders,
				APIResources:             json.RawMessage("[]"),
				EventResources:           json.RawMessage("[]"),
			},
		},
		RelatedIntegrationDependencies: json.RawMessage("[]"),
		Successors:                     json.RawMessage("[]"),
		Links:                          json.RawMessage("[]"),
		Tags:                           json.RawMessage("[]"),
		Labels:                         json.RawMessage("[]"),
		DocumentationLabels:            json.RawMessage("[]"),
	}
}

func fixIntegrationDependenciesColumns() []string {
	return []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "app_id", "app_template_version_id", "ord_id", "local_tenant_id",
		"correlation_ids", "title", "short_description", "description", "package_id", "visibility",
		"last_update", "release_status", "sunset_date", "successors", "mandatory", "related_integration_dependencies", "links", "tags", "labels",
		"documentation_labels", "resource_hash", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
}

func fixIntegrationDependenciesRow(id string) []driver.Value {
	return []driver.Value{id, ready, fixedTimestamp, time.Time{}, time.Time{}, nil, appID, repo.NewValidNullableString(appTemplateVersionID), ordID, localTenantID,
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), title, repo.NewValidNullableString(shortDescription), repo.NewValidNullableString(description), packageID, publicVisibility,
		repo.NewValidNullableString(lastUpdate), releaseStatus, repo.NewValidNullableString(sunsetDate), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableBool(&mandatory), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")),
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewValidNullableString(resourceHash), repo.NewNullableString(&versionValue), repo.NewNullableBool(&versionDeprecated), repo.NewNullableString(&versionDeprecatedSince), repo.NewNullableBool(&versionForRemoval)}
}

func fixIntegrationDependencyCreateArgs(id string, integrationDependency *model.IntegrationDependency) []driver.Value {
	return []driver.Value{id, appID, repo.NewValidNullableString(*integrationDependency.ApplicationTemplateVersionID), integrationDependency.OrdID, integrationDependency.LocalTenantID,
		repo.NewNullableStringFromJSONRawMessage(integrationDependency.CorrelationIDs), integrationDependency.Title, repo.NewValidNullableString(*integrationDependency.ShortDescription), repo.NewValidNullableString(*integrationDependency.Description), integrationDependency.PackageID, repo.NewValidNullableString(*integrationDependency.LastUpdate),
		integrationDependency.Visibility, integrationDependency.ReleaseStatus, repo.NewValidNullableString(*integrationDependency.SunsetDate), repo.NewNullableStringFromJSONRawMessage(integrationDependency.Successors), repo.NewNullableBool(integrationDependency.Mandatory), repo.NewNullableStringFromJSONRawMessage(integrationDependency.RelatedIntegrationDependencies), repo.NewNullableStringFromJSONRawMessage(integrationDependency.Links), repo.NewNullableStringFromJSONRawMessage(integrationDependency.Tags), repo.NewNullableStringFromJSONRawMessage(integrationDependency.Labels),
		repo.NewNullableStringFromJSONRawMessage(integrationDependency.DocumentationLabels), repo.NewNullableString(&integrationDependency.Version.Value), repo.NewNullableBool(integrationDependency.Version.Deprecated), repo.NewNullableString(integrationDependency.Version.DeprecatedSince), repo.NewNullableBool(integrationDependency.Version.ForRemoval), ready, fixedTimestamp, time.Time{}, time.Time{}, nil, repo.NewValidNullableString(*integrationDependency.ResourceHash)}
}

func fixIntegrationDependencyUpdateArgs(integrationDependency *integrationdependency.Entity) []driver.Value {
	return []driver.Value{integrationDependency.OrdID, integrationDependency.LocalTenantID,
		integrationDependency.CorrelationIDs, integrationDependency.Title, integrationDependency.ShortDescription, integrationDependency.Description, integrationDependency.PackageID, integrationDependency.LastUpdate, integrationDependency.Visibility,
		integrationDependency.ReleaseStatus, integrationDependency.SunsetDate, integrationDependency.Successors, integrationDependency.Mandatory, integrationDependency.RelatedIntegrationDependencies, integrationDependency.Links, integrationDependency.Tags, integrationDependency.Labels,
		integrationDependency.DocumentationLabels, integrationDependency.Version.Value, integrationDependency.Version.Deprecated, integrationDependency.Version.DeprecatedSince, integrationDependency.Version.ForRemoval, integrationDependency.Ready, integrationDependency.CreatedAt, integrationDependency.UpdatedAt, integrationDependency.DeletedAt, integrationDependency.Error, integrationDependency.ResourceHash, integrationDependency.ID}
}
