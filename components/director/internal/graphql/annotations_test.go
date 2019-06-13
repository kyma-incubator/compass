package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestAnnotations_UnmarshalGQL_Success(t *testing.T) {
	as := assert.New(t)

	var tests = []struct {
		input    map[string]interface{}
		expected Annotations
	}{
		//given
		{map[string]interface{}{"annotation": "val1"}, Annotations{"annotation": "val1"}},
		{map[string]interface{}{"annotation": 123}, Annotations{"annotation": 123}},
		{map[string]interface{}{"annotation": []string{"val1", "val2"}}, Annotations{"annotation": []string{"val1", "val2"}}},
	}

	for _, test := range tests {
		//when
		a := Annotations{}
		err := a.UnmarshalGQL(test.input)

		//then
		as.NoError(err)
		as.Equal(test.expected, a)
	}
}

func TestAnnotations_UnmarshalGQL_Error(t *testing.T) {
	t.Run("should return error on invalid map", func(t *testing.T) {
		//given
		a := Annotations{}
		fixAnnotations := map[int]interface{}{
			123: "invalid map",
		}

		//when
		err := a.UnmarshalGQL(fixAnnotations)

		//then
		require.Error(t, err)
		assert.Empty(t, a)
	})

	t.Run("should return error on invalid input type", func(t *testing.T) {
		//given
		a := Annotations{}
		invalidAnnotations := "invalidAnnotations"

		//when
		err := a.UnmarshalGQL(invalidAnnotations)

		//then
		require.Error(t, err)
		assert.Empty(t, a)
	})
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
