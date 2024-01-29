package integrationdependency

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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

// AspectConverter converts Aspects between the model.Aspect service-layer representation and the graphql-layer representation graphql.Aspect.
//
//go:generate mockery --name=AspectConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectConverter interface {
	MultipleToGraphQL(in []*model.Aspect, aspectEventResourcesByAspectID map[string][]*model.AspectEventResource) ([]*graphql.Aspect, error)
	MultipleInputFromGraphQL(in []*graphql.AspectInput) ([]*model.AspectInput, error)
}

type converter struct {
	version         VersionConverter
	aspectConverter AspectConverter
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Integration Dependency.
func NewConverter(version VersionConverter, aspectConverter AspectConverter) *converter {
	return &converter{version: version, aspectConverter: aspectConverter}
}

// FromEntity converts the provided Entity repo-layer representation of an Integration Dependency to the service-layer representation model.IntegrationDependency.
func (c *converter) FromEntity(entity *Entity) *model.IntegrationDependency {
	if entity == nil {
		return nil
	}

	return &model.IntegrationDependency{
		ApplicationID:                  repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID:   repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		OrdID:                          repo.StringPtrFromNullableString(entity.OrdID),
		LocalTenantID:                  repo.StringPtrFromNullableString(entity.LocalTenantID),
		CorrelationIDs:                 repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		Title:                          entity.Title,
		ShortDescription:               repo.StringPtrFromNullableString(entity.ShortDescription),
		Description:                    repo.StringPtrFromNullableString(entity.Description),
		PackageID:                      repo.StringPtrFromNullableString(entity.PackageID),
		Version:                        c.version.FromEntity(entity.Version),
		LastUpdate:                     repo.StringPtrFromNullableString(entity.LastUpdate),
		Visibility:                     entity.Visibility,
		ReleaseStatus:                  repo.StringPtrFromNullableString(entity.ReleaseStatus),
		SunsetDate:                     repo.StringPtrFromNullableString(entity.SunsetDate),
		Successors:                     repo.JSONRawMessageFromNullableString(entity.Successors),
		Mandatory:                      repo.BoolPtrFromNullableBool(entity.Mandatory),
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
	if integrationDependencyModel == nil {
		return nil
	}

	return &Entity{
		ApplicationID:                  repo.NewNullableString(integrationDependencyModel.ApplicationID),
		ApplicationTemplateVersionID:   repo.NewNullableString(integrationDependencyModel.ApplicationTemplateVersionID),
		OrdID:                          repo.NewNullableString(integrationDependencyModel.OrdID),
		LocalTenantID:                  repo.NewNullableString(integrationDependencyModel.LocalTenantID),
		CorrelationIDs:                 repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.CorrelationIDs),
		Title:                          integrationDependencyModel.Title,
		ShortDescription:               repo.NewNullableString(integrationDependencyModel.ShortDescription),
		Description:                    repo.NewNullableString(integrationDependencyModel.Description),
		Version:                        c.convertVersionToEntity(integrationDependencyModel.Version),
		PackageID:                      repo.NewNullableString(integrationDependencyModel.PackageID),
		LastUpdate:                     repo.NewNullableString(integrationDependencyModel.LastUpdate),
		Visibility:                     integrationDependencyModel.Visibility,
		ReleaseStatus:                  repo.NewNullableString(integrationDependencyModel.ReleaseStatus),
		SunsetDate:                     repo.NewNullableString(integrationDependencyModel.SunsetDate),
		Successors:                     repo.NewNullableStringFromJSONRawMessage(integrationDependencyModel.Successors),
		Mandatory:                      repo.NewNullableBool(integrationDependencyModel.Mandatory),
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

// ToGraphQL converts the provided service-layer representation of an Integration Dependency to the graphql-layer one.
func (c *converter) ToGraphQL(in *model.IntegrationDependency, aspects []*model.Aspect, aspectEventResourcesByAspectID map[string][]*model.AspectEventResource) (*graphql.IntegrationDependency, error) {
	if in == nil {
		return nil, nil
	}

	gqlAspects, err := c.aspectConverter.MultipleToGraphQL(aspects, aspectEventResourcesByAspectID)
	if err != nil {
		return nil, err
	}

	return &graphql.IntegrationDependency{
		Name:          in.Title,
		Description:   in.Description,
		OrdID:         in.OrdID,
		PartOfPackage: in.PackageID,
		Visibility:    str.Ptr(in.Visibility),
		ReleaseStatus: in.ReleaseStatus,
		Mandatory:     in.Mandatory,
		Aspects:       gqlAspects,
		Version:       c.version.ToGraphQL(in.Version),
		BaseEntity: &graphql.BaseEntity{
			ID:        in.ID,
			Ready:     in.Ready,
			CreatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.CreatedAt),
			UpdatedAt: graphql.TimePtrToGraphqlTimestampPtr(in.UpdatedAt),
			DeletedAt: graphql.TimePtrToGraphqlTimestampPtr(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

// InputFromGraphQL converts the provided graphql-layer representation of an Integration Dependency to the service-layer one.
func (c *converter) InputFromGraphQL(in *graphql.IntegrationDependencyInput) (*model.IntegrationDependencyInput, error) {
	if in == nil {
		return nil, nil
	}

	aspects, err := c.aspectConverter.MultipleInputFromGraphQL(in.Aspects)
	if err != nil {
		return nil, err
	}

	return &model.IntegrationDependencyInput{
		Title:         in.Name,
		Description:   in.Description,
		OrdID:         str.Ptr(in.OrdID),
		OrdPackageID:  in.PartOfPackage,
		Visibility:    str.PtrStrToStr(in.Visibility),
		ReleaseStatus: in.ReleaseStatus,
		Mandatory:     in.Mandatory,
		Aspects:       aspects,
		VersionInput:  c.version.InputFromGraphQL(in.Version),
	}, nil
}
