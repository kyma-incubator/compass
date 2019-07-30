package jsonschema_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
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
