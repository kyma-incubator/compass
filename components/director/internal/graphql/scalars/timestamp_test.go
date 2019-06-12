package scalars

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_UnmarshalGQL(t *testing.T) {
	//given
	var timestamp Timestamp
	fixTime := "2002-10-02T10:00:00-05:00"
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	expectedTimestamp := Timestamp(parsedTime)

	//when
	err = timestamp.UnmarshalGQL(fixTime)

	//then
	require.NoError(t, err)
	assert.Equal(t, expectedTimestamp, timestamp)
}

func TestTimestamp_MarshalGQL(t *testing.T) {
	//given
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	fixTimestamp := Timestamp(parsedTime)
	expectedTimestamp := `{"timestamp":"2002-10-02T10:00:00-05:00"}`
	buf := bytes.Buffer{}

	//when
	fixTimestamp.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTimestamp, buf.String())
}
