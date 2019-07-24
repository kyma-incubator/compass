package model_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_SetLabel(t *testing.T) {
	// given
	testCases := []struct {
		Name           string
		InitialRuntime model.Runtime
		InputKey       string
		InputValue     interface{}
		ExpectedLabels map[string][]string
	}{
		{
			Name: "New Label",
			InitialRuntime: model.Runtime{
				Labels: map[string]interface{}{
					"test": []string{"testVal"},
				},
			},
			InputKey:   "foo",
			InputValue: []string{"bar", "baz"},
			ExpectedLabels: map[string][]string{
				"test": {"testVal"},
				"foo":  {"bar", "baz"},
			},
		},
		{
			Name: "Nil map",
			InitialRuntime: model.Runtime{
				Labels: nil,
			},
			InputKey:   "foo",
			InputValue: []string{"bar", "baz"},
			ExpectedLabels: map[string][]string{
				"foo": {"bar", "baz"},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			rtm := testCase.InitialRuntime

			// when

			rtm.SetLabel(testCase.InputKey, testCase.InputValue)

			// then

			for key, val := range testCase.ExpectedLabels {
				assert.Equal(t, val, rtm.Labels[key])
			}
		})
	}

}

func TestRuntime_DeleteLabel(t *testing.T) {
	// given
	testCases := []struct {
		Name                string
		InputRuntime        model.Runtime
		InputKey            string
		ExpectedLabels      map[string]interface{}
		ExpectedErr         error
	}{
		{
			Name:     "Whole Label",
			InputKey: "foo",
			InputRuntime: model.Runtime{
				Labels: map[string]interface{}{
					"no":  "delete",
					"foo": []string{"bar", "baz"},
				},
			},
			ExpectedErr:         nil,
			ExpectedLabels: map[string]interface{}{
				"no": "delete",
			},
		},
		{
			Name:     "Error",
			InputKey: "foobar",
			InputRuntime: model.Runtime{
				Labels: map[string]interface{}{
					"no": "delete",
				},
			},
			ExpectedErr:         fmt.Errorf("label %s doesn't exist", "foobar"),
			ExpectedLabels: map[string]interface{}{
				"no": "delete",
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			rtm := testCase.InputRuntime

			// when

			err := rtm.DeleteLabel(testCase.InputKey)

			// then

			require.Equal(t, testCase.ExpectedErr, err)

			for key, val := range testCase.ExpectedLabels {
				assert.Equal(t, val, rtm.Labels[key])
			}
		})
	}
}

func TestRuntimeInput_ToRuntime(t *testing.T) {
	// given
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	testCases := []struct {
		Name     string
		Input    *model.RuntimeInput
		Expected *model.Runtime
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": []string{"val", "val2"},
				},
			},
			Expected: &model.Runtime{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Labels: map[string]interface{}{
					"test": []string{"val", "val2"},
				},
				Status:    &model.RuntimeStatus{},
				AgentAuth: &model.Auth{},
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
			result := testCase.Input.ToRuntime(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestRuntimeInput_ValidateInput(t *testing.T) {
	//GIVEN
	testError := errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character")
	testCases := []struct {
		Name        string
		Input       model.RuntimeInput
		ExpectedErr error
	}{
		{
			Name:        "Correct Runtime name",
			Input:       model.RuntimeInput{Name: "correct-name.yeah"},
			ExpectedErr: nil,
		},
		{
			Name:        "Returns errors when Runtime name is empty",
			Input:       model.RuntimeInput{Name: ""},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Runtime name contains UpperCase letter",
			Input:       model.RuntimeInput{Name: "Not-correct-name.yeah"},
			ExpectedErr: testError,
		},
		{
			Name:        "Returns errors when Runtime name contains special not allowed character",
			Input:       model.RuntimeInput{Name: "not-correct-n@me.yeah"},
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
