package integrationdependency

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// VersionConverter converts Version between model.Version, graphql.Version and repo-layer version.Version
//
//go:generate mockery --name=VersionConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type VersionConverter interface {
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

type converter struct {
	version VersionConverter
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Integration Dependency.
func NewConverter(version VersionConverter) *converter {
	return &converter{version: version}
}

// FromEntity converts the provided Entity repo-layer representation of an Integration Dependency to the service-layer representation model.IntegrationDependency.
func (c *converter) FromEntity(entity *Entity) *model.IntegrationDependency {
	return &model.IntegrationDependency{
		ApplicationID:                  repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID:   repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		OrdID:                          repo.StringPtrFromNullableString(entity.OrdID),
		LocalTenantID:                  repo.StringPtrFromNullableString(entity.LocalTenantID),
		CorrelationIDs:                 repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		Name:                           entity.Name,
		ShortDescription:               repo.StringPtrFromNullableString(entity.ShortDescription),
		Description:                    repo.StringPtrFromNullableString(entity.Description),
		PackageID:                      repo.StringPtrFromNullableString(entity.PackageID),
		Version:                        c.version.FromEntity(entity.Version),
		LastUpdate:                     repo.StringPtrFromNullableString(entity.LastUpdate),
		Visibility:                     entity.Visibility,
		ReleaseStatus:                  repo.StringPtrFromNullableString(entity.ReleaseStatus),
		SunsetDate:                     repo.StringPtrFromNullableString(entity.SunsetDate),
		Successors:                     repo.JSONRawMessageFromNullableString(entity.Successors),
		Mandatory:                      entity.Mandatory,
		RelatedIntegrationDependencies: repo.JSONRawMessageFromNullableString(entity.RelatedIntegrationDependencies),
		Links:                          repo.JSONRawMessageFromNullableString(entity.Links),
		Tags:                           repo.JSONRawMessageFromNullableString(entity.Tags),
		Labels:                         repo.JSONRawMessageFromNullableString(entity.Labels),
		DocumentationLabels:            repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		ResourceHash:                   repo.StringPtrFromNullableString(entity.ResourceHash),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}
}

// ToEntity converts the provided service-layer representation of an Integration Dependency to the repository-layer one.
func (c *converter) ToEntity(integrationDependencyModel *model.IntegrationDependency) *Entity {
	return &Entity{
		ApplicationID:                  repo.NewNullableString(integrationDependencyModel.ApplicationID),
		ApplicationTemplateVersionID:   repo.NewNullableString(integrationDependencyModel.ApplicationTemplateVersionID),
		OrdID:                          repo.NewNullableString(integrationDependencyModel.OrdID),
		LocalTenantID:                  repo.NewNullableString(integrationDependencyModel.LocalTenantID),
		CorrelationIDs:                 repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.CorrelationIDs),
		Name:                           integrationDependencyModel.Name,
		ShortDescription:               repo.NewNullableString(integrationDependencyModel.ShortDescription),
		Description:                    repo.NewNullableString(integrationDependencyModel.Description),
		Version:                        c.convertVersionToEntity(integrationDependencyModel.Version),
		PackageID:                      repo.NewNullableString(integrationDependencyModel.PackageID),
		LastUpdate:                     repo.NewNullableString(integrationDependencyModel.LastUpdate),
		Visibility:                     integrationDependencyModel.Visibility,
		ReleaseStatus:                  repo.NewNullableString(integrationDependencyModel.ReleaseStatus),
		SunsetDate:                     repo.NewNullableString(integrationDependencyModel.SunsetDate),
		Successors:                     repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.Successors),
		Mandatory:                      integrationDependencyModel.Mandatory,
		RelatedIntegrationDependencies: repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.RelatedIntegrationDependencies),
		Links:                          repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.Links),
		Tags:                           repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.Tags),
		Labels:                         repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.Labels),
		DocumentationLabels:            repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.DocumentationLabels),
		ResourceHash:                   repo.NewNullableString(integrationDependencyModel.ResourceHash),
		BaseEntity: &repo.BaseEntity{
			ID:        integrationDependencyModel.ID,
			Ready:     integrationDependencyModel.Ready,
			CreatedAt: integrationDependencyModel.CreatedAt,
			UpdatedAt: integrationDependencyModel.UpdatedAt,
			DeletedAt: integrationDependencyModel.DeletedAt,
			Error:     repo.NewNullableString(integrationDependencyModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}
