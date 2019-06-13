package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabels_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errmsg   string
		expected Labels
	}{
		//given
		"correct input": {
			input:    map[string]interface{}{"label1": []string{"val1", "val2"}},
			err:      false,
			expected: Labels{"label1": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errmsg: "input should not be nil",
		},
		"error: invalid input map type": {
			input:  map[string]interface{}{"label": "invalid type"},
			err:    true,
			errmsg: "given value `string` must be a string array",
		},
		"error: invalid input": {
			input:  "invalid labels",
			err:    true,
			errmsg: "unexpected input type: string, should be map[string][]string",
		},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			l := Labels{}
			err := l.UnmarshalGQL(tc.input)

			//then
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errmsg)
				assert.Empty(t, l)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, l)
			}
		})
	}
}

func TestLabels_MarshalGQL(t *testing.T) {
	//given
	fixLabels := Labels{
		"label1": []string{"val1", "val2"},
	}
	expectedLabels := `{"label1":["val1","val2"]}`
	buf := bytes.Buffer{}

	//when
	fixLabels.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, buf.String(), expectedLabels)
}
