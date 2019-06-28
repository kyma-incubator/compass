package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotations_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errMsg   string
		expected Annotations
	}{
		//given
		"correct input map[string]string": {
			input:    map[string]interface{}{"annotation": "val1"},
			err:      false,
			expected: Annotations{"annotation": "val1"},
		},
		"correct input map[string]int": {
			input:    map[string]interface{}{"annotation": 123},
			err:      false,
			expected: Annotations{"annotation": 123},
		},
		"correct input map[string][]string": {
			input:    map[string]interface{}{"annotation": []string{"val1", "val2"}},
			err:      false,
			expected: Annotations{"annotation": []string{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errMsg: "input should not be nil"},
		"error: invalid input type": {
			input:  map[int]interface{}{123: "invalid map"},
			err:    true,
			errMsg: "unexpected Annotations type: map[int]interface {}, should be map[string]interface{}"},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			a := Annotations{}
			err := a.UnmarshalGQL(tc.input)

			//then
			if tc.err {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.errMsg)
				assert.Empty(t, a)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, a)
			}
		})
	}
}

func TestAnnotations_MarshalGQL(t *testing.T) {
	as := assert.New(t)

	var tests = []struct {
		input    Annotations
		expected string
	}{
		//given
		{Annotations{"annotation": 123}, `{"annotation":123}`},
		{Annotations{"annotation": []string{"val1", "val2"}}, `{"annotation":["val1","val2"]}`},
	}

	for _, test := range tests {
		//when
		buf := bytes.Buffer{}
		test.input.MarshalGQL(&buf)

		//then
		as.NotNil(buf)
		as.Equal(test.expected, buf.String())
	}
}
