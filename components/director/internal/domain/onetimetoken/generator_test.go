package onetimetoken

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tokenLength = 10
)

func TestTokenGenerator_NewToken(t *testing.T) {
	t.Run("should generate token", func(t *testing.T) {
		// GIVEN
		tokenGenerator := NewTokenGenerator(tokenLength)

		// WHEN
		newToken, apperr := tokenGenerator.NewToken()

		// then
		require.NoError(t, apperr)

		decoded, err := base64.URLEncoding.DecodeString(newToken)
		require.NoError(t, err)

		assert.Equal(t, tokenLength, len(decoded))
	})
}
