package str_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/assert"
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
			result := str.Unique(testCase.Input)

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
			result := str.SliceToMap(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestMapToSlice(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    map[string]struct{}
		Expected []string
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
			result := str.MapToSlice(testCase.Input)

			// then
			assert.ElementsMatch(t, testCase.Expected, result)
		})
	}
}

func TestCast(t *testing.T) {
	t.Run("errors when casting non-string data", func(t *testing.T) {
		_, err := str.Cast([]byte{1, 2})

		require.EqualError(t, err, "unable to cast the value to a string type")
	})

	t.Run("returns valid string", func(t *testing.T) {
		s, err := str.Cast("abc")

		require.NoError(t, err)
		require.Equal(t, "abc", s)
	})
}
