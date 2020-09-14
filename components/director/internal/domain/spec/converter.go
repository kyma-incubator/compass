package spec

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c converter) FromEntity(specEnt Entity) model.Spec {
	spec := model.Spec{
		ID:                specEnt.ID,
		Tenant:            specEnt.TenantID,
		APIDefinitionID:   repo.StringPtrFromNullableString(specEnt.APIDefinitionID),
		EventDefinitionID: repo.StringPtrFromNullableString(specEnt.EventDefinitionID),
	}
	if !specEnt.SpecData.Valid && !specEnt.SpecFormat.Valid && !specEnt.SpecType.Valid {
		return spec
	}

	specFormat := repo.StringPtrFromNullableString(specEnt.SpecFormat)
	if specFormat != nil {
		spec.Format = model.SpecFormat(*specFormat)
	}

	specType := repo.StringPtrFromNullableString(specEnt.SpecType)
	if specFormat != nil {
		spec.Type = model.SpecType(*specType)
	}
	spec.CustomType = repo.StringPtrFromNullableString(specEnt.CustomType)
	spec.Data = repo.StringPtrFromNullableString(specEnt.SpecData)
	return spec
}

func (c converter) ToEntity(apiModel model.Spec) Entity {
	return Entity{
		ID:                apiModel.ID,
		TenantID:          apiModel.Tenant,
		APIDefinitionID:   repo.NewNullableString(apiModel.APIDefinitionID),
		EventDefinitionID: repo.NewNullableString(apiModel.EventDefinitionID),
		SpecData:          repo.NewNullableString(apiModel.Data),
		SpecFormat:        repo.NewNullableString(str.Ptr(string(apiModel.Format))),
		SpecType:          repo.NewNullableString(str.Ptr(string(apiModel.Type))),
		CustomType:        repo.NewNullableString(apiModel.CustomType),
	}
}
