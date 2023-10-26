package systemfetcher_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

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

func fixInputValuesForSystem(t *testing.T, s systemfetcher.System) model.ApplicationFromTemplateInputValues {
	systemPayload, err := json.Marshal(s.SystemPayload)
	require.NoError(t, err)
	return model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "name",
			Value:       gjson.GetBytes(systemPayload, "displayName").String(),
		},
	}
}

func fixInputValuesForSystemWhichAppTemplateHasPlaceholders(t *testing.T, s systemfetcher.System) model.ApplicationFromTemplateInputValues {
	systemPayload, err := json.Marshal(s.SystemPayload)
	require.NoError(t, err)
	return model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "name",
			Value:       gjson.GetBytes(systemPayload, "displayName").String(),
		},
	}
}

func fixAppInputBySystem(t *testing.T, system systemfetcher.System) model.ApplicationRegisterInput {
	systemPayload, err := json.Marshal(system.SystemPayload)
	require.NoError(t, err)

	initStatusCond := model.ApplicationStatusConditionInitial
	return model.ApplicationRegisterInput{
		Name:            gjson.GetBytes(systemPayload, "displayName").String(),
		Description:     str.Ptr(gjson.GetBytes(systemPayload, "productDescription").String()),
		BaseURL:         str.Ptr(gjson.GetBytes(systemPayload, "baseUrl").String()),
		ProviderName:    str.Ptr(gjson.GetBytes(systemPayload, "infrastructureProvider").String()),
		SystemNumber:    str.Ptr(gjson.GetBytes(systemPayload, "systemNumber").String()),
		StatusCondition: &initStatusCond,
		Labels: map[string]interface{}{
			"managed": "true",
		},
	}
}

func fixWebhookModel(id string, whMode model.WebhookMode) model.Webhook {
	return model.Webhook{
		ID:         id,
		ObjectID:   id,
		ObjectType: model.ApplicationWebhookReference,
		Type:       model.WebhookTypeConfigurationChanged,
		Mode:       &whMode,
	}
}
