package formation

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
}

// NewConverter creates new formation converter
func NewConverter() *converter {
	return &converter{}
}

// FromGraphQL converts graphql.FormationInput to model.Formation
func (c *converter) FromGraphQL(i graphql.FormationInput) model.Formation {
	return model.Formation{Name: i.Name}
}

// ToGraphQL converts model.Formation to graphql.Formation
func (c *converter) ToGraphQL(i *model.Formation) (*graphql.Formation, error) {
	if i == nil {
		return nil, nil
	}

	var formationErr graphql.FormationError
	if i.Error != nil {
		if err := json.Unmarshal(i.Error, &formationErr); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling formation error with ID %q", i.ID)
		}
	}

	return &graphql.Formation{
		ID:                  i.ID,
		Name:                i.Name,
		FormationTemplateID: i.FormationTemplateID,
		State:               string(i.State),
		Error:               formationErr,
	}, nil
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.Formation) ([]*graphql.Formation, error) {
	if in == nil {
		return nil, nil
	}
	formations := make([]*graphql.Formation, 0, len(in))
	for _, f := range in {
		if f == nil {
			continue
		}

		formation, err := c.ToGraphQL(f)
		if err != nil {
			return nil, err
		}
		formations = append(formations, formation)
	}

	return formations, nil
}

func (c *converter) ToEntity(in *model.Formation) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                  in.ID,
		TenantID:            in.TenantID,
		FormationTemplateID: in.FormationTemplateID,
		Name:                in.Name,
		State:               string(in.State),
		Error:               repo.NewNullableStringFromJSONRawMessage(in.Error),
	}
}

func (c *converter) FromEntity(entity *Entity) *model.Formation {
	if entity == nil {
		return nil
	}

	return &model.Formation{
		ID:                  entity.ID,
		TenantID:            entity.TenantID,
		FormationTemplateID: entity.FormationTemplateID,
		Name:                entity.Name,
		State:               model.FormationState(entity.State),
		Error:               repo.JSONRawMessageFromNullableString(entity.Error),
	}
}
