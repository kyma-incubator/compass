package eventapi

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=FetchRequestConverter -output=automock -outpkg=automock -case=underscore
type FetchRequestConverter interface {
	ToGraphQL(in *model.FetchRequest) *graphql.FetchRequest
	InputFromGraphQL(in *graphql.FetchRequestInput) *model.FetchRequestInput
}

//go:generate mockery -name=VersionConverter -output=automock -outpkg=automock -case=underscore
type VersionConverter interface {
	ToGraphQL(in *model.Version) *graphql.Version
	InputFromGraphQL(in *graphql.VersionInput) *model.VersionInput
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
		Spec:          c.eventApiSpecToGraphQL(in.Spec),
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
		Spec:        c.eventApiSpecInputFromGraphQL(in.Spec),
		Group:       in.Group,
		Version:     c.vc.InputFromGraphQL(in.Version),
	}
}

func (c *converter) eventApiSpecToGraphQL(in *model.EventAPISpec) *graphql.EventAPISpec {
	if in == nil {
		return nil
	}

	var data graphql.CLOB
	if in.Data != nil {
		data = graphql.CLOB(*in.Data)
	}

	var format *graphql.SpecFormat
	if in.Format != nil {
		f := graphql.SpecFormat(*in.Format)
		format = &f
	}

	return &graphql.EventAPISpec{
		Data:         &data,
		Type:         graphql.EventAPISpecType(in.Type),
		Format:       format,
		FetchRequest: c.fr.ToGraphQL(in.FetchRequest),
	}
}

func (c *converter) eventApiSpecInputFromGraphQL(in *graphql.EventAPISpecInput) *model.EventAPISpecInput {
	if in == nil {
		return nil
	}

	var data []byte
	if in.Data != nil {
		data = []byte(*in.Data)
	}

	var format model.SpecFormat
	if in.Format != "" {
		format = model.SpecFormat(in.Format)
	}

	return &model.EventAPISpecInput{
		Data:          &data,
		Format:        &format,
		EventSpecType: model.EventAPISpecType(in.EventSpecType),
		FetchRequest:  c.fr.InputFromGraphQL(in.FetchRequest),
	}
}
