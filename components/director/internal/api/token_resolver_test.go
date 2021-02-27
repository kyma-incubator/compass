package api

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/internalschema"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/tokens"

	"github.com/stretchr/testify/assert"

	apiMocks "github.com/kyma-incubator/compass/components/director/internal/api/automock"
	onetimeTokenMocks "github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken/automock"
	persistenceMocks "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	timeMocks "github.com/kyma-incubator/compass/components/director/pkg/time/automock"
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

func TestTokenResolver_RegenerateOneTimeToken(t *testing.T) {

	const (
		systemAuthID = "sysAuthID"
		connectorURL = "http://connector.url"
		token        = "tokenValue"
	)

	t.Run("fails when systemAuth cannot be fetched", func(t *testing.T) {
		sysAuthSvc := &onetimeTokenMocks.SystemAuthService{}
		appSvc := &onetimeTokenMocks.ApplicationService{}
		appConverter := &onetimeTokenMocks.ApplicationConverter{}
		extTenantsSvc := &onetimeTokenMocks.ExternalTenantsService{}
		doer := &onetimeTokenMocks.HTTPDoer{}
		tokenGenerator := &onetimeTokenMocks.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetByID", context.Background(), systemAuthID).Return(nil, errors.New("error while fetching"))
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)
		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err, "error while fetching")
	})

	t.Run("fails when new token cannot be generated", func(t *testing.T) {
		sysAuthSvc := &onetimeTokenMocks.SystemAuthService{}
		appSvc := &onetimeTokenMocks.ApplicationService{}
		appConverter := &onetimeTokenMocks.ApplicationConverter{}
		extTenantsSvc := &onetimeTokenMocks.ExternalTenantsService{}
		doer := &onetimeTokenMocks.HTTPDoer{}
		tokenGenerator := &onetimeTokenMocks.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetByID", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		tokenGenerator.On("NewToken").Return("", errors.New("error while token generating"))

		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err, "while generating onetime token error while token generating")
	})

	t.Run("succeeds when systemAuth cannot be updated", func(t *testing.T) {
		sysAuthSvc := &onetimeTokenMocks.SystemAuthService{}
		appSvc := &onetimeTokenMocks.ApplicationService{}
		appConverter := &onetimeTokenMocks.ApplicationConverter{}
		extTenantsSvc := &onetimeTokenMocks.ExternalTenantsService{}
		doer := &onetimeTokenMocks.HTTPDoer{}
		tokenGenerator := &onetimeTokenMocks.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetByID", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(errors.New("error while updating"))
		tokenGenerator.On("NewToken").Return(token, nil)
		timeService.On("Now").Return(time.Now())
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, model.OneTimeToken{}, token)
		assert.Error(t, err)
	})

	t.Run("succeeds when no errors are thrown", func(t *testing.T) {
		sysAuthSvc := &onetimeTokenMocks.SystemAuthService{}
		appSvc := &onetimeTokenMocks.ApplicationService{}
		appConverter := &onetimeTokenMocks.ApplicationConverter{}
		extTenantsSvc := &onetimeTokenMocks.ExternalTenantsService{}
		doer := &onetimeTokenMocks.HTTPDoer{}
		tokenGenerator := &onetimeTokenMocks.TokenGenerator{}
		timeService := &timeMocks.Service{}
		intSystemToAdapterMapping := make(map[string]string)

		sysAuthSvc.On("GetByID", context.Background(), systemAuthID).Return(&model.SystemAuth{Value: &model.Auth{}}, nil)
		sysAuthSvc.On("Update", context.Background(), mock.Anything).Return(nil)
		tokenGenerator.On("NewToken").Return(token, nil)
		now := time.Now()
		timeService.On("Now").Return(now)
		expectedToken := &model.OneTimeToken{
			Token:        token,
			ConnectorURL: connectorURL,
			Type:         tokens.ApplicationToken,
			CreatedAt:    now,
			Used:         false,
			UsedAt:       time.Time{},
		}
		tokenService := onetimetoken.NewTokenService(sysAuthSvc, appSvc, appConverter, extTenantsSvc, doer, tokenGenerator, connectorURL,
			intSystemToAdapterMapping, timeService)

		token, err := tokenService.RegenerateOneTimeToken(context.Background(), systemAuthID, tokens.ApplicationToken)

		assert.Equal(t, expectedToken, &token)
		assert.Nil(t, err)
	})
}
