package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
)

const (
	appId     = "app-id"
	runtimeId = "runtime-id"
	token     = "abcd-efgh"
)

func TestTokenResolver_GenerateApplicationToken(t *testing.T) {

	t.Run("should generate Application token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", appId, tokens.ApplicationToken).Return(token, nil)

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateApplicationToken(context.Background(), appId)

		// then
		require.NoError(t, err)
		assert.Equal(t, token, generatedToken.Token)
	})

	t.Run("should return error when failed generate Application token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", appId, tokens.ApplicationToken).Return("", apperrors.Internal("error"))

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateApplicationToken(context.Background(), appId)

		// then
		require.Error(t, err)
		assert.Empty(t, generatedToken)
	})

}

func TestTokenResolver_GenerateRuntimeToken(t *testing.T) {

	t.Run("should generate Runtime token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", runtimeId, tokens.RuntimeToken).Return(token, nil)

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateRuntimeToken(context.Background(), runtimeId)

		// then
		require.NoError(t, err)
		assert.Equal(t, token, generatedToken.Token)
	})

	t.Run("should return error when failed generate Runtime token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", runtimeId, tokens.RuntimeToken).Return("", apperrors.Internal("error"))

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateRuntimeToken(context.Background(), runtimeId)

		// then
		require.Error(t, err)
		assert.Empty(t, generatedToken)
	})
}
