package scalars

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLabels_UnmarshalGQL(t *testing.T) {
	//given
	l := Labels{}
	fixLabels := map[string]interface{}{
		"label1": []string{"val1", "val2"},
	}
	expectedLabels := Labels{
		"label1": []string{"val1", "val2"},
	}

	//when
	err := l.UnmarshalGQL(fixLabels)

	//then
	require.NoError(t, err)
	assert.Equal(t, l, expectedLabels)
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
