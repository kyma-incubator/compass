package mp_package

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

func (c *converter) ToEntity(in *model.Package) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:                in.ID,
		TenantID:          in.TenantID,
		ApplicationID:     in.ApplicationID,
		OrdID:             in.OrdID,
		Vendor:            repo.NewNullableString(in.Vendor),
		Title:             in.Title,
		ShortDescription:  in.ShortDescription,
		Description:       in.Description,
		Version:           in.Version,
		PackageLinks:      in.PackageLinks,
		Links:             in.Links,
		LicenceType:       repo.NewNullableString(in.LicenceType),
		Tags:              in.Tags,
		Countries:         in.Countries,
		Labels:            in.Labels,
		PolicyLevel:       in.PolicyLevel,
		CustomPolicyLevel: repo.NewNullableString(in.CustomPolicyLevel),
		PartOfProducts:    in.PartOfProducts,
		LineOfBusiness:    in.LineOfBusiness,
		Industry:          in.Industry,
	}

	return output
}

func (c *converter) FromEntity(entity *Entity) (*model.Package, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Package entity is nil")
	}

	output := &model.Package{
		ID:                entity.ID,
		TenantID:          entity.TenantID,
		ApplicationID:     entity.ApplicationID,
		OrdID:             entity.OrdID,
		Vendor:            repo.StringPtrFromNullableString(entity.Vendor),
		Title:             entity.Title,
		ShortDescription:  entity.ShortDescription,
		Description:       entity.Description,
		Version:           entity.Version,
		PackageLinks:      entity.PackageLinks,
		Links:             entity.Links,
		LicenceType:       repo.StringPtrFromNullableString(entity.LicenceType),
		Tags:              entity.Tags,
		Countries:         entity.Countries,
		Labels:            entity.Labels,
		PolicyLevel:       entity.PolicyLevel,
		CustomPolicyLevel: repo.StringPtrFromNullableString(entity.CustomPolicyLevel),
		PartOfProducts:    entity.PartOfProducts,
		LineOfBusiness:    entity.LineOfBusiness,
		Industry:          entity.Industry,
	}

	return output, nil
}
