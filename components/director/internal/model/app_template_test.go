package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestApplicationTemplateInput_ToApplicationTemplate(t *testing.T) {
	// GIVEN
	testID := "test"
	testName := "name"
	testDescription := str.Ptr("desc")
	testAppInputJSON := `{"Name": "app"}`
	testPlaceholders := []model.ApplicationTemplatePlaceholder{
		{Name: "a", Description: str.Ptr("c"), JSONPath: str.Ptr("e")},
		{Name: "b", Description: str.Ptr("d"), JSONPath: str.Ptr("f")},
	}
	testAccessLevel := model.GlobalApplicationTemplateAccessLevel

	webhookMode := model.WebhookModeSync
	webhookURL := "foourl"
	testWebhooks := []*model.WebhookInput{
		{
			Type: model.WebhookTypeConfigurationChanged,
			URL:  &webhookURL,
			Mode: &webhookMode,
		},
	}
	expectedTestWebhooks := []model.Webhook{
		{
			ObjectID:   testID,
			ObjectType: model.ApplicationTemplateWebhookReference,
			Type:       testWebhooks[0].Type,
			URL:        testWebhooks[0].URL,
			Mode:       testWebhooks[0].Mode,
		},
	}

	testCases := []struct {
		Name     string
		Input    *model.ApplicationTemplateInput
		Expected model.ApplicationTemplate
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationTemplateInput{
				Name:                 testName,
				Description:          testDescription,
				ApplicationNamespace: str.Ptr("ns"),
				ApplicationInputJSON: testAppInputJSON,
				Placeholders:         testPlaceholders,
				AccessLevel:          testAccessLevel,
				Labels:               map[string]interface{}{"test": "test"},
				Webhooks:             testWebhooks,
			},
			Expected: model.ApplicationTemplate{
				ID:                   testID,
				Name:                 testName,
				Description:          testDescription,
				ApplicationNamespace: str.Ptr("ns"),
				ApplicationInputJSON: testAppInputJSON,
				Placeholders:         testPlaceholders,
				AccessLevel:          testAccessLevel,
				Webhooks:             expectedTestWebhooks,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: model.ApplicationTemplate{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToApplicationTemplate(testID)

			for i, webhook := range result.Webhooks {
				testCase.Expected.Webhooks[i].ID = webhook.ID
			}

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestApplicationFromTemplateInputValues_FindPlaceholderValue(t *testing.T) {
	// GIVEN
	var values model.ApplicationFromTemplateInputValues = []*model.ApplicationTemplateValueInput{
		{Placeholder: "a", Value: "foo"},
		{Placeholder: "b", Value: "bar"},
	}

	expectedSuccessRes := "foo"
	expectedErr := fmt.Errorf("value for placeholder name '%s' not found", "baz")

	testCases := []struct {
		Name           string
		Input          string
		ExpectedResult *string
		ExpectedError  *error
	}{
		{
			Name:           "Success",
			Input:          "a",
			ExpectedResult: &expectedSuccessRes,
		},
		{
			Name:          "Error",
			Input:         "baz",
			ExpectedError: &expectedErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result, err := values.FindPlaceholderValue(testCase.Input)

			// THEN
			if testCase.ExpectedResult != nil {
				assert.Equal(t, testCase.ExpectedResult, &result)
			}

			if testCase.ExpectedError != nil {
				assert.Equal(t, *testCase.ExpectedError, err)
			}
		})
	}
}
