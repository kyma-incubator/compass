package jsonschema_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateString(t *testing.T) {
	// given

	validInputJsonSchema := `{
	  "$id": "https://foo.com/bar.schema.json",
	  "title": "foobarbaz",
	  "type": "object",
	  "properties": {
		"foo": {
		  "type": "string",
		  "description": "foo"
		},
		"bar": {
		  "type": "string",
		  "description": "bar"
		},
		"baz": {
		  "description": "baz",
		  "type": "integer",
		  "minimum": 0
		}
	  },
      "required": ["foo", "bar", "baz"]
	}`

	inputJson := `{
	  "foo": "test",
	  "bar": "test",
	  "baz": 123
	}`
	invalidInputJson := `{ "abc": 123 }`

	testCases := []struct {
		Name            string
		InputJsonSchema string
		InputJson       string
		ExpectedResult  bool
	}{
		{
			Name:            "Success",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       inputJson,
			ExpectedResult:  true,
		},
		{
			Name:            "Json schema and json doesn't match",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       invalidInputJson,
			ExpectedResult:  false,
		},
		{
			Name:            "Empty",
			InputJsonSchema: "",
			InputJson:       "",
			ExpectedResult:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			svc, err := jsonschema.NewValidatorFromStringSchema(testCase.InputJsonSchema)
			require.NoError(t, err)

			// when
			ok, err := svc.ValidateString(testCase.InputJson)
			require.NoError(t, err)

			// then
			assert.Equal(t, testCase.ExpectedResult, ok)
		})
	}
}

func TestValidator_ValidateRaw(t *testing.T) {
	// given
	validInputJSONSchema := map[string]interface{}{
		"$id":   "https://foo.com/bar.schema.json",
		"title": "foobarbaz",
		"type":  "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type":        "string",
				"description": "foo",
			},
			"bar": map[string]interface{}{
				"type":        "string",
				"description": "bar",
			},
			"baz": map[string]interface{}{
				"description": "baz",
				"type":        "integer",
				"minimum":     0,
			},
		},
		"required": []interface{}{"foo", "bar", "baz"},
	}

	inputJSON := map[string]interface{}{
		"foo": "test",
		"bar": "test",
		"baz": 123,
	}
	invalidInputJSON := map[string]interface{}{"abc": 123}

	testCases := []struct {
		Name            string
		InputJSONSchema interface{}
		InputJSON       interface{}
		ExpectedResult  bool
	}{
		{
			Name:            "Success",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       inputJSON,
			ExpectedResult:  true,
		},
		{
			Name:            "Json schema and json doesn't match",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       invalidInputJSON,
			ExpectedResult:  false,
		},
		{
			Name:            "Empty",
			InputJSONSchema: nil,
			InputJSON:       "anything",
			ExpectedResult:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			svc, err := jsonschema.NewValidatorFromRawSchema(testCase.InputJSONSchema)
			require.NoError(t, err)

			// when
			ok, err := svc.ValidateRaw(testCase.InputJSON)
			require.NoError(t, err)

			// then
			assert.Equal(t, testCase.ExpectedResult, ok)
		})
	}
}
