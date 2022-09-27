package systemfetcher_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

/*
const (
	testExternal = "external"
	testProvider = "Compass"
)

func newModelBusinessTenantMapping(id, name string) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         "",
		Type:           tenant.Account,
		Provider:       testProvider,
		Status:         tenant.Active,
	}
}
*/

func fixInputValuesForSystem(s systemfetcher.System) model.ApplicationFromTemplateInputValues {
	return model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "name",
			Value:       s.DisplayName,
		},
	}
}

func fixAppInputBySystem(system systemfetcher.System) model.ApplicationRegisterInput {
	initStatusCond := model.ApplicationStatusConditionInitial
	return model.ApplicationRegisterInput{
		Name:            system.DisplayName,
		Description:     &system.ProductDescription,
		BaseURL:         &system.BaseURL,
		ProviderName:    &system.InfrastructureProvider,
		SystemNumber:    &system.SystemNumber,
		StatusCondition: &initStatusCond,
		Labels: map[string]interface{}{
			"managed": "true",
		},
	}
}
