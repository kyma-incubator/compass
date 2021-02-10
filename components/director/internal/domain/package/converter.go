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
		PackageLinks:      repo.NewNullableStringFromJSONRawMessage(in.PackageLinks),
		Links:             repo.NewNullableStringFromJSONRawMessage(in.Links),
		LicenseType:       repo.NewNullableString(in.LicenseType),
		Tags:              repo.NewNullableStringFromJSONRawMessage(in.Tags),
		Countries:         repo.NewNullableStringFromJSONRawMessage(in.Countries),
		Labels:            repo.NewNullableStringFromJSONRawMessage(in.Labels),
		PolicyLevel:       in.PolicyLevel,
		CustomPolicyLevel: repo.NewNullableString(in.CustomPolicyLevel),
		PartOfProducts:    repo.NewNullableStringFromJSONRawMessage(in.PartOfProducts),
		LineOfBusiness:    repo.NewNullableStringFromJSONRawMessage(in.LineOfBusiness),
		Industry:          repo.NewNullableStringFromJSONRawMessage(in.Industry),
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
		PackageLinks:      repo.JSONRawMessageFromNullableString(entity.PackageLinks),
		Links:             repo.JSONRawMessageFromNullableString(entity.Links),
		LicenseType:       repo.StringPtrFromNullableString(entity.LicenseType),
		Tags:              repo.JSONRawMessageFromNullableString(entity.Tags),
		Countries:         repo.JSONRawMessageFromNullableString(entity.Countries),
		Labels:            repo.JSONRawMessageFromNullableString(entity.Labels),
		PolicyLevel:       entity.PolicyLevel,
		CustomPolicyLevel: repo.StringPtrFromNullableString(entity.CustomPolicyLevel),
		PartOfProducts:    repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LineOfBusiness:    repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Industry:          repo.JSONRawMessageFromNullableString(entity.Industry),
	}

	return output, nil
}
