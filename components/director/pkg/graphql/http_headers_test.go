package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPHeaders_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected HTTPHeaders
	}{
		//given
		"correct input": {
			input:    map[string][]string{"header1": {"val1", "val2"}},
			err:      false,
			expected: HTTPHeaders{"header1": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "Invalid data [reason=input should not be nil]",
		},
		"error: invalid input": {
			input:  "invalid headers",
			err:    true,
			errmsg: "unexpected input type: string, should be map[string][]string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			h := HTTPHeaders{}
			err := h.UnmarshalGQL(tc.input)

			// THEN
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, h)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, h)
			}
		})
	}
}

func TestHTTPHeaders_MarshalGQL(t *testing.T) {
	//given
	fixHeaders := HTTPHeaders{
		"header": []string{"val1", "val2"},
	}
	expectedHeaders := `{"header":["val1","val2"]}`
	buf := bytes.Buffer{}

	// WHEN
	fixHeaders.MarshalGQL(&buf)

	// THEN
	assert.NotNil(t, buf)
	assert.Equal(t, expectedHeaders, buf.String())
}

func Test_NewHTTPHeadersSerialized(t *testing.T) {
	t.Run("Success when invoking NewHTTPHeadersSerialized", func(t *testing.T) {
		// GIVEN
		expected := HTTPHeadersSerialized("{\"header1\":[\"val1\",\"val2\"]}")
		given := map[string][]string{"header1": []string{"val1", "val2"}}

		// WHEN
		marshaled, err := NewHTTPHeadersSerialized(given)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, marshaled)
	})
}

func Test_HTTPHeadersSerializedUnmarshal(t *testing.T) {
	t.Run("Success when unmarshaling serialized HTTPHeaders", func(t *testing.T) {
		// GIVEN
		expected := map[string][]string{"header1": []string{"val1", "val2"}}
		marshaled := HTTPHeadersSerialized("{\"header1\":[\"val1\",\"val2\"]}")

		// WHEN
		headers, err := marshaled.Unmarshal()

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, headers)
	})
}
