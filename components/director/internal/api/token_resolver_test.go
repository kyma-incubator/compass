package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"

	"github.com/stretchr/testify/assert"

	apiMocks "github.com/kyma-incubator/compass/components/director/internal/api/automock"
	"github.com/pkg/errors"
)

func TestTokenResolver_GenerateCSRToken(t *testing.T) {
	const (
		authID     = "authID"
		tokenValue = "tokenValue"
	)

	t.Run("fails when transaction fails to begin", func(t *testing.T) {
		// GIVEN
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		tokenService := &apiMocks.TokenService{}
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		// WHEN
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authID)
		// THEN
		assert.Error(t, err, "error while transaction begin")
		assert.Nil(t, token)
	})

	t.Run("fails when one time token cannot be regenerated", func(t *testing.T) {
		// GIVEN
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("")).ThatDoesntExpectCommit()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		tokenService := &apiMocks.TokenService{}
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authID, tokens.CSRToken).
			Return(model.OneTimeToken{}, errors.New("error while regenerating"))
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		// WHEN
		token, err := tokenResolver.GenerateCSRToken(context.TODO(), authID)
		// THEN
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("fails when transaction cannot be committed", func(t *testing.T) {
		// GIVEN
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error during commit")).ThatFailsOnCommit()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		tokenService := &apiMocks.TokenService{}
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authID, tokens.CSRToken).
			Return(model.OneTimeToken{}, nil)
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		// WHEN
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authID)
		// THEN
		assert.Error(t, err)
		assert.Nil(t, token)
	})

	t.Run("succeeds when no errors are thrown", func(t *testing.T) {
		// GIVEN
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("")).ThatSucceeds()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		tokenService := &apiMocks.TokenService{}
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authID, tokens.CSRToken).
			Return(model.OneTimeToken{Token: tokenValue}, nil)
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		// WHEN
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authID)
		// THEN
		assert.NoError(t, err)
		assert.Equal(t, tokenValue, token.Token)
	})
}
