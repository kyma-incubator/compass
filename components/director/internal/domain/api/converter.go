package api

import (
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

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

type converter struct {
	fr      FetchRequestConverter
	version VersionConverter
}

func NewConverter(fr FetchRequestConverter, version VersionConverter) *converter {
	return &converter{fr: fr, version: version}
}

func (c *converter) ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition {
	if in == nil {
		return nil
	}

	return &graphql.APIDefinition{
		ID:          in.ID,
		BundleID:    in.BundleID,
		Name:        in.Name,
		Description: in.Description,
		Spec:        c.SpecToGraphQL(in.ID, in.Spec),
		TargetURL:   in.TargetURL,
		Group:       in.Group,
		Version:     c.version.ToGraphQL(in.Version),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition {
	var apis []*graphql.APIDefinition
	for _, a := range in {
		if a == nil {
			continue
		}
		apis = append(apis, c.ToGraphQL(a))
	}

	return apis
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) ([]*model.APIDefinitionInput, error) {
	var arr []*model.APIDefinitionInput
	for _, item := range in {
		api, err := c.InputFromGraphQL(item)
		if err != nil {
			return nil, err
		}

		arr = append(arr, api)
	}

	return arr, nil
}

func (c *converter) InputFromGraphQL(in *graphql.APIDefinitionInput) (*model.APIDefinitionInput, error) {
	if in == nil {
		return nil, nil
	}

	spec, err := c.apiSpecInputFromGraphQL(in.Spec)
	if err != nil {
		return nil, err
	}

	return &model.APIDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		TargetURL:   in.TargetURL,
		Group:       in.Group,
		Spec:        spec,
		Version:     c.version.InputFromGraphQL(in.Version),
	}, nil
}

func (c *converter) SpecToGraphQL(definitionID string, in *model.APISpec) *graphql.APISpec {
	if in == nil {
		return nil
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.APISpec{
		Data:         data,
		Type:         graphql.APISpecType(in.Type),
		Format:       graphql.SpecFormat(in.Format),
		DefinitionID: definitionID,
	}
}

func (c *converter) apiSpecInputFromGraphQL(in *graphql.APISpecInput) (*model.APISpecInput, error) {
	if in == nil {
		return nil, nil
	}

	fetchReq, err := c.fr.InputFromGraphQL(in.FetchRequest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting FetchRequest from GraphQL input")
	}

	return &model.APISpecInput{
		Data:         (*string)(in.Data),
		Type:         model.APISpecType(in.Type),
		Format:       model.SpecFormat(in.Format),
		FetchRequest: fetchReq,
	}, nil
}

func (c *converter) FromEntity(entity Entity) model.APIDefinition {

	return model.APIDefinition{
		ID:          entity.ID,
		BundleID:    entity.BundleID,
		Name:        entity.Name,
		TargetURL:   entity.TargetURL,
		Tenant:      entity.TenantID,
		Description: repo.StringPtrFromNullableString(entity.Description),
		Group:       repo.StringPtrFromNullableString(entity.Group),
		Spec:        c.apiSpecFromEntity(entity.EntitySpec),
		Version:     c.version.FromEntity(entity.Version),
	}
}

func (c *converter) ToEntity(apiModel model.APIDefinition) Entity {

	return Entity{
		ID:          apiModel.ID,
		TenantID:    apiModel.Tenant,
		BundleID:    apiModel.BundleID,
		Name:        apiModel.Name,
		Description: repo.NewNullableString(apiModel.Description),
		Group:       repo.NewNullableString(apiModel.Group),
		TargetURL:   apiModel.TargetURL,

		EntitySpec: c.apiSpecToEntity(apiModel.Spec),
		Version:    c.convertVersionToEntity(apiModel.Version),
	}
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.version.ToEntity(*inVer)
}

func (c *converter) apiSpecToEntity(spec *model.APISpec) EntitySpec {
	var apiSpecEnt EntitySpec
	if spec != nil {
		apiSpecEnt = EntitySpec{
			SpecFormat: repo.NewNullableString(str.Ptr(string(spec.Format))),
			SpecType:   repo.NewNullableString(str.Ptr(string(spec.Type))),
			SpecData:   repo.NewNullableString(spec.Data),
		}
	}

	return apiSpecEnt
}

func (c *converter) apiSpecFromEntity(specEnt EntitySpec) *model.APISpec {
	if !specEnt.SpecData.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecType.Valid {
		return nil
	}

	apiSpec := model.APISpec{}
	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		apiSpec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		apiSpec.Type = model.APISpecType(*specType)
	}
	apiSpec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)
	return &apiSpec
}
