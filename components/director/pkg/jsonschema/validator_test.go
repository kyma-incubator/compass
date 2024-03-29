package jsonschema_test

import (
	"testing"

	"github.com/pkg/errors"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"

	"github.com/stretchr/testify/assert"
)

func TestValidator_ValidateString(t *testing.T) {
	// given

	validInputJSONSchema := `{
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

	inputJSON := `{
	  "foo": "test",
	  "bar": "test",
	  "baz": 123
	}`
	invalidInputJSON := `{ "abc": 123 }`

	testCases := []struct {
		Name            string
		InputJSONSchema string
		InputJSON       string
		ExpectedResult  bool
		ExpectedError   error
	}{
		{
			Name:            "Success",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       inputJSON,
			ExpectedResult:  true,
			ExpectedError:   nil,
		},
		{
			Name:            "JSON schema and json doesn't match",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       invalidInputJSON,
			ExpectedResult:  false,
			ExpectedError:   nil,
		},
		{
			Name:            "Empty",
			InputJSONSchema: "",
			InputJSON:       "",
			ExpectedResult:  true,
			ExpectedError:   nil,
		},
		{
			Name:            "Invalid json",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       "{test",
			ExpectedResult:  false,
			ExpectedError:   errors.New("invalid character"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc, err := jsonschema.NewValidatorFromStringSchema(testCase.InputJSONSchema)
			require.NoError(t, err)

			// when
			result, err := svc.ValidateString(testCase.InputJSON)
			// then
			assert.Equal(t, testCase.ExpectedResult, result.Valid)
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				if !testCase.ExpectedResult {
					assert.NotNil(t, result.Error)
				} else {
					assert.Nil(t, result.Error)
				}
			}
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
		ExpectedError   error
	}{
		{
			Name:            "Success",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       inputJSON,
			ExpectedResult:  true,
			ExpectedError:   nil,
		},
		{
			Name:            "JSON schema and json doesn't match",
			InputJSONSchema: validInputJSONSchema,
			InputJSON:       invalidInputJSON,
			ExpectedResult:  false,
			ExpectedError:   nil,
		},
		{
			Name:            "Empty",
			InputJSONSchema: nil,
			InputJSON:       "anything",
			ExpectedResult:  true,
			ExpectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc, err := jsonschema.NewValidatorFromRawSchema(testCase.InputJSONSchema)
			require.NoError(t, err)

			// when
			result, err := svc.ValidateRaw(testCase.InputJSON)
			// then
			assert.Equal(t, testCase.ExpectedResult, result.Valid)
			require.Equal(t, testCase.ExpectedError, err)
			if testCase.ExpectedError != nil {
				return
			}
			if testCase.ExpectedResult {
				require.NoError(t, result.Error)
			} else {
				require.Error(t, result.Error)
			}
		})
	}
}

func TestNewValidatorFromStringSchema_NotValidSchema(t *testing.T) {
	//GIVEN
	stringSchema := `"schema"`
	// WHEN
	_, err := jsonschema.NewValidatorFromStringSchema(stringSchema)
	// THEN
	require.Error(t, err)
	assert.EqualError(t, err, "schema is invalid")
}
