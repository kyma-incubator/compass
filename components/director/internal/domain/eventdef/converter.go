package eventdef

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) *model.Version
	ToEntity(version model.Version) version.Version
}

type converter struct {
	fr FetchRequestConverter
	vc VersionConverter
}

func NewConverter(fr FetchRequestConverter, vc VersionConverter) *converter {
	return &converter{fr: fr, vc: vc}
}

func (c *converter) ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition {
	if in == nil {
		return nil
	}

	return &graphql.EventDefinition{
		ID:            in.ID,
		ApplicationID: in.ApplicationID,
		Name:          in.Name,
		Description:   in.Description,
		Group:         in.Group,
		Spec:          c.eventAPISpecToGraphQL(in.ID, in.Spec),
		Version:       c.vc.ToGraphQL(in.Version),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition {
	var apis []*graphql.EventDefinition
	for _, a := range in {
		if a == nil {
			continue
		}
		apis = append(apis, c.ToGraphQL(a))
	}

	return apis
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) []*model.EventDefinitionInput {
	var arr []*model.EventDefinitionInput
	for _, item := range in {
		api := c.InputFromGraphQL(item)
		arr = append(arr, api)
	}

	return arr
}

func (c *converter) InputFromGraphQL(in *graphql.EventDefinitionInput) *model.EventDefinitionInput {
	if in == nil {
		return nil
	}

	return &model.EventDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		Spec:        c.eventAPISpecInputFromGraphQL(in.Spec),
		Group:       in.Group,
		Version:     c.vc.InputFromGraphQL(in.Version),
	}
}

func (c *converter) eventAPISpecToGraphQL(definitionID string, in *model.EventSpec) *graphql.EventSpec {
	if in == nil {
		return nil
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.EventSpec{
		Data:         data,
		Type:         graphql.EventSpecType(in.Type),
		Format:       graphql.SpecFormat(in.Format),
		DefinitionID: definitionID,
	}
}

func (c *converter) eventAPISpecInputFromGraphQL(in *graphql.EventSpecInput) *model.EventSpecInput {
	if in == nil {
		return nil
	}

	return &model.EventSpecInput{
		Data:          (*string)(in.Data),
		Format:        model.SpecFormat(in.Format),
		EventSpecType: model.EventSpecType(in.Type),
		FetchRequest:  c.fr.InputFromGraphQL(in.FetchRequest),
	}
}

func (c *converter) FromEntity(entity Entity) (model.EventDefinition, error) {
	return model.EventDefinition{
		ID:            entity.ID,
		Tenant:        entity.TenantID,
		ApplicationID: entity.AppID,
		Name:          entity.Name,
		Description:   repo.StringPtrFromNullableString(entity.Description),
		Group:         repo.StringPtrFromNullableString(entity.GroupName),
		Version:       c.vc.FromEntity(entity.Version),
		Spec:          c.apiSpecFromEntity(entity.EntitySpec),
	}, nil
}

func (c *converter) ToEntity(eventModel model.EventDefinition) (Entity, error) {
	return Entity{
		ID:          eventModel.ID,
		TenantID:    eventModel.Tenant,
		AppID:       eventModel.ApplicationID,
		Name:        eventModel.Name,
		Description: repo.NewNullableString(eventModel.Description),
		GroupName:   repo.NewNullableString(eventModel.Group),
		Version:     c.convertVersionToEntity(eventModel.Version),
		EntitySpec:  c.apiSpecToEntity(eventModel.Spec),
	}, nil
}

func (c *converter) convertVersionToEntity(inVer *model.Version) version.Version {
	if inVer == nil {
		return version.Version{}
	}

	return c.vc.ToEntity(*inVer)
}

func (c *converter) apiSpecToEntity(spec *model.EventSpec) EntitySpec {
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

func (c *converter) apiSpecFromEntity(specEnt EntitySpec) *model.EventSpec {
	if !specEnt.SpecType.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecData.Valid {
		return nil
	}

	apiSpec := model.EventSpec{}
	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		apiSpec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		apiSpec.Type = model.EventSpecType(*specType)
	}
	apiSpec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)

	return &apiSpec
}
