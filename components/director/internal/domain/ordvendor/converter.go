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
		OrdID:         in.OrdID,
		TenantID:      in.TenantID,
		ApplicationID: in.ApplicationID,
		Title:         in.Title,
		Type:          in.Type,
		Labels:        repo.NewNullableStringFromJSONRawMessage(in.Labels),
	}

	return output
}

func (c *converter) FromEntity(entity *Entity) (*model.Vendor, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Vendor entity is nil")
	}

	output := &model.Vendor{
		OrdID:         entity.OrdID,
		TenantID:      entity.TenantID,
		ApplicationID: entity.ApplicationID,
		Title:         entity.Title,
		Type:          entity.Type,
		Labels:        repo.JSONRawMessageFromNullableString(entity.Labels),
	}

	return output, nil
}
