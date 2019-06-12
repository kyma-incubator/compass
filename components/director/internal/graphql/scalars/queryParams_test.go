package scalars

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
