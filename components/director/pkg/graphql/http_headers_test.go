package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewHttpHeaders(t *testing.T) {
	for name, tc := range map[string]struct {
		input    map[string][]string
		err      bool
		errmsg   string
		expected HttpHeaders
	}{
		//given
		"correct input": {
			input:    map[string][]string{"header1": []string{"val1", "val2"}},
			err:      false,
			expected: HttpHeaders("{\"header1\":[\"val1\",\"val2\"]}"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			h, err := NewHttpHeaders(tc.input)

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
