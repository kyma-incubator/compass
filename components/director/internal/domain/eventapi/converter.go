package eventapi

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"
	"github.com/pkg/errors"
)

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
	FromEntity(version version.Version) (*model.Version, error)
	ToEntity(version model.Version) (version.Version, error)
}

type converter struct {
	fr FetchRequestConverter
	vc VersionConverter
}

func NewConverter(fr FetchRequestConverter, vc VersionConverter) *converter {
	return &converter{fr: fr, vc: vc}
}

func (c *converter) ToGraphQL(in *model.EventAPIDefinition) *graphql.EventAPIDefinition {
	if in == nil {
		return nil
	}

	return &graphql.EventAPIDefinition{
		ID:            in.ID,
		ApplicationID: in.ApplicationID,
		Name:          in.Name,
		Description:   in.Description,
		Group:         in.Group,
		Spec:          c.eventAPISpecToGraphQL(in.ID, in.Spec),
		Version:       c.vc.ToGraphQL(in.Version),
	}
}

func (c *converter) MultipleToGraphQL(in []*model.EventAPIDefinition) []*graphql.EventAPIDefinition {
	var apis []*graphql.EventAPIDefinition
	for _, a := range in {
		if a == nil {
			continue
		}
		apis = append(apis, c.ToGraphQL(a))
	}

	return apis
}

func (c *converter) MultipleInputFromGraphQL(in []*graphql.EventAPIDefinitionInput) []*model.EventAPIDefinitionInput {
	var arr []*model.EventAPIDefinitionInput
	for _, item := range in {
		api := c.InputFromGraphQL(item)
		arr = append(arr, api)
	}

	return arr
}

func (c *converter) InputFromGraphQL(in *graphql.EventAPIDefinitionInput) *model.EventAPIDefinitionInput {
	if in == nil {
		return nil
	}

	return &model.EventAPIDefinitionInput{
		Name:        in.Name,
		Description: in.Description,
		Spec:        c.eventAPISpecInputFromGraphQL(in.Spec),
		Group:       in.Group,
		Version:     c.vc.InputFromGraphQL(in.Version),
	}
}

func (c *converter) eventAPISpecToGraphQL(definitionID string, in *model.EventAPISpec) *graphql.EventAPISpec {
	if in == nil {
		return nil
	}

	var data *graphql.CLOB
	if in.Data != nil {
		tmp := graphql.CLOB(*in.Data)
		data = &tmp
	}

	return &graphql.EventAPISpec{
		Data:         data,
		Type:         graphql.EventAPISpecType(in.Type),
		Format:       graphql.SpecFormat(in.Format),
		DefinitionID: definitionID,
	}
}

func (c *converter) eventAPISpecInputFromGraphQL(in *graphql.EventAPISpecInput) *model.EventAPISpecInput {
	if in == nil {
		return nil
	}

	return &model.EventAPISpecInput{
		Data:          (*string)(in.Data),
		Format:        model.SpecFormat(in.Format),
		EventSpecType: model.EventAPISpecType(in.EventSpecType),
		FetchRequest:  c.fr.InputFromGraphQL(in.FetchRequest),
	}
}

func (c *converter) FromEntity(eventEntity Entity) (model.EventAPIDefinition, error) {
	ver, err := c.convertVersionFromEntity(eventEntity.Version)
	if err != nil {
		return model.EventAPIDefinition{}, errors.Wrap(err, "while converting version")
	}

	return model.EventAPIDefinition{
		ID:            eventEntity.ID,
		Tenant:        eventEntity.TenantID,
		ApplicationID: eventEntity.AppID,
		Name:          eventEntity.Name,
		Description:   repo.StringPtrFromNullableString(eventEntity.Description),
		Group:         repo.StringPtrFromNullableString(eventEntity.GroupName),
		Version:       ver,
		Spec:          c.apiSpecFromEntity(eventEntity.EntitySpec),
	}, nil
}

func (c *converter) ToEntity(eventModel model.EventAPIDefinition) (Entity, error) {
	ver, err := c.convertVersionToEntity(eventModel.Version)
	if err != nil {
		return Entity{}, err
	}

	return Entity{
		ID:          eventModel.ID,
		TenantID:    eventModel.Tenant,
		AppID:       eventModel.ApplicationID,
		Name:        eventModel.Name,
		Description: repo.NewNullableString(eventModel.Description),
		GroupName:   repo.NewNullableString(eventModel.Group),
		Version:     ver,
		EntitySpec:  c.apiSpecToEntity(eventModel.Spec),
	}, nil
}

func (c *converter) convertVersionFromEntity(inVer *version.Version) (*model.Version, error) {
	if inVer == nil {
		return nil, nil
	}

	tmp, err := c.vc.FromEntity(*inVer)
	if err != nil {
		return nil, errors.Wrap(err, "while converting version")
	}
	return tmp, nil
}

func (c *converter) convertVersionToEntity(inVer *model.Version) (*version.Version, error) {
	if inVer == nil {
		return nil, nil
	}

	tmp, err := c.vc.ToEntity(*inVer)
	if err != nil {
		return nil, errors.Wrap(err, "while converting version")
	}
	return &tmp, nil
}

func (c *converter) apiSpecToEntity(spec *model.EventAPISpec) *EntitySpec {
	if spec == nil {
		return nil
	}

	return &EntitySpec{
		SpecFormat: repo.NewNullableString(strings.Ptr(string(spec.Format))),
		SpecType:   repo.NewNullableString(strings.Ptr(string(spec.Type))),
		SpecData:   repo.NewNullableString(spec.Data),
	}
}

func (c *converter) apiSpecFromEntity(specEnt *EntitySpec) *model.EventAPISpec {
	if specEnt == nil {
		return nil
	}
	apiSpec := &model.EventAPISpec{}
	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		apiSpec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		apiSpec.Type = model.EventAPISpecType(*specType)
	}
	apiSpec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)

	return apiSpec
}
