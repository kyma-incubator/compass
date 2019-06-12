package scalars

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotations_UnmarshalGQL(t *testing.T) {
	//given
	a := Annotations{}
	fixAnnotations := map[string]interface{}{
		"annotation1": "val1",
	}
	expectedAnnotations := Annotations{
		"annotation1": "val1",
	}

	//when
	err := a.UnmarshalGQL(fixAnnotations)

	//then
	require.NoError(t, err)
	assert.Equal(t, a, expectedAnnotations)
}

func TestAnnotations_MarshalGQL_Int(t *testing.T) {
	//given
	fixAnnotations := Annotations{
		"annotation": 123,
	}

	expectedAnnotations := `{"annotation":123}`
	buf := bytes.Buffer{}
	//when
	fixAnnotations.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedAnnotations, buf.String())
}

func TestAnnotations_MarshalGQL_StringArray(t *testing.T) {
	//given
	fixAnnotations := Annotations{
		"annotation": []string{"val1", "val2"},
	}

	expectedAnnotations2 := `{"annotation":["val1","val2"]}`
	buf := bytes.Buffer{}
	//when
	fixAnnotations.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedAnnotations2, buf.String())
}
