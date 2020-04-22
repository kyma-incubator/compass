package label

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractValueFromJSONPath_WithValidInput(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected []interface{}
	}{
		{
			Name:     "Single word",
			Input:    `$[*] ? (@ == "dd")`,
			Expected: []interface{}{"dd"},
		}, {
			Name:     "Single with space",
			Input:    `$[*] ? (@ == "aa cc")`,
			Expected: []interface{}{"aa cc"},
		}, {
			Name:     "Many words",
			Input:    `$[*] ? (@ == "aacc" || @ == "bbee")`,
			Expected: []interface{}{"aacc", "bbee"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			extractedVal, err := ExtractValueFromJSONPath(testCase.Input)
			require.NotNil(t, extractedVal)
			assert.Equal(t, testCase.Expected, extractedVal)
			assert.NoError(t, err)
		})
	}
}

func Test_ExtractValueFromJSONPath_WithInvalidInput(t *testing.T) {
	testCases := []struct {
		Name  string
		Input string
	}{
		{
			Name:  "Empty input",
			Input: ``,
		}, {
			Name:  "Invalid string",
			Input: `some invalid stirng`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			extractedVal, err := ExtractValueFromJSONPath(testCase.Input)

			assert.Nil(t, extractedVal)
			assert.Error(t, err)
		})
	}
}
