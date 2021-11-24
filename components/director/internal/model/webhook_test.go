package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestWebhookInput_ToApplicationWebhook(t *testing.T) {
	// given
	applicationID := "foo"
	id := "bar"
	template := `{}`
	webhookMode := model.WebhookModeSync
	webhookURL := "foourl"
	testCases := []struct {
		Name     string
		Input    *model.WebhookInput
		Expected *model.Webhook
	}{
		{
			Name: "All properties given",
			Input: &model.WebhookInput{
				Type: model.WebhookTypeConfigurationChanged,
				URL:  &webhookURL,
				Auth: &model.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:           &webhookMode,
				URLTemplate:    &template,
				InputTemplate:  &template,
				HeaderTemplate: &template,
				OutputTemplate: &template,
			},
			Expected: &model.Webhook{
				ObjectID:   applicationID,
				ObjectType: model.ApplicationWebhookReference,
				ID:         id,
				Type:       model.WebhookTypeConfigurationChanged,
				URL:        &webhookURL,
				Auth: &model.Auth{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:           &webhookMode,
				URLTemplate:    &template,
				InputTemplate:  &template,
				HeaderTemplate: &template,
				OutputTemplate: &template,
			},
		},
		{
			Name:  "Empty",
			Input: &model.WebhookInput{},
			Expected: &model.Webhook{
				ObjectID:   applicationID,
				ObjectType: model.ApplicationWebhookReference,
				ID:         id,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := testCase.Input.ToWebhook(id, applicationID, model.ApplicationWebhookReference)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestWebhookInput_ToApplicationTemplateWebhook(t *testing.T) {
	// given
	applicationTemplateID := "foo"
	id := "bar"
	template := `{}`
	webhookMode := model.WebhookModeSync
	webhookURL := "foourl"
	testCases := []struct {
		Name     string
		Input    *model.WebhookInput
		Expected *model.Webhook
	}{
		{
			Name: "All properties given",
			Input: &model.WebhookInput{
				Type: model.WebhookTypeConfigurationChanged,
				URL:  &webhookURL,
				Auth: &model.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:           &webhookMode,
				URLTemplate:    &template,
				InputTemplate:  &template,
				HeaderTemplate: &template,
				OutputTemplate: &template,
			},
			Expected: &model.Webhook{
				ObjectID:   applicationTemplateID,
				ObjectType: model.ApplicationTemplateWebhookReference,
				ID:         id,
				Type:       model.WebhookTypeConfigurationChanged,
				URL:        &webhookURL,
				Auth: &model.Auth{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
				Mode:           &webhookMode,
				URLTemplate:    &template,
				InputTemplate:  &template,
				HeaderTemplate: &template,
				OutputTemplate: &template,
			},
		},
		{
			Name:  "Empty",
			Input: &model.WebhookInput{},
			Expected: &model.Webhook{
				ObjectID:   applicationTemplateID,
				ObjectType: model.ApplicationTemplateWebhookReference,
				ID:         id,
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := testCase.Input.ToWebhook(id, applicationTemplateID, model.ApplicationTemplateWebhookReference)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
