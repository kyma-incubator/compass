package strings_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnique(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    []string
		Expected []string
	}{
		{
			Name:     "Unique values",
			Input:    []string{"foo", "bar", "baz"},
			Expected: []string{"foo", "bar", "baz"},
		},
		{
			Name:     "Duplicates",
			Input:    []string{"foo", "bar", "foo", "baz", "baz"},
			Expected: []string{"foo", "bar", "baz"},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result := strings.Unique(testCase.Input)

			// then
			assert.ElementsMatch(t, testCase.Expected, result)
		})
	}
}

func TestSliceToMap(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    []string
		Expected map[string]struct{}
	}{
		{
			Name:  "Unique values",
			Input: []string{"foo", "bar", "baz"},
			Expected: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
		{
			Name:  "Duplicates",
			Input: []string{"foo", "bar", "foo", "baz", "baz"},
			Expected: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result := strings.SliceToMap(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestMapToSlice(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input map[string]struct{}
		Expected    []string
	}{
		{
			Name: "Unique values",
			Input: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
			Expected: []string{"foo", "bar", "baz"},
		},
		{
			Name: "Duplicates",
			Input: map[string]struct{}{
				"foo": {},
				"bar": {},
				"baz": {},
			},
			Expected: []string{"foo", "bar", "baz"},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {

			// when
			result := strings.MapToSlice(testCase.Input)

			// then
			assert.ElementsMatch(t, testCase.Expected, result)
		})
	}
}
