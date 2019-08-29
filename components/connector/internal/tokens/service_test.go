package tokens

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/stretchr/testify/assert"
)

const (
	clientId = "client-id"
)

func TestTokenService(t *testing.T) {
	for _, testCase := range []struct {
		description       string
		tokenType         TokenType
		expectedTokenData TokenData
	}{
		{
			description: "should save, resolve and delete ApplicationToken",
			tokenType:   ApplicationToken,
			expectedTokenData: TokenData{
				Type:     ApplicationToken,
				ClientId: clientId,
			},
		},
		{
			description: "should save, resolve and delete RuntimeToken",
			tokenType:   RuntimeToken,
			expectedTokenData: TokenData{
				Type:     RuntimeToken,
				ClientId: clientId,
			},
		},
		{
			description: "should save, resolve and delete CSRToken",
			tokenType:   CSRToken,
			expectedTokenData: TokenData{
				Type:     CSRToken,
				ClientId: clientId,
			},
		},
	} {
		t.Run("test case: "+testCase.description, func(t *testing.T) {
			// given
			tokenService := newTokenService()

			// when
			token, err := tokenService.CreateToken(clientId, testCase.tokenType)

			// then
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// when
			tokenData, err := tokenService.Resolve(token)

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedTokenData, tokenData)

			// when
			tokenService.Delete(token)

			// then
			tokenData, err = tokenService.Resolve(token)
			assert.Error(t, err)
			assert.True(t, err.Code() == apperrors.CodeNotFound)
			assert.Empty(t, tokenData)
		})
	}
}

func TestTokenService_Resolve(t *testing.T) {

	t.Run("should return error when token not found", func(t *testing.T) {
		// given
		tokenService := newTokenService()

		// when
		tokenData, err := tokenService.Resolve("non-existing-token")

		// then
		assert.Error(t, err)
		assert.True(t, err.Code() == apperrors.CodeNotFound)
		assert.Empty(t, tokenData)
	})
}

func newTokenService() Service {
	tokenStore := NewTokenCache(1*time.Minute, 1*time.Minute, 1*time.Minute)
	generator := NewTokenGenerator(10)
	return NewTokenService(tokenStore, generator)
}
