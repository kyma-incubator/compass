package model_test

import (
	"fmt"
	"testing"

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
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
			},
			Expected: &model.Runtime{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Labels: map[string][]string{
					"test": {"val", "val2"},
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
