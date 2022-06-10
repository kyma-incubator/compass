package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParams_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected QueryParams
	}{
		//given
		"correct input": {
			input:    map[string][]string{"param1": {"val1", "val2"}},
			err:      false,
			expected: QueryParams{"param1": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "Invalid data [reason=input should not be nil]",
		},
		"error: invalid input": {
			input:  "invalid params",
			err:    true,
			errmsg: "unexpected input type: string, should be map[string][]string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			params := QueryParams{}
			err := params.UnmarshalGQL(tc.input)

			// THEN
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, params)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, params)
			}
		})
	}
}

func TestQueryParams_MarshalGQL(t *testing.T) {
	//given
	fixParams := QueryParams{
		"param": []string{"val1", "val2"},
	}
	expectedParams := `{"param":["val1","val2"]}`
	buf := bytes.Buffer{}

	// WHEN
	fixParams.MarshalGQL(&buf)

	// THEN
	assert.NotNil(t, buf)
	assert.Equal(t, expectedParams, buf.String())
}

func Test_NewQueryParamsSerialized(t *testing.T) {
	t.Run("Success when invoking NewQueryParamsSerialized", func(t *testing.T) {
		// GIVEN
		expected := QueryParamsSerialized("{\"param1\":[\"val1\",\"val2\"]}")
		given := map[string][]string{"param1": []string{"val1", "val2"}}

		// WHEN
		marshaled, err := NewQueryParamsSerialized(given)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, marshaled)
	})
}

func Test_QueryParamsSerializedUnmarshal(t *testing.T) {
	t.Run("Success when unmarshaling serialized QueryParams", func(t *testing.T) {
		// GIVEN
		expected := map[string][]string{"param1": []string{"val1", "val2"}}
		marshaled := QueryParamsSerialized("{\"param1\":[\"val1\",\"val2\"]}")

		// WHEN
		params, err := marshaled.Unmarshal()

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, params)
	})
}
