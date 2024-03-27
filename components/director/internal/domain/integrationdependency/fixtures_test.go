package integrationdependency_test

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
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
	defaultLabelsRaw        = `{"displayName":"bar","test":["val","val2"]}`
)

var (
	fixedTimestamp           = time.Now()
	appID                    = "aaaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	appTemplateVersionID     = "fffffffff-ffff-aaaa-ffff-aaaaaaaaaaaa"
	mandatory                = false
	ready                    = true
	supportMultipleProviders = true
	versionValue             = "v1.1"
	versionDeprecated        = false
	versionDeprecatedSince   = "v1.0"
	versionForRemoval        = false
	testErr                  = errors.New("test error")
	defaultLabelsModel       = graphql.Labels{
		"displayName": "bar",
		"test": []interface{}{
			"val",
			"val2",
		},
	}
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
		Labels:                         json.RawMessage(defaultLabelsRaw),
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

func fixGQLIntegrationDependency(id string) *graphql.IntegrationDependency {
	return &graphql.IntegrationDependency{
		Name:          title,
		Description:   str.Ptr(description),
		Mandatory:     &mandatory,
		OrdID:         str.Ptr(ordID),
		PartOfPackage: str.Ptr(packageID),
		Visibility:    str.Ptr(publicVisibility),
		ReleaseStatus: str.Ptr(releaseStatus),
		Aspects:       []*graphql.Aspect{},
		Version: &graphql.Version{
			Value:           versionValue,
			Deprecated:      &versionDeprecated,
			DeprecatedSince: &versionDeprecatedSince,
			ForRemoval:      &versionForRemoval,
		},
		Labels: &defaultLabelsModel,
		BaseEntity: &graphql.BaseEntity{
			ID:        id,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
		},
	}
}

func fixIntegrationDependencyEntity(integrationDependencyID, appID string) *integrationdependency.Entity {
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
		Labels:                         repo.NewValidNullableString(defaultLabelsRaw),
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

func fixIntegrationDependencyInputModelWithPackageOrdID(packageOrdID string) model.IntegrationDependencyInput {
	return model.IntegrationDependencyInput{
		OrdID:            str.Ptr(ordID),
		LocalTenantID:    str.Ptr(localTenantID),
		CorrelationIDs:   json.RawMessage("[]"),
		Title:            title,
		ShortDescription: str.Ptr(shortDescription),
		Description:      str.Ptr(description),
		OrdPackageID:     str.Ptr(packageOrdID),
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
				EventResources: []*model.AspectEventResourceInput{
					{
						OrdID:      ordID,
						MinVersion: str.Ptr("1.0.0"),
						Subset:     json.RawMessage("[]"),
					},
				},
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

func fixGQLIntegrationDependencyInputWithPackageAndWithoutProperties(packageOrdID string) *graphql.IntegrationDependencyInput {
	return &graphql.IntegrationDependencyInput{
		Name:          title,
		OrdID:         ordID,
		PartOfPackage: str.Ptr(packageID),
		Description:   str.Ptr(description),
		Aspects: []*graphql.AspectInput{
			{
				Name:           title,
				Description:    str.Ptr(description),
				Mandatory:      &mandatory,
				APIResources:   []*graphql.AspectAPIDefinitionInput{},
				EventResources: []*graphql.AspectEventDefinitionInput{},
			},
		},
	}
}

func fixGQLIntegrationDependencyInputWithPackageAndWithProperties(appNamespace, packageOrdID string) *graphql.IntegrationDependencyInput {
	return &graphql.IntegrationDependencyInput{
		Name:          title,
		OrdID:         ordID,
		Mandatory:     &mandatory,
		Visibility:    str.Ptr(publicVisibility),
		ReleaseStatus: str.Ptr(releaseStatus),
		PartOfPackage: str.Ptr(packageID),
		Description:   str.Ptr(description),
		Aspects: []*graphql.AspectInput{
			{
				Name:           title,
				Description:    str.Ptr(description),
				Mandatory:      &mandatory,
				APIResources:   []*graphql.AspectAPIDefinitionInput{},
				EventResources: []*graphql.AspectEventDefinitionInput{},
			},
		},
	}
}

func fixGQLIntegrationDependencyWithGeneratedProperties(appNamespace, aspectID, packageOrdID string) *graphql.IntegrationDependency {
	return &graphql.IntegrationDependency{
		Name:          title,
		Description:   str.Ptr(description),
		Mandatory:     &mandatory,
		OrdID:         str.Ptr(ordID),
		PartOfPackage: str.Ptr(packageID),
		Visibility:    str.Ptr(publicVisibility),
		ReleaseStatus: str.Ptr(releaseStatus),
		Aspects: []*graphql.Aspect{
			{
				Name:           title,
				Description:    str.Ptr(description),
				Mandatory:      &mandatory,
				APIResources:   []*graphql.AspectAPIDefinition{},
				EventResources: []*graphql.AspectEventDefinition{},
				BaseEntity: &graphql.BaseEntity{
					ID:        aspectID,
					Ready:     true,
					Error:     nil,
					CreatedAt: timeToTimestampPtr(fixedTimestamp),
					UpdatedAt: timeToTimestampPtr(time.Time{}),
					DeletedAt: timeToTimestampPtr(time.Time{}),
				},
			},
		},
		Version: &graphql.Version{
			Value:           versionValue,
			Deprecated:      &versionDeprecated,
			DeprecatedSince: &versionDeprecatedSince,
			ForRemoval:      &versionForRemoval,
		},
		BaseEntity: &graphql.BaseEntity{
			ID:        integrationDependencyID,
			Ready:     true,
			Error:     nil,
			CreatedAt: timeToTimestampPtr(fixedTimestamp),
			UpdatedAt: timeToTimestampPtr(time.Time{}),
			DeletedAt: timeToTimestampPtr(time.Time{}),
		},
	}
}
func fixIntegrationDependencyInputModelWithoutPackage() model.IntegrationDependencyInput {
	intDep := fixIntegrationDependencyInputModelWithPackageOrdID("")
	intDep.OrdPackageID = nil
	return intDep
}

func fixGQLIntegrationDependencyInput() *graphql.IntegrationDependencyInput {
	return &graphql.IntegrationDependencyInput{
		Name:          title,
		OrdID:         ordID,
		PartOfPackage: str.Ptr(packageID),
		Visibility:    str.Ptr(publicVisibility),
		ReleaseStatus: str.Ptr(releaseStatus),
		Description:   str.Ptr(description),
		Mandatory:     &mandatory,
		Aspects: []*graphql.AspectInput{
			{
				Name:           title,
				Description:    str.Ptr(description),
				Mandatory:      &mandatory,
				APIResources:   []*graphql.AspectAPIDefinitionInput{},
				EventResources: []*graphql.AspectEventDefinitionInput{},
			},
		},
	}
}

func fixGQLIntegrationDependencyInputWithoutPackage() *graphql.IntegrationDependencyInput {
	intDep := fixGQLIntegrationDependencyInput()
	intDep.PartOfPackage = nil
	return intDep
}

func fixGQLIntegrationDependencyInputWithPackageOrdID(packageOrdID string) *graphql.IntegrationDependencyInput {
	intDep := fixGQLIntegrationDependencyInput()
	intDep.PartOfPackage = str.Ptr(packageID)
	return intDep
}

func fixIntegrationDependenciesColumns() []string {
	return []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "app_id", "app_template_version_id", "ord_id", "local_tenant_id",
		"correlation_ids", "title", "short_description", "description", "package_id", "visibility",
		"last_update", "release_status", "sunset_date", "successors", "mandatory", "related_integration_dependencies", "links", "tags", "labels",
		"documentation_labels", "resource_hash", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
}

func fixIntegrationDependenciesRow(id, appID string) []driver.Value {
	return []driver.Value{id, ready, fixedTimestamp, time.Time{}, time.Time{}, nil, appID, repo.NewValidNullableString(appTemplateVersionID), ordID, localTenantID,
		repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), title, repo.NewValidNullableString(shortDescription), repo.NewValidNullableString(description), packageID, publicVisibility,
		repo.NewValidNullableString(lastUpdate), releaseStatus, repo.NewValidNullableString(sunsetDate), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableBool(&mandatory), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage("[]")), repo.NewNullableStringFromJSONRawMessage(json.RawMessage(defaultLabelsRaw)),
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
func timeToTimestampPtr(time time.Time) *graphql.Timestamp {
	t := graphql.Timestamp(time)
	return &t
}
