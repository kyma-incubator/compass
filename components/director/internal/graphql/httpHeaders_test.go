package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpHeaders_UnmarshalGQL(t *testing.T) {
	//given
	h := HttpHeaders{}
	fixHeaders := map[string]interface{}{
		"header1": []string{"val1", "val2"},
	}
	expectedHeaders := HttpHeaders{
		"header1": []string{"val1", "val2"},
	}

	//when
	err := h.UnmarshalGQL(fixHeaders)

	//then
	require.NoError(t, err)
	assert.Equal(t, h, expectedHeaders)
}

func TestHttpHeaders_UnmarshalGQL_Error(t *testing.T) {
	t.Run("should return error on invalid map", func(t *testing.T) {
		//given
		h := HttpHeaders{}
		fixHeaders := map[string]interface{}{
			"header": "invalid type",
		}
		//when
		err := h.UnmarshalGQL(fixHeaders)

		//then
		require.Error(t, err)
		assert.Empty(t, h)
	})

	t.Run("should return error on invalid input type", func(t *testing.T) {
		//given
		h := HttpHeaders{}
		invalidHeaders := "headers"

		//when
		err := h.UnmarshalGQL(invalidHeaders)

		//then
		require.Error(t, err)
		assert.Empty(t, h)
	})
}

func TestHttpHeaders_MarshalGQL(t *testing.T) {
	//given
	fixHeaders := HttpHeaders{
		"header": []string{"val1", "val2"},
	}
	expectedHeaders := `{"header":["val1","val2"]}`
	buf := bytes.Buffer{}

	//when
	fixHeaders.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedHeaders, buf.String())
}
