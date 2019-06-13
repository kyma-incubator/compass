package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParams_UnmarshalGQL(t *testing.T) {
	//given
	p := QueryParams{}
	fixParams := map[string]interface{}{
		"param1": []string{"val1", "val2"},
	}
	expectedParams := QueryParams{
		"param1": []string{"val1", "val2"},
	}

	//when
	err := p.UnmarshalGQL(fixParams)

	//then
	require.NoError(t, err)
	assert.Equal(t, p, expectedParams)
}

func TestQueryParams_UnmarshalGQL_Error(t *testing.T) {
	t.Run("should return error on invalid map", func(t *testing.T) {
		//given
		params := QueryParams{}
		fixParams := map[string]interface{}{
			"param": "invalid type",
		}
		//when
		err := params.UnmarshalGQL(fixParams)

		//then
		require.Error(t, err)
		assert.Empty(t, params)
	})

	t.Run("should return error on invalid input type", func(t *testing.T) {
		//given
		params := QueryParams{}
		invalidParams := "params"

		//when
		err := params.UnmarshalGQL(invalidParams)

		//then
		require.Error(t, err)
		assert.Empty(t, params)
	})
}

func TestQueryParams_MarshalGQL(t *testing.T) {
	//given
	fixParams := QueryParams{
		"param": []string{"val1", "val2"},
	}
	expectedParams := `{"param":["val1","val2"]}`
	buf := bytes.Buffer{}

	//when
	fixParams.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedParams, buf.String())
}
