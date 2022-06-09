package formationtemplate

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

// NewConverter creates a new instance of gqlConverter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// FromInputGraphQL converts from GraphQL input to internal model
func (c *converter) FromInputGraphQL(in *graphql.FormationTemplateInput) *model.FormationTemplateInput {
	return &model.FormationTemplateInput{
		Name:                          in.Name,
		ApplicationTypes:              in.ApplicationTypes,
		RuntimeTypes:                  in.RuntimeTypes,
		MissingArtifactInfoMessage:    in.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: in.MissingArtifactWarningMessage,
	}
}

func (c *converter) FromModelInputToModel(in *model.FormationTemplateInput, id string) *model.FormationTemplate {
	return &model.FormationTemplate{
		ID:                            id,
		Name:                          in.Name,
		ApplicationTypes:              in.ApplicationTypes,
		RuntimeTypes:                  in.RuntimeTypes,
		MissingArtifactInfoMessage:    in.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: in.MissingArtifactWarningMessage,
	}
}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.FormationTemplate) *graphql.FormationTemplate {
	return &graphql.FormationTemplate{
		ID:                            in.ID,
		Name:                          in.Name,
		ApplicationTypes:              in.ApplicationTypes,
		RuntimeTypes:                  in.RuntimeTypes,
		MissingArtifactInfoMessage:    in.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: in.MissingArtifactWarningMessage,
	}
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.FormationTemplate) []*graphql.FormationTemplate {
	formationTemplates := make([]*graphql.FormationTemplate, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		formationTemplates = append(formationTemplates, c.ToGraphQL(r))
	}

	return formationTemplates
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationTemplate) (*Entity, error) {
	if in == nil {
		return nil, nil
	}
	marshalledApplicationTypes, err := json.Marshal(in.ApplicationTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling application types")
	}
	marshalledRuntimeTypes, err := json.Marshal(in.RuntimeTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while marshalling runtime types")
	}

	return &Entity{
		ID:                            in.ID,
		Name:                          in.Name,
		ApplicationTypes:              string(marshalledApplicationTypes),
		RuntimeTypes:                  string(marshalledRuntimeTypes),
		MissingArtifactInfoMessage:    in.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: in.MissingArtifactWarningMessage,
	}, nil
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(in *Entity) (*model.FormationTemplate, error) {
	if in == nil {
		return nil, nil
	}

	var unmarshalledApplicationTypes []string
	err := json.Unmarshal([]byte(in.ApplicationTypes), &unmarshalledApplicationTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling application types")
	}
	var unmarshalledRuntimeTypes []string
	err = json.Unmarshal([]byte(in.RuntimeTypes), &unmarshalledRuntimeTypes)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling application types")
	}
	return &model.FormationTemplate{
		ID:                            in.ID,
		Name:                          in.Name,
		ApplicationTypes:              unmarshalledApplicationTypes,
		RuntimeTypes:                  unmarshalledRuntimeTypes,
		MissingArtifactInfoMessage:    in.MissingArtifactInfoMessage,
		MissingArtifactWarningMessage: in.MissingArtifactWarningMessage,
	}, nil
}
