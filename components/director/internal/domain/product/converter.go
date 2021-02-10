package product

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

func (c *converter) ToEntity(in *model.Product) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		OrdID:            in.OrdID,
		TenantID:         in.TenantID,
		ApplicationID:    in.ApplicationID,
		Title:            in.Title,
		ShortDescription: in.ShortDescription,
		Vendor:           in.Vendor,
		Parent:           repo.NewNullableString(in.Parent),
		PPMSObjectID:     repo.NewNullableString(in.PPMSObjectID),
		Labels:           repo.NewNullableStringFromJSONRawMessage(in.Labels),
	}

	return output
}

func (c *converter) FromEntity(entity *Entity) (*model.Product, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Product entity is nil")
	}

	output := &model.Product{
		OrdID:            entity.OrdID,
		TenantID:         entity.TenantID,
		ApplicationID:    entity.ApplicationID,
		Title:            entity.Title,
		ShortDescription: entity.ShortDescription,
		Vendor:           entity.Vendor,
		Parent:           repo.StringPtrFromNullableString(entity.Parent),
		PPMSObjectID:     repo.StringPtrFromNullableString(entity.PPMSObjectID),
		Labels:           repo.JSONRawMessageFromNullableString(entity.Labels),
	}

	return output, nil
}
