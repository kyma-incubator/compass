package ordpackage

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
func (c *converter) ToEntity(in *model.Package) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:                           in.ID,
		ApplicationID:                repo.NewNullableString(in.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(in.ApplicationTemplateVersionID),
		OrdID:                        in.OrdID,
		Vendor:                       repo.NewNullableString(in.Vendor),
		Title:                        in.Title,
		ShortDescription:             in.ShortDescription,
		Description:                  in.Description,
		Version:                      in.Version,
		PackageLinks:                 repo.NewNullableStringFromJSONRawMessage(in.PackageLinks),
		Links:                        repo.NewNullableStringFromJSONRawMessage(in.Links),
		LicenseType:                  repo.NewNullableString(in.LicenseType),
		SupportInfo:                  repo.NewNullableString(in.SupportInfo),
		Tags:                         repo.NewNullableStringFromJSONRawMessage(in.Tags),
		Countries:                    repo.NewNullableStringFromJSONRawMessage(in.Countries),
		Labels:                       repo.NewNullableStringFromJSONRawMessage(in.Labels),
		PolicyLevel:                  in.PolicyLevel,
		CustomPolicyLevel:            repo.NewNullableString(in.CustomPolicyLevel),
		PartOfProducts:               repo.NewNullableStringFromJSONRawMessage(in.PartOfProducts),
		LineOfBusiness:               repo.NewNullableStringFromJSONRawMessage(in.LineOfBusiness),
		Industry:                     repo.NewNullableStringFromJSONRawMessage(in.Industry),
		ResourceHash:                 repo.NewNullableString(in.ResourceHash),
		DocumentationLabels:          repo.NewNullableStringFromJSONRawMessage(in.DocumentationLabels),
	}

	return output
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.Package, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Package entity is nil")
	}

	output := &model.Package{
		ID:                           entity.ID,
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		OrdID:                        entity.OrdID,
		Vendor:                       repo.StringPtrFromNullableString(entity.Vendor),
		Title:                        entity.Title,
		ShortDescription:             entity.ShortDescription,
		Description:                  entity.Description,
		Version:                      entity.Version,
		PackageLinks:                 repo.JSONRawMessageFromNullableString(entity.PackageLinks),
		Links:                        repo.JSONRawMessageFromNullableString(entity.Links),
		LicenseType:                  repo.StringPtrFromNullableString(entity.LicenseType),
		SupportInfo:                  repo.StringPtrFromNullableString(entity.SupportInfo),
		Tags:                         repo.JSONRawMessageFromNullableString(entity.Tags),
		Countries:                    repo.JSONRawMessageFromNullableString(entity.Countries),
		Labels:                       repo.JSONRawMessageFromNullableString(entity.Labels),
		PolicyLevel:                  entity.PolicyLevel,
		CustomPolicyLevel:            repo.StringPtrFromNullableString(entity.CustomPolicyLevel),
		PartOfProducts:               repo.JSONRawMessageFromNullableString(entity.PartOfProducts),
		LineOfBusiness:               repo.JSONRawMessageFromNullableString(entity.LineOfBusiness),
		Industry:                     repo.JSONRawMessageFromNullableString(entity.Industry),
		ResourceHash:                 repo.StringPtrFromNullableString(entity.ResourceHash),
		DocumentationLabels:          repo.JSONRawMessageFromNullableString(entity.DocumentationLabels),
	}

	return output, nil
}
