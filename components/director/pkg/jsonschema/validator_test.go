package jsonschema_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/stretchr/testify/assert"
)

func TestValidator_Validate(t *testing.T) {
	// given
	testCases := []struct {
		Name            string
		InputJsonSchema string
		InputJson       string
		ExpectedBool    bool
		ExpectedErr     bool
		ExpectedErrMsg  string
	}{
		{
			Name: "Success",
			InputJsonSchema: `{
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
			  }
			}`,
			InputJson: `{
			  "foo": "test",
			  "bar": "test",
			  "baz": 123
			}`,
			ExpectedBool:   true,
			ExpectedErr:    false,
			ExpectedErrMsg: "",
		},
		{
			Name: "Json schema and json doesn't match",
			InputJsonSchema: `{
			  "$id": "https://foo.com/bar.schema.json",
			  "title": "foo",
			  "type": "object",
			  "properties": {
				"foo": {
				  "type": "string",
				  "description": "foo"
				},
			  }
			}`,
			InputJson:      `{ "foo": 123}`,
			ExpectedBool:   false,
			ExpectedErr:    true,
			ExpectedErrMsg: "invalid character '}' looking for beginning of object key string",
		},
		{
			Name:            "Empty",
			InputJsonSchema: "",
			InputJson:       "",
			ExpectedBool:    false,
			ExpectedErr:     true,
			ExpectedErrMsg:  "EOF",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {

			svc := jsonschema.NewValidator()

			// when
			ok, err := svc.Validate(testCase.InputJsonSchema, testCase.InputJson)

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
