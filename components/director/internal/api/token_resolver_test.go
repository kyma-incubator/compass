package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/internalschema"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"

	"github.com/stretchr/testify/assert"

	apiMocks "github.com/kyma-incubator/compass/components/director/internal/api/automock"
	persistenceMocks "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
)

func TestTokenResolver_GenerateCSRToken(t *testing.T) {

	const (
		authId     = "authID"
		tokenValue = "tokenValue"
	)

	t.Run("fails when transaction fails to begin", func(t *testing.T) {
		transactioner := &persistenceMocks.Transactioner{}
		tokenService := &apiMocks.TokenService{}
		transactioner.On("Begin").Return(nil, errors.New("error while transaction begin"))
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authId)
		assert.Error(t, err, "error while transaction begin")
		assert.Nil(t, token)
	})

	t.Run("fails when one time token cannot be regenerated", func(t *testing.T) {
		transactioner := &persistenceMocks.Transactioner{}
		tokenService := &apiMocks.TokenService{}
		persistenceTx := &persistenceMocks.PersistenceTx{}
		transactioner.On("Begin").Return(persistenceTx, nil)
		transactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authId, tokens.CSRToken).
			Return(model.OneTimeToken{}, errors.New("error while regenerating"))
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authId)
		assert.Error(t, err)
		assert.Equal(t, &internalschema.Token{}, token)

	})

	t.Run("fails when transaction cannot be commited", func(t *testing.T) {
		transactioner := &persistenceMocks.Transactioner{}
		tokenService := &apiMocks.TokenService{}
		persistenceTx := &persistenceMocks.PersistenceTx{}
		transactioner.On("Begin").Return(persistenceTx, nil)
		persistenceTx.On("Commit").Return(errors.New("error"))
		transactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authId, tokens.CSRToken).
			Return(model.OneTimeToken{}, nil)
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authId)
		assert.Error(t, err)
		assert.Nil(t, token)

	})

	t.Run("succeeds when no errors are thrown", func(t *testing.T) {
		transactioner := &persistenceMocks.Transactioner{}
		tokenService := &apiMocks.TokenService{}
		persistenceTx := &persistenceMocks.PersistenceTx{}
		transactioner.On("Begin").Return(persistenceTx, nil)
		persistenceTx.On("Commit").Return(nil)
		transactioner.On("RollbackUnlessCommitted", mock.Anything, mock.Anything).Return()
		tokenService.On("RegenerateOneTimeToken", mock.Anything, authId, tokens.CSRToken).
			Return(model.OneTimeToken{Token: tokenValue}, nil)
		tokenResolver := NewTokenResolver(transactioner, tokenService)
		token, err := tokenResolver.GenerateCSRToken(context.Background(), authId)
		assert.Nil(t, err)
		assert.Equal(t, tokenValue, token.Token)
	})

}
