package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPageCursor_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected PageCursor
	}{
		//given
		"correct input": {
			input:    "cursor",
			err:      false,
			expected: PageCursor("cursor"),
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
			var pageCursor PageCursor
			err := pageCursor.UnmarshalGQL(tc.input)

			// THEN
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, pageCursor)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, pageCursor)
			}
		})
	}
}

func TestPageCursor_MarshalGQL(t *testing.T) {
	//given
	fixCursor := PageCursor("cursor")
	expectedCursor := `"cursor"`
	buf := bytes.Buffer{}

	// WHEN
	fixCursor.MarshalGQL(&buf)

	// THEN
	assert.NotNil(t, buf)
	assert.Equal(t, expectedCursor, buf.String())
}
