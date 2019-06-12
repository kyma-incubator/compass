package scalars

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPageCursor_UnmarshalGQL(t *testing.T) {
	//given
	var pageCursor PageCursor
	fixCursor := "cursor"
	expectedCursor := PageCursor("cursor")

	//when
	err := pageCursor.UnmarshalGQL(fixCursor)

	//then
	require.NoError(t, err)
	assert.Equal(t, pageCursor, expectedCursor)
}

func TestPageCursor_MarshalGQL(t *testing.T) {
	//given
	fixCursor := PageCursor("cursor")

	expectedCursor := `{"pageCursor":"cursor"}`
	buf := bytes.Buffer{}
	//when
	fixCursor.MarshalGQL(&buf)

	//then
	assert.NotNil(t, buf)
	assert.Equal(t, expectedCursor, buf.String())
}
