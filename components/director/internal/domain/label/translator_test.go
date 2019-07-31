package label

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ExtractValueFromJSONPath_WithValidInput(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Single word",
			Input:    `$[*] ? (@ == "dd")`,
			Expected: "dd",
		}, {
			Name:     "Separated with space",
			Input:    `$[*] ? (@ == "aa cc")`,
			Expected: "aa cc",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			extractedVal := ExtractValueFromJSONPath(testCase.Input)

			assert.Equal(t, testCase.Expected, *extractedVal)
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
			extractedVal := ExtractValueFromJSONPath(testCase.Input)

			assert.Nil(t, extractedVal)
		})
	}
}
