package label

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_ValueToStringsSlice(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    interface{}
		Expected []string
		Error    error
	}{
		{
			Name:     "Single value",
			Input:    []interface{}{`abc`},
			Expected: []string{"abc"},
		}, {
			Name:     "Multiple values",
			Input:    []interface{}{`abc`, `cde`},
			Expected: []string{"abc", "cde"},
		}, {
			Name:  "Error when unable to convert to slice of interfaces",
			Input: "some text",
			Error: errors.New("Internal Server Error: cannot convert label value to slice of strings"),
		}, {
			Name:  "Error when unable to convert element to string",
			Input: []interface{}{byte(1)},
			Error: errors.New("Internal Server Error: cannot cast label value as a string"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			sliceOfStrings, err := ValueToStringsSlice(testCase.Input)

			if testCase.Error == nil {
				require.NotNil(t, sliceOfStrings)
				require.Equal(t, testCase.Expected, sliceOfStrings)
			} else {
				assert.EqualError(t, err, testCase.Error.Error())
			}
		})
	}
}
