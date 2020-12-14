package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	appAuthId     = "app-id"
	runtimeAuthId = "runtime-id"
	token         = "abcd-efgh"
)

func TestTokenResolver_GenerateApplicationToken(t *testing.T) {

	t.Run("should generate Application token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", mock.Anything, appAuthId, tokens.ApplicationToken).Return(token, nil)

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateApplicationToken(context.Background(), appAuthId)

		// then
		require.NoError(t, err)
		assert.Equal(t, token, generatedToken.Token)
		mock.AssertExpectationsForObjects(t, tokenSvc)
	})

	t.Run("should return error when failed generate Application token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", mock.Anything, appAuthId, tokens.ApplicationToken).Return("", apperrors.Internal("error"))

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateApplicationToken(context.Background(), appAuthId)

		// then
		require.Error(t, err)
		assert.Empty(t, generatedToken)
		mock.AssertExpectationsForObjects(t, tokenSvc)
	})

}

func TestTokenResolver_GenerateRuntimeToken(t *testing.T) {

	t.Run("should generate Runtime token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", mock.Anything, runtimeAuthId, tokens.RuntimeToken).Return(token, nil)

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateRuntimeToken(context.Background(), runtimeAuthId)

		// then
		require.NoError(t, err)
		assert.Equal(t, token, generatedToken.Token)
		mock.AssertExpectationsForObjects(t, tokenSvc)
	})

	t.Run("should return error when failed generate Runtime token", func(t *testing.T) {
		// given
		tokenSvc := &mocks.Service{}
		tokenSvc.On("CreateToken", mock.Anything, runtimeAuthId, tokens.RuntimeToken).Return("", apperrors.Internal("error"))

		tokenResolver := NewTokenResolver(tokenSvc)

		// when
		generatedToken, err := tokenResolver.GenerateRuntimeToken(context.Background(), runtimeAuthId)

		// then
		require.Error(t, err)
		assert.Empty(t, generatedToken)
		mock.AssertExpectationsForObjects(t, tokenSvc)
	})
}
