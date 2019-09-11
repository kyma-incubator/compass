package graphql

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
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

func TestMarshalJSON(t *testing.T) {
	//given
	fixJSON:= JSON("schema")
	expectedJSON := `"schema"`
	buf := bytes.Buffer{}

	//when
	fixJSON.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedJSON, buf.String())
}
