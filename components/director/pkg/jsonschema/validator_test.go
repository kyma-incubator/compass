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
		SecondJsonInput string
		ExpectedBool    bool
		ExpectedErr     bool
		ExpectedErrMsg  string
		isMultipleInput bool
	}{
		{
			Name:            "Success",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       inputJson,
			ExpectedBool:    true,
			isMultipleInput: false,
		},
		{
			Name:            "Success multiple - all jsons valid",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       inputJson,
			SecondJsonInput: inputJson,
			ExpectedBool:    true,
			isMultipleInput: true,
		},
		{
			Name:            "Multiple invalid - one json valid and one json invalid",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       inputJson,
			SecondJsonInput: invalidInputJson,
			ExpectedBool:    false,
			isMultipleInput: true,
		},
		{
			Name:            "Multiple invalid - multiple jsons invalid",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       invalidInputJson,
			SecondJsonInput: invalidInputJson,
			ExpectedBool:    false,
			isMultipleInput: true,
		},
		{
			Name:            "Json schema and json doesn't match",
			InputJsonSchema: validInputJsonSchema,
			InputJson:       invalidInputJson,
			ExpectedBool:    false,
			isMultipleInput: false,
		},
		{
			Name:            "Empty",
			InputJsonSchema: "",
			InputJson:       "",
			ExpectedBool:    false,
			ExpectedErr:     true,
			ExpectedErrMsg:  "EOF",
			isMultipleInput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			svc := jsonschema.NewValidator()

			var ok bool
			var err error

			// when
			if testCase.isMultipleInput == true {
				ok, err = svc.Validate(testCase.InputJsonSchema, testCase.InputJson, testCase.SecondJsonInput)
			} else {
				ok, err = svc.Validate(testCase.InputJsonSchema, testCase.InputJson)
			}

			// then
			if testCase.ExpectedErr == true {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedBool, ok)
		})
	}
}
