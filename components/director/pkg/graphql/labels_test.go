package graphql_test

import (
	"bytes"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
)

func TestLabels_UnmarshalGQL(t *testing.T) {
	for name, tc := range map[string]struct {
		input    interface{}
		err      bool
		errMsg   string
		expected graphql.Labels
	}{
		//given
		"correct input map[string]string": {
			input:    `{"annotation": "val1"}`,
			err:      false,
			expected: graphql.Labels{"annotation": "val1"},
		},
		"correct input map[string]int": {
			input:    `{"annotation": 123}`,
			err:      false,
			expected: graphql.Labels{"annotation": float64(123)},
		},
		"correct input map[string][]string": {
			input:    `{"annotation": ["val1", "val2"]}`,
			err:      false,
			expected: graphql.Labels{"annotation": []interface{}{"val1", "val2"}},
		},
		"error: input is nil": {
			input:  nil,
			err:    true,
			errMsg: "input should not be nil"},
		"error: invalid input type": {
			input:  `{123: "invalid map"}`,
			err:    true,
			errMsg: "Label input is not a valid JSON"},
	} {
		t.Run(name, func(t *testing.T) {
			//when
			a := graphql.Labels{}
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

func TestLabels_MarshalGQL(t *testing.T) {
	as := assert.New(t)

	var tests = []struct {
		input    graphql.Labels
		expected string
	}{
		//given
		{graphql.Labels{"annotation": 123}, `"{\"annotation\":123}"`},
		{graphql.Labels{"annotation": []string{"val1", "val2"}}, `"{\"annotation\":[\"val1\",\"val2\"]}"`},
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
