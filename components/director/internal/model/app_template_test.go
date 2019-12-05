package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestApplicationTemplateInput_ToApplicationTemplate(t *testing.T) {
	// given
	testID := "test"
	testName := "name"
	testDescription := str.Ptr("desc")
	testAppInputJSON := `{"Name": "app"}`
	testPlaceholders := []model.ApplicationTemplatePlaceholder{
		{Name: "a", Description: str.Ptr("c")},
		{Name: "b", Description: str.Ptr("d")},
	}
	testAccessLevel := model.GlobalApplicationTemplateAccessLevel

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
				ApplicationInputJSON: testAppInputJSON,
				Placeholders:         testPlaceholders,
				AccessLevel:          testAccessLevel,
			},
			Expected: model.ApplicationTemplate{
				ID:                   testID,
				Name:                 testName,
				Description:          testDescription,
				ApplicationInputJSON: testAppInputJSON,
				Placeholders:         testPlaceholders,
				AccessLevel:          testAccessLevel,
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

			// when
			result := testCase.Input.ToApplicationTemplate(testID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
