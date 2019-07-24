package graphql

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	for name, tc := range map[string]struct {
		Input      interface{}
		ErrOccurs  bool
		ErrMessage string
		Expected   JSON
	}{
		//given
		"correct input map[string]string": {
			Input:     map[string]interface{}{"annotation": "val1"},
			ErrOccurs: false,
			Expected:  map[string]interface{}{"annotation": "val1"},
		},
		"correct input string": {
			Input:     "annotation",
			ErrOccurs: false,
			Expected:  "annotation",
		},
		"correct input map[string]int": {
			Input:     map[string]interface{}{"annotation": 123},
			ErrOccurs: false,
			Expected:  map[string]interface{}{"annotation": 123},
		},
		"correct input map[string][]string": {
			Input:     map[string]interface{}{"annotation": []string{"val1", "val2"}},
			ErrOccurs: false,
			Expected:  map[string]interface{}{"annotation": []string{"val1", "val2"}},
		},
		"correct input map[int]interface{}": {
			Input:     map[int]interface{}{123: "valid map"},
			ErrOccurs: false,
			Expected:  map[int]interface{}{123: "valid map"},
		},
		"error: input is nil": {
			Input:      nil,
			ErrOccurs:  true,
			ErrMessage: "input should not be nil"},
	}{
		t.Run(name, func(t *testing.T) {
			//when
			val, err := UnmarshalJSON(tc.Input)

			//then
			if tc.ErrOccurs {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.ErrMessage)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.Expected, val)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	as := assert.New(t)

	//given
	var tests = []struct {
		input    JSON
		expected string
	}{
		{"annotation", `"annotation"`},
		{323, `323`},
		{map[string]interface{}{"annotation": 123}, `{"annotation":123}`},
		{map[string]interface{}{"annotation": []string{"val1", "val2"}}, `{"annotation":["val1","val2"]}`},
	}

	for _, test := range tests {
		//when
		m := MarshalJSON(test.input)
		buf := bytes.Buffer{}
		m.MarshalGQL(&buf)

		//then
		as.NotNil(buf)
		expected := fmt.Sprintf("%s\n", test.expected)
		as.Equal(expected, buf.String())
	}
}
