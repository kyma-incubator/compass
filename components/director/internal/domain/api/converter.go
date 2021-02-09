package api

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

//go:generate mockery -name=SpecConverter -output=automock -outpkg=automock -case=underscore
type SpecConverter interface {
	ToGraphQLAPISpec(in *model.Spec) (*graphql.APISpec, error)
	InputFromGraphQLAPISpec(in *graphql.APISpecInput) (*model.SpecInput, error)
}

type converter struct {
	version       VersionConverter
	specConverter SpecConverter
}

func NewConverter(version VersionConverter, specConverter SpecConverter) *converter {
	return &converter{version: version, specConverter: specConverter}
}

func (c *converter) ToGraphQL(in *model.APIDefinition, spec *model.Spec) (*graphql.APIDefinition, error) {
	if in == nil {
		return nil, nil
	}

	s, err := c.specConverter.ToGraphQLAPISpec(spec)
	if err != nil {
		return nil, err
	}

	return &graphql.APIDefinition{
		ID:          in.ID,
		BundleID:    in.BundleID,
		Name:        in.Name,
		Description: in.Description,
		Spec:        s,
		TargetURL:   in.TargetURL,
		Group:       in.Group,
		Version:     c.version.ToGraphQL(in.Version),
		BaseEntity: &graphql.BaseEntity{
			Ready:     in.Ready,
			CreatedAt: graphql.Timestamp(in.CreatedAt),
			UpdatedAt: graphql.Timestamp(in.UpdatedAt),
			DeletedAt: graphql.Timestamp(in.DeletedAt),
			Error:     in.Error,
		},
	}, nil
}

func (c *converter) MultipleToGraphQL(in []*model.APIDefinition, specs []*model.Spec) ([]*graphql.APIDefinition, error) {
	if len(in) != len(specs) {
		return nil, errors.New("different apis and specs count provided")
	}

	var apis []*graphql.APIDefinition
	for i, a := range in {
		if a == nil {
			continue
		}

		api, err := c.ToGraphQL(a, specs[i])
		if err != nil {
			return nil, err
		}

		apis = append(apis, api)
	}

	return apis, nil
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, []*model.SpecInput, error) {
	var apiDefs []*model.APIDefinitionInput
	var specs []*model.SpecInput

	for _, item := range in {
		api, spec, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, nil, err
		}

		apiDefs = append(apiDefs, api)
		specs = append(specs, spec)
	}

	return apiDefs, specs, nil
}

func (c *converter) InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, *model.SpecInput, error) {
	if in == nil {
		return nil, nil, nil
	}

	spec, err := c.specConverter.InputFromGraphQLAPISpec(in.Spec)
	if err != nil {
		return nil, nil, err
	}

	return &model.APIDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		TargetURL:   in.TargetURL,
		Group:       in.Group,
		Version:     c.version.InputFromGraphQL(in.Version),
	}, spec, nil
}

func (c *converter) FromEntity(entity Entity) model.APIDefinition {

	return model.APIDefinition{
		BundleID:    entity.BndlID,
		Name:        entity.Name,
		TargetURL:   entity.TargetURL,
		Tenant:      entity.TenantID,
		Description: repo.StringPtrFromNullableString(entity.Description),
		Group:       repo.StringPtrFromNullableString(entity.Group),
		Version:     c.version.FromEntity(entity.Version),
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

func (c *converter) ToEntity(apiModel model.APIDefinition) *Entity {

	return &Entity{
		TenantID:    apiModel.Tenant,
		BndlID:      apiModel.BundleID,
		Name:        apiModel.Name,
		Description: repo.NewNullableString(apiModel.Description),
		Group:       repo.NewNullableString(apiModel.Group),
		TargetURL:   apiModel.TargetURL,
		Version:     c.convertVersionToEntity(apiModel.Version),
		BaseEntity: &repo.BaseEntity{
			ID:        apiModel.ID,
			Ready:     apiModel.Ready,
			CreatedAt: apiModel.CreatedAt,
			UpdatedAt: apiModel.UpdatedAt,
			DeletedAt: apiModel.DeletedAt,
			Error:     repo.NewNullableString(apiModel.Error),
		},
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}
