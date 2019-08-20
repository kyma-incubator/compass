package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestApplicationWebhookInput_ToWebhook(t *testing.T) {
	// given
	applicationID := "foo"
	id := "bar"
	tenant := "baz"
	testCases := []struct {
		Name     string
		Input    *model.WebhookInput
		Expected *model.Webhook
	}{
		{
			Name: "All properties given",
			Input: &model.WebhookInput{
				Type: model.WebhookTypeConfigurationChanged,
				URL:  "foourl",
				Auth: &model.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
			},
			Expected: &model.Webhook{
				ApplicationID: applicationID,
				ID:            id,
				Tenant:        tenant,
				Type:          model.WebhookTypeConfigurationChanged,
				URL:           "foourl",
				Auth: &model.Auth{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
			},
		},
		{
			Name:  "Empty",
			Input: &model.WebhookInput{},
			Expected: &model.Webhook{
				ApplicationID: applicationID,
				ID:            id,
				Tenant:        tenant,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToWebhook(id, tenant, applicationID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
