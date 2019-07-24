package jsonschema_test

import (
	"testing"

	"github.com/xeipuuv/gojsonschema"

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
		ExpectedBool    bool
	}{
		{
			Name:            "Success",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       inputJson,
			ExpectedBool:    true,
		},
		{
			Name:            "Json schema and json doesn't match",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       invalidInputJson,
			ExpectedBool:    false,
		},
		{
			Name:            "Empty",
			InputJsonSchema: "",
			InputJson:       "",
			ExpectedBool:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			var schema *gojsonschema.Schema
			var err error

			if testCase.InputJsonSchema == "" {
				schema = nil
			} else {
				stringLoader := gojsonschema.NewStringLoader(testCase.InputJsonSchema)
				schema, err = gojsonschema.NewSchema(stringLoader)
				require.NoError(t, err)
			}

			svc := jsonschema.NewValidator(schema)

			// when
			ok, err := svc.Validate(testCase.InputJson)
			require.NoError(t, err)

			// then
			assert.Equal(t, testCase.ExpectedBool, ok)
		})
	}
}
