package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplication_AddLabel(t *testing.T) {
	// given
	testCases := []struct {
		Name               string
		InitialApplication model.Application
		InputKey           string
		InputValues        []string
		ExpectedLabels     map[string][]string
	}{
		{
			Name: "New Label",
			InitialApplication: model.Application{
				Labels: map[string][]string{
					"test": {"testVal"},
				},
			},
			InputKey:    "foo",
			InputValues: []string{"bar", "baz", "bar"},
			ExpectedLabels: map[string][]string{
				"test": {"testVal"},
				"foo":  {"bar", "baz"},
			},
		},
		{
			Name: "Nil map",
			InitialApplication: model.Application{
				Labels: nil,
			},
			InputKey:    "foo",
			InputValues: []string{"bar", "baz"},
			ExpectedLabels: map[string][]string{
				"foo": {"bar", "baz"},
			},
		},
		{
			Name: "Append Values",
			InitialApplication: model.Application{
				Labels: map[string][]string{
					"foo": {"bar", "baz"},
				},
			},
			InputKey:    "foo",
			InputValues: []string{"zzz", "bar"},
			ExpectedLabels: map[string][]string{
				"foo": {"bar", "baz", "zzz"},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			app := testCase.InitialApplication

			// when

			app.AddLabel(testCase.InputKey, testCase.InputValues)

			// then

			for key, val := range testCase.ExpectedLabels {
				assert.ElementsMatch(t, val, app.Labels[key])
			}
		})
	}

}

func TestApplication_DeleteLabel(t *testing.T) {
	// given
	testCases := []struct {
		Name                string
		InputApplication    model.Application
		InputKey            string
		InputValuesToDelete []string
		ExpectedLabels      map[string][]string
		ExpectedErr         error
	}{
		{
			Name:     "Whole Label",
			InputKey: "foo",
			InputApplication: model.Application{
				Labels: map[string][]string{
					"no":  {"delete"},
					"foo": {"bar", "baz"},
				},
			},
			InputValuesToDelete: []string{},
			ExpectedErr:         nil,
			ExpectedLabels: map[string][]string{
				"no": {"delete"},
			},
		},
		{
			Name:     "Label Values",
			InputKey: "foo",
			InputApplication: model.Application{
				Labels: map[string][]string{
					"no":  {"delete"},
					"foo": {"foo", "bar", "baz"},
				},
			},
			InputValuesToDelete: []string{"bar", "baz"},
			ExpectedErr:         nil,
			ExpectedLabels: map[string][]string{
				"no":  {"delete"},
				"foo": {"foo"},
			},
		},
		{
			Name:     "Error",
			InputKey: "foobar",
			InputApplication: model.Application{
				Labels: map[string][]string{
					"no": {"delete"},
				},
			},
			InputValuesToDelete: []string{"bar", "baz"},
			ExpectedErr:         fmt.Errorf("label %s doesn't exist", "foobar"),
			ExpectedLabels: map[string][]string{
				"no": {"delete"},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			app := testCase.InputApplication

			// when

			err := app.DeleteLabel(testCase.InputKey, testCase.InputValuesToDelete)

			// then

			require.Equal(t, testCase.ExpectedErr, err)

			for key, val := range testCase.ExpectedLabels {
				assert.ElementsMatch(t, val, app.Labels[key])
			}
		})
	}
}

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
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
				HealthCheckURL: &url,
			},
			Expected: &model.Application{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
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
	validationErrorMsg := []string{"a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"}
	testCases := []struct {
		Name           string
		Input          model.ApplicationInput
		ExpectedErrMsg []string
	}{
		{
			Name:           "Correct Application name",
			Input:          model.ApplicationInput{Name: "correct-name.yeah"},
			ExpectedErrMsg: nil,
		},
		{
			Name:           "Returns errors when Application name is empty",
			Input:          model.ApplicationInput{Name: ""},
			ExpectedErrMsg: validationErrorMsg,
		},
		{
			Name:           "Returns errors when Application name contains UpperCase letter",
			Input:          model.ApplicationInput{Name: "Not-correct-name.yeah"},
			ExpectedErrMsg: validationErrorMsg,
		},
		{
			Name:           "Returns errors when Application name contains special not allowed character",
			Input:          model.ApplicationInput{Name: "not-correct-n@me.yeah"},
			ExpectedErrMsg: validationErrorMsg,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			//WHEN
			errorMsg := testCase.Input.Validate()

			//THEN
			if testCase.ExpectedErrMsg != nil {
				require.Len(t, errorMsg, 1)
				assert.Contains(t, errorMsg[0], testCase.ExpectedErrMsg[0])
			} else {
				assert.Nil(t, errorMsg)
			}
		})
	}
}
