package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLOB_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected CLOB
	}{
		//given
		"correct input": {
			input:    "very_big_clob",
			err:      false,
			expected: CLOB("very_big_clob"),
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "Invalid data [reason=input should not be nil]",
		},
		"error: invalid input": {
			input:  123,
			err:    true,
			errmsg: "unexpected input type: int, should be string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// WHEN
			var c CLOB
			err := c.UnmarshalGQL(tc.input)

			// THEN
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, c)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, c)
			}
		})
	}
}

func TestCLOB_MarshalGQL(t *testing.T) {
	//given
	fixClob := CLOB("very_big_clob")
	expectedClob := `"very_big_clob"`
	buf := bytes.Buffer{}

	// WHEN
	fixClob.MarshalGQL(&buf)

	// THEN
	assert.NotNil(t, buf)
	assert.Equal(t, expectedClob, buf.String())
}
