package capability

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass Capability.
func NewConverter(version VersionConverter) *converter {
	return &converter{version: version}
}

// FromEntity converts the provided Entity repo-layer representation of a Capability to the service-layer representation model.Capability.
func (c *converter) FromEntity(entity *Entity) *model.Capability {
	return &model.Capability{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		PackageID:                    repo.StringPtrFromNullableString(entity.PackageID),
		Name:                         entity.Name,
		Description:                  repo.StringPtrFromNullableString(entity.Description),
		OrdID:                        repo.StringPtrFromNullableString(entity.OrdID),
		Type:                         entity.Type,
		CustomType:                   repo.StringPtrFromNullableString(entity.CustomType),
		LocalTenantID:                repo.StringPtrFromNullableString(entity.LocalTenantID),
		ShortDescription:             repo.StringPtrFromNullableString(entity.ShortDescription),
		SystemInstanceAware:          repo.BoolPtrFromNullableBool(entity.SystemInstanceAware),
		Tags:                         repo.JSONRawMessageFromNullableString(entity.Tags),
		RelatedEntityTypes:           repo.JSONRawMessageFromNullableString(entity.RelatedEntityTypes),
		Links:                        repo.JSONRawMessageFromNullableString(entity.Links),
		ReleaseStatus:                repo.StringPtrFromNullableString(entity.ReleaseStatus),
		Labels:                       repo.JSONRawMessageFromNullableString(entity.Labels),
		Visibility:                   &entity.Visibility,
		Version:                      c.version.FromEntity(entity.Version),
		ResourceHash:                 repo.StringPtrFromNullableString(entity.ResourceHash),
		DocumentationLabels:          repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		CorrelationIDs:               repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		LastUpdate:                   repo.StringPtrFromNullableString(entity.LastUpdate),
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

// ToEntity converts the provided service-layer representation of a Capability to the repository-layer one.
func (c *converter) ToEntity(capabilityModel *model.Capability) *Entity {
	visibility := capabilityModel.Visibility
	if visibility == nil {
		visibility = str.Ptr("public")
	}

	return &Entity{
		ApplicationID:                repo.NewNullableString(capabilityModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(capabilityModel.ApplicationTemplateVersionID),
		PackageID:                    repo.NewNullableString(capabilityModel.PackageID),
		Name:                         capabilityModel.Name,
		Description:                  repo.NewNullableString(capabilityModel.Description),
		OrdID:                        repo.NewNullableString(capabilityModel.OrdID),
		Type:                         capabilityModel.Type,
		CustomType:                   repo.NewNullableString(capabilityModel.CustomType),
		LocalTenantID:                repo.NewNullableString(capabilityModel.LocalTenantID),
		ShortDescription:             repo.NewNullableString(capabilityModel.ShortDescription),
		SystemInstanceAware:          repo.NewNullableBool(capabilityModel.SystemInstanceAware),
		Tags:                         repo.NewNullableStringFromJSONRawMessage(capabilityModel.Tags),
		RelatedEntityTypes:           repo.NewNullableStringFromJSONRawMessage(capabilityModel.RelatedEntityTypes),
		Links:                        repo.NewNullableStringFromJSONRawMessage(capabilityModel.Links),
		ReleaseStatus:                repo.NewNullableString(capabilityModel.ReleaseStatus),
		Labels:                       repo.NewNullableStringFromJSONRawMessage(capabilityModel.Labels),
		Visibility:                   *visibility,
		Version:                      c.convertVersionToEntity(capabilityModel.Version),
		ResourceHash:                 repo.NewNullableString(capabilityModel.ResourceHash),
		DocumentationLabels:          repo.NewNullableStringFromJSONRawMessage(capabilityModel.DocumentationLabels),
		CorrelationIDs:               repo.NewNullableStringFromJSONRawMessage(capabilityModel.CorrelationIDs),
		LastUpdate:                   repo.NewNullableString(capabilityModel.LastUpdate),
		BaseEntity: &repo.BaseEntity{
			ID:        capabilityModel.ID,
			Ready:     capabilityModel.Ready,
			CreatedAt: capabilityModel.CreatedAt,
			UpdatedAt: capabilityModel.UpdatedAt,
			DeletedAt: capabilityModel.DeletedAt,
			Error:     repo.NewNullableString(capabilityModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}
