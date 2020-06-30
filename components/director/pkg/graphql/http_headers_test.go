package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpHeaders_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected HttpHeaders
	}{
		//given
		"correct input": {
			input:    map[string]interface{}{"header1": []interface{}{"val1", "val2"}},
			err:      false,
			expected: HttpHeaders{"header1": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "Invalid data [reason=input should not be nil]",
		},
		"error: invalid input map type": {
			input:  map[string]interface{}{"header": "invalid type"},
			err:    true,
			errmsg: "given value `string` must be a string array",
		},
		"error: invalid input": {
			input:  "invalid headers",
			err:    true,
			errmsg: "unexpected input type: string, should be map[string][]string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			h := HttpHeaders{}
			err := h.UnmarshalGQL(tc.input)

			//then
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

func TestHttpHeaders_MarshalGQL(t *testing.T) {
	//given
	fixHeaders := HttpHeaders{
		"header": []string{"val1", "val2"},
	}
	expectedHeaders := `{"header":["val1","val2"]}`
	buf := bytes.Buffer{}

	//when
	fixHeaders.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedHeaders, buf.String())
}

func Test_NewHttpHeadersSerialized(t *testing.T) {
	t.Run("Success when invoking NewHttpHeadersSerialized", func(t *testing.T) {
		// GIVEN
		expected := HttpHeadersSerialized("{\"header1\":[\"val1\",\"val2\"]}")
		given := map[string][]string{"header1": []string{"val1", "val2"}}

		// WHEN
		marshaled, err := NewHttpHeadersSerialized(given)

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, marshaled)
	})
}

func Test_HttpHeadersSerializedUnmarshal(t *testing.T) {
	t.Run("Success when unmarshaling serialized HttpHeaders", func(t *testing.T) {
		// GIVEN
		expected := map[string][]string{"header1": []string{"val1", "val2"}}
		marshaled := HttpHeadersSerialized("{\"header1\":[\"val1\",\"val2\"]}")

		// WHEN
		headers, err := marshaled.Unmarshal()

		// THEN
		require.NoError(t, err)
		require.Equal(t, expected, headers)
	})
}
