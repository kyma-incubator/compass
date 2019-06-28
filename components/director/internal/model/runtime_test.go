package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntime_AddLabel(t *testing.T) {
	// given
	testCases := []struct {
		Name           string
		InitialRuntime model.Runtime
		InputKey       string
		InputValues    []string
		ExpectedLabels map[string][]string
	}{
		{
			Name: "New Label",
			InitialRuntime: model.Runtime{
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
			InitialRuntime: model.Runtime{
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
			InitialRuntime: model.Runtime{
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
			rtm := testCase.InitialRuntime

			// when

			rtm.AddLabel(testCase.InputKey, testCase.InputValues)

			// then

			for key, val := range testCase.ExpectedLabels {
				assert.ElementsMatch(t, val, rtm.Labels[key])
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
		InputValuesToDelete []string
		ExpectedLabels      map[string][]string
		ExpectedErr         error
	}{
		{
			Name:     "Whole Label",
			InputKey: "foo",
			InputRuntime: model.Runtime{
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
			InputRuntime: model.Runtime{
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
			InputRuntime: model.Runtime{
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
			rtm := testCase.InputRuntime

			// when

			err := rtm.DeleteLabel(testCase.InputKey, testCase.InputValuesToDelete)

			// then

			require.Equal(t, testCase.ExpectedErr, err)

			for key, val := range testCase.ExpectedLabels {
				assert.ElementsMatch(t, val, rtm.Labels[key])
			}
		})
	}
}

func TestRuntime_AddAnnotation(t *testing.T) {
	// given
	testCases := []struct {
		Name                string
		InputRuntime        model.Runtime
		InputKey            string
		InputValue          string
		ExpectedAnnotations map[string]interface{}
		ExpectedErr         error
	}{
		{
			Name:       "Success",
			InputKey:   "foo",
			InputValue: "bar",
			InputRuntime: model.Runtime{
				Annotations: map[string]interface{}{
					"test": "val",
				},
			},
			ExpectedErr: nil,
			ExpectedAnnotations: map[string]interface{}{
				"test": "val",
				"foo":  "bar",
			},
		},
		{
			Name:       "Nil map",
			InputKey:   "foo",
			InputValue: "bar",
			InputRuntime: model.Runtime{
				Annotations: nil,
			},
			ExpectedErr: nil,
			ExpectedAnnotations: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			Name:       "Error",
			InputKey:   "foo",
			InputValue: "bar",
			InputRuntime: model.Runtime{
				Annotations: map[string]interface{}{
					"foo": "val",
				},
			},
			ExpectedErr: fmt.Errorf("annotation %s does already exist", "foo"),
			ExpectedAnnotations: map[string]interface{}{
				"foo": "val",
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			rtm := testCase.InputRuntime

			// when

			err := rtm.AddAnnotation(testCase.InputKey, testCase.InputValue)

			// then

			require.Equal(t, testCase.ExpectedErr, err)
			assert.Equal(t, testCase.ExpectedAnnotations, rtm.Annotations)
		})
	}
}

func TestRuntime_DeleteAnnotation(t *testing.T) {
	// given
	testCases := []struct {
		Name                string
		InputRuntime        model.Runtime
		InputKey            string
		ExpectedAnnotations map[string]interface{}
		ExpectedErr         error
	}{
		{
			Name:     "Success",
			InputKey: "foo",
			InputRuntime: model.Runtime{
				Annotations: map[string]interface{}{
					"no":  "delete",
					"foo": "bar",
				},
			},
			ExpectedErr: nil,
			ExpectedAnnotations: map[string]interface{}{
				"no": "delete",
			},
		},
		{
			Name:     "Error",
			InputKey: "foobar",
			InputRuntime: model.Runtime{
				Annotations: map[string]interface{}{
					"no": "delete",
				},
			},
			ExpectedErr: fmt.Errorf("annotation %s doesn't exist", "foobar"),
			ExpectedAnnotations: map[string]interface{}{
				"no": "delete",
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			rtm := testCase.InputRuntime

			// when

			err := rtm.DeleteAnnotation(testCase.InputKey)

			// then

			require.Equal(t, testCase.ExpectedErr, err)
			assert.Equal(t, testCase.ExpectedAnnotations, rtm.Annotations)
		})
	}
}

func TestRuntimeInput_ToRuntime(t *testing.T) {
	// given
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	timestamp := time.Now()
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
				Annotations: map[string]interface{}{
					"key": "value",
				},
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
			},
			Expected: &model.Runtime{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Annotations: map[string]interface{}{
					"key": "value",
				},
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
				Status: &model.RuntimeStatus{
					Condition: model.RuntimeStatusConditionInitial,
					Timestamp: timestamp,
				},
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
			if result != nil && result.Status != nil {
				result.Status.Timestamp = timestamp
			}

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
