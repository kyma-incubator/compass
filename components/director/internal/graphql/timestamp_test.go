package graphql

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

func TestTimestamp_UnmarshalGQL_Error(t *testing.T) {
	t.Run("invalid input", func(t *testing.T) {
		//given
		var timestamp Timestamp
		invalidInput := 123

		//when
		err := timestamp.UnmarshalGQL(invalidInput)

		//then
		require.Error(t, err)
		assert.Empty(t, timestamp)

	})

	t.Run("can't parse time", func(t *testing.T) {
		//given
		var timestamp Timestamp
		invalidTime := "invalid time"

		//when
		err := timestamp.UnmarshalGQL(invalidTime)

		//then
		require.Error(t, err)
		assert.Empty(t, timestamp)

	})
}

func TestTimestamp_MarshalGQL(t *testing.T) {
	//given
	parsedTime, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)
	fixTimestamp := Timestamp(parsedTime)
	expectedTimestamp := `2002-10-02T10:00:00-05:00`
	buf := bytes.Buffer{}

	//when
	fixTimestamp.MarshalGQL(&buf)

	//then
	assert.Equal(t, expectedTimestamp, buf.String())
}
