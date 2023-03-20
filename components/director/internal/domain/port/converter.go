package port

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass Port.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an DataProduct to the service-layer representation model.DataProduct.
func (c *converter) FromEntity(entity *Entity) *model.Port {
	return &model.Port{
		ID:                  entity.ID,
		DataProductID:       entity.DataProductID,
		ApplicationID:       entity.ApplicationID,
		Name:                repo.StringPtrFromNullableString(entity.Name),
		PortType:            repo.StringPtrFromNullableString(entity.PortType),
		Description:         repo.StringPtrFromNullableString(entity.Description),
		ProducerCardinality: repo.StringPtrFromNullableString(entity.ProducerCardinality),
		Disabled:            entity.Disabled.Bool,
	}
}

// ToEntity converts the provided service-layer representation of an Port to the repository-layer one.
func (c *converter) ToEntity(portModel *model.Port) *Entity {
	return &Entity{
		ID:                  portModel.ID,
		DataProductID:       portModel.DataProductID,
		ApplicationID:       portModel.ApplicationID,
		Name:                repo.NewNullableString(portModel.Name),
		PortType:            repo.NewNullableString(portModel.PortType),
		Description:         repo.NewNullableString(portModel.Description),
		ProducerCardinality: repo.NewNullableString(portModel.ProducerCardinality),
		Disabled:            repo.NewNullableBool(&portModel.Disabled),
	}
}
