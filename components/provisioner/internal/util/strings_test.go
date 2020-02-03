package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_RemoveNotAllowedCharacters(t *testing.T) {
	testCases := []struct {
		given       string
		expected    string
		description string
	}{
		{
			given:       "provider",
			expected:    "provider",
			description: "string containing only letters",
		},
		{
			given:       "provider-123",
			expected:    "provider",
			description: "string with letters, additional hyphen and digits",
		},
		{
			given:       "!@#provider",
			expected:    "provider",
			description: "string with additional non-alphanumeric characters",
		},
		{
			given:       "",
			expected:    "",
			description: "empty string",
		},
		{
			given:       "!@$%$",
			expected:    "",
			description: "string containing only non-alphanumeric characters",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			output := RemoveNotAllowedCharacters(testCase.given)
			assert.Equal(t, testCase.expected, output)
		})
	}
}

func Test_StartWithLetter(t *testing.T) {
	testCases := []struct {
		given       string
		expected    string
		description string
	}{
		{
			given:       "name",
			expected:    "name",
			description: "string containing only letters",
		},
		{
			given:       "name-123",
			expected:    "name-123",
			description: "string with letters, additional hyphen and digits",
		},
		{
			given:       "!@#name",
			expected:    "c-!@#name",
			description: "string with additional non-alphanumeric characters",
		},
		{
			given:       "",
			expected:    "c",
			description: "empty string",
		},
		{
			given:       "!@$%$",
			expected:    "c-!@$%$",
			description: "string containing only non-alphanumeric characters",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			output := StartWithLetter(testCase.given)
			assert.Equal(t, testCase.expected, output)
		})
	}
}
