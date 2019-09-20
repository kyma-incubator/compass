package token_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/token"
	"github.com/kyma-incubator/compass/components/director/internal/domain/token/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_GenerateOneTimeTokenForApp(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	appID := "08d805a5-87f0-4194-adc7-277ec10de2ef"
	ctx := context.TODO()
	tokenModel := model.OneTimeToken{Token: "Token", ConnectorURL: "conenctorURL"}
	expectedToken := graphql.OneTimeToken{Token: "Token", ConnectorURL: "conenctorURL"}
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, token.ApplicationToken).Return(tokenModel, nil)
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQL", tokenModel).Return(expectedToken)
		persist, transact := txGen.ThatSucceeds()
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		oneTimeToken, err := r.GenerateOneTimeTokenForApp(ctx, appID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, oneTimeToken)
		assert.Equal(t, expectedToken, *oneTimeToken)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - transaction commit failed", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, token.ApplicationToken).Return(tokenModel, nil)
		persist, transact := txGen.ThatFailsOnCommit()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForApp(ctx, appID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - service return error", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, token.ApplicationToken).Return(tokenModel, testErr)
		persist, transact := txGen.ThatDoesntExpectCommit()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForApp(ctx, appID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - begin transaction failed", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		persist, transact := txGen.ThatFailsOnBegin()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForApp(ctx, appID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})
}

func TestResolver_GenerateOneTimeTokenForRuntime(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	runtimeID := "08d805a5-87f0-4194-adc7-277ec10de2ef"
	ctx := context.TODO()
	tokenModel := model.OneTimeToken{Token: "Token", ConnectorURL: "conenctorURL"}
	expectedToken := graphql.OneTimeToken{Token: "Token", ConnectorURL: "conenctorURL"}
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, token.RuntimeToken).Return(tokenModel, nil)
		persist, transact := txGen.ThatSucceeds()
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQL", tokenModel).Return(expectedToken)
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		oneTimeToken, err := r.GenerateOneTimeTokenForRuntime(ctx, runtimeID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, oneTimeToken)
		assert.Equal(t, expectedToken, *oneTimeToken)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - transaction commit failed", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, token.RuntimeToken).Return(tokenModel, nil)
		persist, transact := txGen.ThatFailsOnCommit()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForRuntime(ctx, runtimeID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - service return error", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, token.RuntimeToken).Return(tokenModel, testErr)
		persist, transact := txGen.ThatDoesntExpectCommit()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForRuntime(ctx, runtimeID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - begin transaction failed", func(t *testing.T) {
		//GIVEN
		svc := &automock.TokenService{}
		persist, transact := txGen.ThatFailsOnBegin()
		conv := &automock.TokenConverter{}
		r := token.NewTokenResolver(transact, svc, conv)

		//WHEN
		_, err := r.GenerateOneTimeTokenForRuntime(ctx, runtimeID)

		//THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})
}
