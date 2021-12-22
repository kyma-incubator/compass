package ordvendor

import (
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Vendor) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:                  in.ID,
		OrdID:               in.OrdID,
		ApplicationID:       in.ApplicationID,
		Title:               in.Title,
		Partners:            repo.NewNullableStringFromJSONRawMessage(in.Partners),
		Labels:              repo.NewNullableStringFromJSONRawMessage(in.Labels),
		DocumentationLabels: repo.NewNullableStringFromJSONRawMessage(in.DocumentationLabels),
	}

	return output
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.Vendor, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Vendor entity is nil")
	}

	output := &model.Vendor{
		ID:                  entity.ID,
		OrdID:               entity.OrdID,
		ApplicationID:       entity.ApplicationID,
		Title:               entity.Title,
		Partners:            repo.JSONRawMessageFromNullableString(entity.Partners),
		Labels:              repo.JSONRawMessageFromNullableString(entity.Labels),
		DocumentationLabels: repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
	}

	return output, nil
}
