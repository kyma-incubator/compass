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

func TestAnnotations_MarshalGQL(t *testing.T) {
	//given
	fixAnnotations1 := Annotations{
		"annotation1": 123,
	}
	fixAnnotations2 := Annotations{
		"annotation2": []string{"val1", "val2"},
	}

	expectedAnnotations1 := `{"annotation1":123}`
	expectedAnnotations2 := `{"annotation2":["val1","val2"]}`
	buf1 := bytes.Buffer{}
	buf2 := bytes.Buffer{}
	//when
	fixAnnotations1.MarshalGQL(&buf1)
	fixAnnotations2.MarshalGQL(&buf2)

	//then
	assert.NotNil(t, buf1)
	assert.NotNil(t, buf2)
	assert.Equal(t, expectedAnnotations1, buf1.String())
	assert.Equal(t, expectedAnnotations2, buf2.String())
}
