package tenantbusinesstype

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new tenant business type converter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.TenantBusinessType) *graphql.TenantBusinessType {
	if in == nil {
		return nil
	}

	return &graphql.TenantBusinessType{
		ID:   in.ID,
		Code: in.Code,
		Name: in.Name,
	}
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.TenantBusinessType) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:   in.ID,
		Code: in.Code,
		Name: in.Name,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.TenantBusinessType {
	if e == nil {
		return nil
	}

	return &model.TenantBusinessType{
		ID:   e.ID,
		Code: e.Code,
		Name: e.Name,
	}
}
