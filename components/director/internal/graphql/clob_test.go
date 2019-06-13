package graphql

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLOB_UnmarshalGQL(t *testing.T) {
	//given
	var clob CLOB
	fixClob := []byte("very_big_clob")
	expectedClob := CLOB("very_big_clob")

	//when
	err := clob.UnmarshalGQL(fixClob)

	//then
	require.NoError(t, err)
	assert.Equal(t, clob, expectedClob)
}

func TestCLOB_UnmarshalGQL_Error(t *testing.T) {
	//given
	c := CLOB{}
	invalidClob := 123

	//when
	err := c.UnmarshalGQL(invalidClob)

	//then
	require.Error(t, err)
	assert.Empty(t, c)
}

func TestCLOB_MarshalGQL(t *testing.T) {
	//given
	fixClob := CLOB("very_big_clob")
	expectedClob := []byte("very_big_clob")
	buf := bytes.Buffer{}

	//when
	fixClob.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedClob, buf.Bytes())
}
