package entitytype

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// VersionConverter converts Version between model.Version, graphql.Version and repo-layer version.Version
//
//go:generate mockery --name=VersionConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

type converter struct {
	vc VersionConverter
}

// NewConverter missing godoc
func NewConverter(vc VersionConverter) *converter {
	return &converter{vc: vc}
}

// ToEntity missing godoc
func (c *converter) ToEntity(entityModel *model.EntityType) *Entity {
	if entityModel == nil {
		return nil
	}

	visibility := entityModel.Visibility
	if visibility == "" {
		visibility = "public"
	}

	output := &Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        entityModel.ID,
			Ready:     entityModel.Ready,
			CreatedAt: entityModel.CreatedAt,
			UpdatedAt: entityModel.UpdatedAt,
			DeletedAt: entityModel.DeletedAt,
			Error:     repo.NewNullableString(entityModel.Error),
		},
		ApplicationID:                repo.NewNullableString(entityModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(entityModel.ApplicationTemplateVersionID),
		OrdID:                        entityModel.OrdID,
		LocalID:                      entityModel.LocalID,
		CorrelationIDs:               repo.NewNullableStringFromJSONRawMessage(entityModel.CorrelationIDs),
		Level:                        entityModel.Level,
		Title:                        entityModel.Title,
		ShortDescription:             repo.NewNullableString(entityModel.ShortDescription),
		Description:                  repo.NewNullableString(entityModel.Description),
		SystemInstanceAware:          repo.NewNullableBool(entityModel.SystemInstanceAware),
		ChangeLogEntries:             repo.NewNullableStringFromJSONRawMessage(entityModel.ChangeLogEntries),
		PackageID:                    entityModel.PackageID,
		Visibility:                   visibility,
		Links:                        repo.NewNullableStringFromJSONRawMessage(entityModel.Links),
		PartOfProducts:               repo.NewNullableStringFromJSONRawMessage(entityModel.PartOfProducts),
		LastUpdate:                   repo.NewNullableString(entityModel.LastUpdate),
		PolicyLevel:                  repo.NewNullableString(entityModel.PolicyLevel),
		CustomPolicyLevel:            repo.NewNullableString(entityModel.CustomPolicyLevel),
		ReleaseStatus:                entityModel.ReleaseStatus,
		SunsetDate:                   repo.NewNullableString(entityModel.SunsetDate),
		DeprecationDate:              repo.NewNullableString(entityModel.DeprecationDate),
		Successors:                   repo.NewNullableStringFromJSONRawMessage(entityModel.Successors),
		Extensible:                   repo.NewNullableStringFromJSONRawMessage(entityModel.Extensible),
		Tags:                         repo.NewNullableStringFromJSONRawMessage(entityModel.Tags),
		Labels:                       repo.NewNullableStringFromJSONRawMessage(entityModel.Labels),
		DocumentationLabels:          repo.NewNullableStringFromJSONRawMessage(entityModel.DocumentationLabels),
		Version:                      c.convertVersionToEntity(entityModel.Version),
		ResourceHash:                 repo.NewNullableString(entityModel.ResourceHash),
	}

	return output
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.vc.ToEntity(*inVer)
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) *model.EntityType {
	if entity == nil {
		return nil
	}

	output := &model.EntityType{
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		OrdID:                        entity.OrdID,
		LocalID:                      entity.LocalID,
		CorrelationIDs:               repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		Level:                        entity.Level,
		Title:                        entity.Title,
		ShortDescription:             repo.StringPtrFromNullableString(entity.ShortDescription),
		Description:                  repo.StringPtrFromNullableString(entity.Description),
		SystemInstanceAware:          repo.BoolPtrFromNullableBool(entity.SystemInstanceAware),
		ChangeLogEntries:             repo.JSONRawMessageFromNullableString(entity.ChangeLogEntries),
		PackageID:                    entity.PackageID,
		Visibility:                   entity.Visibility,
		Links:                        repo.JSONRawMessageFromNullableString(entity.Links),
		PartOfProducts:               repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LastUpdate:                   repo.StringPtrFromNullableString(entity.LastUpdate),
		PolicyLevel:                  repo.StringPtrFromNullableString(entity.PolicyLevel),
		CustomPolicyLevel:            repo.StringPtrFromNullableString(entity.CustomPolicyLevel),
		ReleaseStatus:                entity.ReleaseStatus,
		SunsetDate:                   repo.StringPtrFromNullableString(entity.SunsetDate),
		DeprecationDate:              repo.StringPtrFromNullableString(entity.DeprecationDate),
		Successors:                   repo.JSONRawMessageFromNullableString(entity.Successors),
		Extensible:                   repo.JSONRawMessageFromNullableString(entity.Extensible),
		Tags:                         repo.JSONRawMessageFromNullableString(entity.Tags),
		Labels:                       repo.JSONRawMessageFromNullableString(entity.Labels),
		DocumentationLabels:          repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		Version:                      c.vc.FromEntity(entity.Version),
		ResourceHash:                 repo.StringPtrFromNullableString(entity.ResourceHash),
	}
	return output
}
