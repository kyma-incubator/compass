package util

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBase64Decoding(t *testing.T) {
	rawString := "hello world"
	encodedOnce := base64.StdEncoding.EncodeToString([]byte(rawString))
	encodedTwice := base64.StdEncoding.EncodeToString([]byte(encodedOnce))

	t.Run("test decode non-base64 string", func(t *testing.T) {
		result := TryDecodeBase64(rawString)
		require.Equal(t, []byte(rawString), result)
	})

	t.Run("test decode single base64 encoded string", func(t *testing.T) {
		result := TryDecodeBase64(encodedOnce)
		require.Equal(t, []byte(rawString), result)
	})

	t.Run("test decode double base64 encoded string", func(t *testing.T) {
		result := TryDecodeBase64(encodedTwice)
		require.Equal(t, []byte(rawString), result)
	})
}
