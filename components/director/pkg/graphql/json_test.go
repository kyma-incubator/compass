package graphql

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnmarshalGQLJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected JSON
	}{
		//given
		"correct input": {
			input:    `{"schema":"schema}"`,
			err:      false,
			expected: JSON(`{"schema":"schema}"`),
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "input should not be nil",
		},
		"error: invalid input": {
			input:  123,
			err:    true,
			errmsg: "unexpected input type: int, should be string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			var j JSON
			err := j.UnmarshalGQL(tc.input)

			//then
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, j)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, j)
			}
		})
	}
}

func TestMarshalGQLJSON(t *testing.T) {
	//given
	fixJSON := JSON("schema")
	expectedJSON := `"schema"`
	buf := bytes.Buffer{}

	//when
	fixJSON.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedJSON, buf.String())
}

func TestJSON_MarshalSchema(t *testing.T) {
	testCases := []struct {
		Name        string
		InputSchema *interface{}
		Expected    *JSON
		ExpectedErr error
	}{
		{
			Name:        "Success",
			InputSchema: interfacePtr(map[string]interface{}{"annotation": []string{"val1", "val2"}}),
			Expected:    jsonPtr(JSON(`{"annotation":["val1","val2"]}`)),
			ExpectedErr: nil,
		},
		{
			Name:        "Success nil input",
			InputSchema: nil,
			Expected:    nil,
			ExpectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//WHEN
			json, err := MarshalSchema(testCase.InputSchema)

			//THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, json)

		})
	}
}

func TestJSON_UnmarshalSchema(t *testing.T) {
	t.Run("Success nil JSON", func(t *testing.T) {
		//GIVEN
		var json *JSON = nil
		var expected *interface{}
		//WHEN
		output, err := json.UnmarshalSchema()
		//THEN
		require.NoError(t, err)
		assert.Equal(t, expected, output)
	})

	t.Run("Success correct schema", func(t *testing.T) {
		//GIVEN
		input := jsonPtr(`{"annotation":["val1","val2"]}`)
		expected := map[string]interface{}{"annotation": []interface{}{"val1", "val2"}}
		//WHEN
		output, err := input.UnmarshalSchema()
		//THEN
		require.NoError(t, err)
		assert.Equal(t, expected, *output)
	})

	t.Run("Error - not correct schema", func(t *testing.T) {
		//GIVEN
		expectedErr := errors.New("invalid character 'b' looking for beginning of value")

		input := jsonPtr(`blblbl"`)
		//WHEN
		output, err := input.UnmarshalSchema()
		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, expectedErr.Error())
		assert.Nil(t, output)
	})
}

func interfacePtr(input interface{}) (*interface{}) {
	var tmp interface{} = input
	return &tmp
}

func jsonPtr(json JSON) *JSON {
	return &json
}
