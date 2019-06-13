package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
			input:    map[string]interface{}{"param1": []string{"val1", "val2"}},
			err:      false,
			expected: QueryParams{"param1": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "input should not be nil",
		},
		"error: invalid input map type": {
			input:  map[string]interface{}{"header": "invalid type"},
			err:    true,
			errmsg: "given value `string` must be a string array",
		},
		"error: invalid input": {
			input:  "invalid params",
			err:    true,
			errmsg: "unexpected input type: string, should be map[string][]string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			params := QueryParams{}
			err := params.UnmarshalGQL(tc.input)

			//then
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

	//when
	fixParams.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedParams, buf.String())
}
