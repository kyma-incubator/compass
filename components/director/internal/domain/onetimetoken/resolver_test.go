package onetimetoken_test

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const suggestTokenHeaderKey = "suggest_token"

var systemID string = ""

func TestResolver_GenerateOneTimeTokenForApp(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	appID := "08d805a5-87f0-4194-adc7-277ec10de2ef"
	ctx := context.TODO()
	tokenModel := &model.OneTimeToken{Token: "Token", ConnectorURL: "connectorURL"}
	expectedToken := graphql.OneTimeTokenForApplication{TokenWithURL: graphql.TokenWithURL{Token: "Token", ConnectorURL: "connectorURL"}}
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, pkgmodel.ApplicationReference).Return(tokenModel, nil)
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQLForApplication", *tokenModel).Return(expectedToken, nil)
		persist, transact := txGen.ThatSucceeds()
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		oneTimeToken, err := r.RequestOneTimeTokenForApplication(ctx, appID, nil)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, oneTimeToken)
		assert.Equal(t, expectedToken, *oneTimeToken)
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})

	t.Run("Success with system auth id", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("RegenerateOneTimeToken", txtest.CtxWithDBMatcher(), systemID).Return(tokenModel, nil)
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQLForApplication", *tokenModel).Return(expectedToken, nil)
		persist, transact := txGen.ThatSucceeds()
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		oneTimeToken, err := r.RequestOneTimeTokenForApplication(ctx, appID, &systemID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, oneTimeToken)
		assert.Equal(t, expectedToken, *oneTimeToken)
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})

	t.Run("Error - transaction commit failed", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, pkgmodel.ApplicationReference).Return(tokenModel, nil)
		persist, transact := txGen.ThatFailsOnCommit()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForApplication(ctx, appID, nil)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})

	t.Run("Error - service return error", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, pkgmodel.ApplicationReference).Return(tokenModel, testErr)
		persist, transact := txGen.ThatDoesntExpectCommit()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForApplication(ctx, appID, nil)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})

	t.Run("Error - begin transaction failed", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		persist, transact := txGen.ThatFailsOnBegin()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForApplication(ctx, appID, nil)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})

	t.Run("Error - converter returns error", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), appID, pkgmodel.ApplicationReference).Return(tokenModel, nil)
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQLForApplication", *tokenModel).Return(graphql.OneTimeTokenForApplication{}, errors.New("some-error"))
		persist, transact := txGen.ThatSucceeds()
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForApplication(ctx, appID, nil)

		// THEN
		require.EqualError(t, err, "while converting one-time token to graphql: some-error")
		mock.AssertExpectationsForObjects(t, persist, transact, svc, conv)
	})
}

func TestResolver_GenerateOneTimeTokenForRuntime(t *testing.T) {
	testErr := errors.New("test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	runtimeID := "08d805a5-87f0-4194-adc7-277ec10de2ef"
	ctx := context.TODO()
	tokenModel := &model.OneTimeToken{Token: "Token", ConnectorURL: "connectorURL"}
	expectedToken := graphql.OneTimeTokenForRuntime{TokenWithURL: graphql.TokenWithURL{Token: "Token", ConnectorURL: "connectorURL"}}
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, pkgmodel.RuntimeReference).Return(tokenModel, nil)
		persist, transact := txGen.ThatSucceeds()
		conv := &automock.TokenConverter{}
		conv.On("ToGraphQLForRuntime", *tokenModel).Return(expectedToken)
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		oneTimeToken, err := r.RequestOneTimeTokenForRuntime(ctx, runtimeID, nil)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, oneTimeToken)
		assert.Equal(t, expectedToken, *oneTimeToken)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - transaction commit failed", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, pkgmodel.RuntimeReference).Return(tokenModel, nil)
		persist, transact := txGen.ThatFailsOnCommit()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForRuntime(ctx, runtimeID, nil)

		// THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - service return error", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		svc.On("GenerateOneTimeToken", txtest.CtxWithDBMatcher(), runtimeID, pkgmodel.RuntimeReference).Return(tokenModel, testErr)
		persist, transact := txGen.ThatDoesntExpectCommit()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForRuntime(ctx, runtimeID, nil)

		// THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Error - begin transaction failed", func(t *testing.T) {
		// GIVEN
		svc := &automock.TokenService{}
		persist, transact := txGen.ThatFailsOnBegin()
		conv := &automock.TokenConverter{}
		r := onetimetoken.NewTokenResolver(transact, svc, conv, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RequestOneTimeTokenForRuntime(ctx, runtimeID, nil)

		// THEN
		require.Error(t, err)
		persist.AssertExpectations(t)
		transact.AssertExpectations(t)
		svc.AssertExpectations(t)
		conv.AssertExpectations(t)
	})
}

func TestResolver_RawEncoded(t *testing.T) {
	ctx := context.TODO()
	tokenGraphql := graphql.OneTimeTokenForApplication{TokenWithURL: graphql.TokenWithURL{Token: "Token", ConnectorURL: "connectorURL", Used: false, Type: graphql.OneTimeTokenTypeApplication}, LegacyConnectorURL: "legacyConnectorURL"}
	expectedRawToken := "{\"token\":\"Token\"," +
		"\"connectorURL\":\"connectorURL\",\"used\":false,\"expiresAt\":null,\"createdAt\":null,\"usedAt\":null,\"type\":\"Application\"}"
	expectedBaseToken := base64.StdEncoding.EncodeToString([]byte(expectedRawToken))
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		r := onetimetoken.NewTokenResolver(nil, nil, nil, "")

		// WHEN
		baseEncodedToken, err := r.RawEncoded(ctx, &tokenGraphql.TokenWithURL)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedBaseToken, baseEncodedToken)
	})

	t.Run("Error - nil token", func(t *testing.T) {
		// GIVEN
		r := onetimetoken.NewTokenResolver(nil, nil, nil, suggestTokenHeaderKey)

		// WHEN
		_, err := r.RawEncoded(ctx, nil)

		// THEN
		require.Error(t, err)
	})
}

func TestResolver_Raw(t *testing.T) {
	ctx := context.TODO()
	tokenGraphql := graphql.OneTimeTokenForApplication{TokenWithURL: graphql.TokenWithURL{Token: "Token", ConnectorURL: "connectorURL", Type: graphql.OneTimeTokenTypeApplication}, LegacyConnectorURL: "legacyConnectorURL"}
	expectedRawToken := "{\"token\":\"Token\"," +
		"\"connectorURL\":\"connectorURL\",\"used\":false,\"expiresAt\":null,\"createdAt\":null,\"usedAt\":null,\"type\":\"Application\"}"

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		r := onetimetoken.NewTokenResolver(nil, nil, nil, suggestTokenHeaderKey)

		// WHEN
		baseEncodedToken, err := r.Raw(ctx, &tokenGraphql.TokenWithURL)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedRawToken, baseEncodedToken)
	})

	t.Run("Error - nil token", func(t *testing.T) {
		// GIVEN
		r := onetimetoken.NewTokenResolver(nil, nil, nil, suggestTokenHeaderKey)

		// WHEN
		_, err := r.Raw(ctx, nil)

		// THEN
		require.Error(t, err)
	})
}
