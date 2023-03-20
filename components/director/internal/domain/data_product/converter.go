package data_product

import (
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the GraphQL, service, and repository layer representations of a Compass DataProduct.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an DataProduct to the service-layer representation model.DataProduct.
func (c *converter) FromEntity(entity *Entity) *model.DataProduct {
	return &model.DataProduct{
		ID:               entity.ID,
		ApplicationID:    entity.ApplicationID,
		OrdID:            repo.StringPtrFromNullableString(entity.OrdID),
		LocalID:          repo.StringPtrFromNullableString(entity.OrdID),
		Title:            repo.StringPtrFromNullableString(entity.Title),
		ShortDescription: repo.StringPtrFromNullableString(entity.ShortDescription),
		Description:      repo.StringPtrFromNullableString(entity.Description),
		Version:          repo.StringPtrFromNullableString(entity.Version),
		ReleaseStatus:    repo.StringPtrFromNullableString(entity.ReleaseStatus),
		Visibility:       &entity.Visibility,
		OrdPackageID:     repo.StringPtrFromNullableString(entity.OrdPackageID),
		Tags:             repo.JSONRawMessageFromNullableString(entity.Tags),
		Industry:         repo.JSONRawMessageFromNullableString(entity.Industry),
		LineOfBusiness:   repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Type:             repo.StringPtrFromNullableString(entity.ProductType),
		DataProductOwner: repo.StringPtrFromNullableString(entity.DataProductOwner),
	}
}

// ToEntity converts the provided service-layer representation of an APIDefinition to the repository-layer one.
func (c *converter) ToEntity(dataProductModel *model.DataProduct) *Entity {
	visibility := dataProductModel.Visibility
	if visibility == nil {
		visibility = str.Ptr("public")
	}

	return &Entity{
		ID:               dataProductModel.ID,
		ApplicationID:    dataProductModel.ApplicationID,
		OrdID:            repo.NewNullableString(dataProductModel.OrdID),
		LocalID:          repo.NewNullableString(dataProductModel.LocalID),
		Title:            repo.NewNullableString(dataProductModel.Title),
		ShortDescription: repo.NewNullableString(dataProductModel.ShortDescription),
		Description:      repo.NewNullableString(dataProductModel.Description),
		Version:          repo.NewNullableString(dataProductModel.Version),
		ReleaseStatus:    repo.NewNullableString(dataProductModel.ReleaseStatus),
		Visibility:       *visibility,
		OrdPackageID:     repo.NewNullableString(dataProductModel.OrdPackageID),
		Tags:             repo.NewNullableStringFromJSONRawMessage(dataProductModel.Tags),
		Industry:         repo.NewNullableStringFromJSONRawMessage(dataProductModel.Industry),
		LineOfBusiness:   repo.NewNullableStringFromJSONRawMessage(dataProductModel.LineOfBusiness),
		ProductType:      repo.NewNullableString(dataProductModel.Type),
		DataProductOwner: repo.NewNullableString(dataProductModel.DataProductOwner),
	}
}
