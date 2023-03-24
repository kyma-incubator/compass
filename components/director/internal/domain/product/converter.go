package product

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
func (c *converter) ToEntity(in *model.Product) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:                  in.ID,
		OrdID:               in.OrdID,
		ApplicationID:       repo.NewNullableString(in.ApplicationID),
		Title:               in.Title,
		ShortDescription:    in.ShortDescription,
		Vendor:              in.Vendor,
		Parent:              repo.NewNullableString(in.Parent),
		CorrelationIDs:      repo.NewNullableStringFromJSONRawMessage(in.CorrelationIDs),
		Tags:                repo.NewNullableStringFromJSONRawMessage(in.Tags),
		Labels:              repo.NewNullableStringFromJSONRawMessage(in.Labels),
		DocumentationLabels: repo.NewNullableStringFromJSONRawMessage(in.DocumentationLabels),
	}

	return output
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.Product, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Product entity is nil")
	}

	output := &model.Product{
		ID:                  entity.ID,
		OrdID:               entity.OrdID,
		ApplicationID:       repo.StringPtrFromNullableString(entity.ApplicationID),
		Title:               entity.Title,
		ShortDescription:    entity.ShortDescription,
		Vendor:              entity.Vendor,
		Parent:              repo.StringPtrFromNullableString(entity.Parent),
		CorrelationIDs:      repo.JSONRawMessageFromNullableString(entity.CorrelationIDs),
		Tags:                repo.JSONRawMessageFromNullableString(entity.Tags),
		Labels:              repo.JSONRawMessageFromNullableString(entity.Labels),
		DocumentationLabels: repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
	}

	return output, nil
}
