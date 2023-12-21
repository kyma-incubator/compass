package dataproduct

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

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass Data Product.
func NewConverter(version VersionConverter) *converter {
	return &converter{version: version}
}

// FromEntity converts the provided Entity repo-layer representation of a Data Product to the service-layer representation model.DataProduct.
func (c *converter) FromEntity(entity *Entity) *model.DataProduct {
	if entity == nil {
		return nil
	}

	return &model.DataProduct{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		OrdID:                        repo.StringPtrFromNullableString(entity.OrdID),
		LocalTenantID:                repo.StringPtrFromNullableString(entity.LocalTenantID),
		CorrelationIDs:               repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		Title:                        entity.Title,
		ShortDescription:             repo.StringPtrFromNullableString(entity.ShortDescription),
		Description:                  repo.StringPtrFromNullableString(entity.Description),
		PackageID:                    repo.StringPtrFromNullableString(entity.PackageID),
		Version:                      c.version.FromEntity(entity.Version),
		LastUpdate:                   repo.StringPtrFromNullableString(entity.LastUpdate),
		Visibility:                   &entity.Visibility,
		ReleaseStatus:                repo.StringPtrFromNullableString(entity.ReleaseStatus),
		Disabled:                     repo.BoolPtrFromNullableBool(entity.Disabled),
		DeprecationDate:              repo.StringPtrFromNullableString(entity.DeprecationDate),
		SunsetDate:                   repo.StringPtrFromNullableString(entity.SunsetDate),
		Successors:                   repo.JSONRawMessageFromNullableString(entity.Successors),
		ChangeLogEntries:             repo.JSONRawMessageFromNullableString(entity.ChangeLogEntries),
		Type:                         entity.Type,
		Category:                     entity.Category,
		EntityTypes:                  repo.JSONRawMessageFromNullableString(entity.EntityTypes),
		InputPorts:                   repo.JSONRawMessageFromNullableString(entity.InputPorts),
		OutputPorts:                  repo.JSONRawMessageFromNullableString(entity.OutputPorts),
		Responsible:                  repo.StringPtrFromNullableString(entity.Responsible),
		DataProductLinks:             repo.JSONRawMessageFromNullableString(entity.DataProductLinks),
		Links:                        repo.JSONRawMessageFromNullableString(entity.Links),
		Industry:                     repo.JSONRawMessageFromNullableString(entity.Industry),
		LineOfBusiness:               repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Tags:                         repo.JSONRawMessageFromNullableString(entity.Tags),
		Labels:                       repo.JSONRawMessageFromNullableString(entity.Labels),
		DocumentationLabels:          repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
		PolicyLevel:                  repo.StringPtrFromNullableString(entity.PolicyLevel),
		CustomPolicyLevel:            repo.StringPtrFromNullableString(entity.CustomPolicyLevel),
		SystemInstanceAware:          repo.BoolPtrFromNullableBool(entity.SystemInstanceAware),
		ResourceHash:                 repo.StringPtrFromNullableString(entity.ResourceHash),
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

// ToEntity converts the provided service-layer representation of a Data Product to the repository-layer one.
func (c *converter) ToEntity(dataProductModel *model.DataProduct) *Entity {
	if dataProductModel == nil {
		return nil
	}

	visibility := dataProductModel.Visibility
	if visibility == nil {
		visibility = str.Ptr("public")
	}

	return &Entity{
		ApplicationID:                repo.NewNullableString(dataProductModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(dataProductModel.ApplicationTemplateVersionID),
		OrdID:                        repo.NewNullableString(dataProductModel.OrdID),
		LocalTenantID:                repo.NewNullableString(dataProductModel.LocalTenantID),
		CorrelationIDs:               repo.NewNullableStringFromJSONRawMessage(dataProductModel.CorrelationIDs),
		Title:                        dataProductModel.Title,
		ShortDescription:             repo.NewNullableString(dataProductModel.ShortDescription),
		Description:                  repo.NewNullableString(dataProductModel.Description),
		PackageID:                    repo.NewNullableString(dataProductModel.PackageID),
		LastUpdate:                   repo.NewNullableString(dataProductModel.LastUpdate),
		Visibility:                   *visibility,
		ReleaseStatus:                repo.NewNullableString(dataProductModel.ReleaseStatus),
		Disabled:                     repo.NewNullableBool(dataProductModel.Disabled),
		DeprecationDate:              repo.NewNullableString(dataProductModel.DeprecationDate),
		SunsetDate:                   repo.NewNullableString(dataProductModel.SunsetDate),
		Successors:                   repo.NewNullableStringFromJSONRawMessage(dataProductModel.Successors),
		ChangeLogEntries:             repo.NewNullableStringFromJSONRawMessage(dataProductModel.ChangeLogEntries),
		Type:                         dataProductModel.Type,
		Category:                     dataProductModel.Category,
		EntityTypes:                  repo.NewNullableStringFromJSONRawMessage(dataProductModel.EntityTypes),
		InputPorts:                   repo.NewNullableStringFromJSONRawMessage(dataProductModel.InputPorts),
		OutputPorts:                  repo.NewNullableStringFromJSONRawMessage(dataProductModel.OutputPorts),
		Responsible:                  repo.NewNullableString(dataProductModel.Responsible),
		DataProductLinks:             repo.NewNullableStringFromJSONRawMessage(dataProductModel.DataProductLinks),
		Links:                        repo.NewNullableStringFromJSONRawMessage(dataProductModel.Links),
		Industry:                     repo.NewNullableStringFromJSONRawMessage(dataProductModel.Industry),
		LineOfBusiness:               repo.NewNullableStringFromJSONRawMessage(dataProductModel.LineOfBusiness),
		Tags:                         repo.NewNullableStringFromJSONRawMessage(dataProductModel.Tags),
		Labels:                       repo.NewNullableStringFromJSONRawMessage(dataProductModel.Labels),
		DocumentationLabels:          repo.NewNullableStringFromJSONRawMessage(dataProductModel.DocumentationLabels),
		PolicyLevel:                  repo.NewNullableString(dataProductModel.PolicyLevel),
		CustomPolicyLevel:            repo.NewNullableString(dataProductModel.CustomPolicyLevel),
		SystemInstanceAware:          repo.NewNullableBool(dataProductModel.SystemInstanceAware),
		ResourceHash:                 repo.NewNullableString(dataProductModel.ResourceHash),
		Version:                      c.convertVersionToEntity(dataProductModel.Version),

		BaseEntity: &repo.BaseEntity{
			ID:        dataProductModel.ID,
			Ready:     dataProductModel.Ready,
			CreatedAt: dataProductModel.CreatedAt,
			UpdatedAt: dataProductModel.UpdatedAt,
			DeletedAt: dataProductModel.DeletedAt,
			Error:     repo.NewNullableString(dataProductModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}
