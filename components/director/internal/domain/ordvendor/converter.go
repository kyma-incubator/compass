package ordvendor

import (
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in *model.Vendor) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:            in.ID,
		OrdID:         in.OrdID,
		TenantID:      in.TenantID,
		ApplicationID: in.ApplicationID,
		Title:         in.Title,
		Partners:      repo.NewNullableStringFromJSONRawMessage(in.Partners),
		Labels:        repo.NewNullableStringFromJSONRawMessage(in.Labels),
	}

	return output
}

func (c *converter) FromEntity(entity *Entity) (*model.Vendor, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Vendor entity is nil")
	}

	output := &model.Vendor{
		ID:            entity.ID,
		OrdID:         entity.OrdID,
		TenantID:      entity.TenantID,
		ApplicationID: entity.ApplicationID,
		Title:         entity.Title,
		Partners:      repo.JSONRawMessageFromNullableString(entity.Partners),
		Labels:        repo.JSONRawMessageFromNullableString(entity.Labels),
	}

	return output, nil
}
