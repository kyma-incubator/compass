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
	testCases := []struct {
		Name     string
		Input    *model.ApplicationWebhookInput
		Expected *model.ApplicationWebhook
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationWebhookInput{
				Type: model.ApplicationWebhookTypeConfigurationChanged,
				URL:  "foourl",
				Auth: &model.AuthInput{
					AdditionalHeaders: map[string][]string{
						"foo": {"foo", "bar"},
						"bar": {"bar", "foo"},
					},
				},
			},
			Expected: &model.ApplicationWebhook{
				ApplicationID: applicationID,
				ID:            id,
				Type:          model.ApplicationWebhookTypeConfigurationChanged,
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
			Input: &model.ApplicationWebhookInput{},
			Expected: &model.ApplicationWebhook{
				ApplicationID: applicationID,
				ID:            id,
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
			result := testCase.Input.ToWebhook(id, applicationID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
