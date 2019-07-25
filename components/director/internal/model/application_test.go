package model_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

//func TestApplication_SetLabel(t *testing.T) {
//	// given
//	testCases := []struct {
//		Name               string
//		InitialApplication model.Application
//		InputKey           string
//		InputValue         interface{}
//		ExpectedLabels     map[string]interface{}
//	}{
//		{
//			Name: "New Label",
//			InitialApplication: model.Application{
//				Labels: map[string]interface{}{
//					"test": "testVal",
//				},
//			},
//			InputKey:   "foo",
//			InputValue: []string{"bar", "baz", "bar"},
//			ExpectedLabels: map[string]interface{}{
//				"test": "testVal",
//				"foo":  []string{"bar", "baz", "bar"},
//			},
//		},
//		{
//			Name: "Nil map",
//			InitialApplication: model.Application{
//				Labels: nil,
//			},
//			InputKey:   "foo",
//			InputValue: []string{"bar", "baz"},
//			ExpectedLabels: map[string]interface{}{
//				"foo": []string{"bar", "baz"},
//			},
//		},
//	}
//
//	for i, testCase := range testCases {
//		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
//			app := testCase.InitialApplication
//
//			// when
//
//			app.SetLabel(testCase.InputKey, testCase.InputValue)
//
//			// then
//
//			for key, val := range testCase.ExpectedLabels {
//				assert.Equal(t, val, app.Labels[key])
//			}
//		})
//	}
//
//}
//
//func TestApplication_DeleteLabel(t *testing.T) {
//	// given
//	testCases := []struct {
//		Name             string
//		InputApplication model.Application
//		InputKey         string
//		ExpectedLabels   map[string]interface{}
//		ExpectedErr      error
//	}{
//		{
//			Name:     "Whole Label",
//			InputKey: "foo",
//			InputApplication: model.Application{
//				Labels: map[string]interface{}{
//					"no":  "delete",
//					"foo": []string{"bar", "baz"},
//				},
//			},
//			ExpectedErr: nil,
//			ExpectedLabels: map[string]interface{}{
//				"no": "delete",
//			},
//		},
//		{
//			Name:     "Error",
//			InputKey: "foobar",
//			InputApplication: model.Application{
//				Labels: map[string]interface{}{
//					"no": "delete",
//				},
//			},
//			ExpectedErr: fmt.Errorf("label %s doesn't exist", "foobar"),
//			ExpectedLabels: map[string]interface{}{
//				"no": "delete",
//			},
//		},
//	}
//
//	for i, testCase := range testCases {
//		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
//			app := testCase.InputApplication
//
//			// when
//
//			err := app.DeleteLabel(testCase.InputKey)
//
//			// then
//
//			require.Equal(t, testCase.ExpectedErr, err)
//
//			for key, val := range testCase.ExpectedLabels {
//				assert.Equal(t, val, app.Labels[key])
//			}
//		})
//	}
//}

func TestApplicationInput_ToApplication(t *testing.T) {
	// given
	url := "https://foo.bar"
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	testCases := []struct {
		Name     string
		Input    *model.ApplicationInput
		Expected *model.Application
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": map[string]interface{}{
						"test": "foo",
					},
				},
				HealthCheckURL: &url,
			},
			Expected: &model.Application{
				Name:           "Foo",
				ID:             id,
				Tenant:         tenant,
				Description:    &desc,
				HealthCheckURL: &url,
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
			result := testCase.Input.ToApplication(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestApplicationInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.ApplicationInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Application name",
			Input:       model.ApplicationInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Application name is empty",
			Input:       model.ApplicationInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains UpperCase letter",
			Input:       model.ApplicationInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Application name contains special not allowed character",
			Input:       model.ApplicationInput{Name: "not-correct-n@me.yeah"},
			ExpectedErr: testError,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			//WHEN
			err := testCase.Input.Validate()

			//THEN
			if err != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, testCase.ExpectedErr)
			}
		})
	}
}
