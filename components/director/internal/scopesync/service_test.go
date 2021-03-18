package scopesync

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/scopesync/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSyncService_UpdateClientScopes(t *testing.T) {

	selectQuery := "select * from system_auths where (value -> 'Credential' -> 'Oauth') is not null"

	t.Run("fails when oauth service cannot list clients", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return(nil, errors.New("error"))
		scopeSyncSvc := NewService(oauthSvc, nil)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while listing clients from hydra")
	})

	t.Run("fails when cannot begin transaction", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while opening database transaction")
	})

	t.Run("fails when cannot execute DB query", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatDoesntExpectCommit()
		mockedTx.On("Select", mock.Anything, selectQuery).Return(errors.New("error while executing query"))
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while getting Oauth system auths")
	})

	t.Run("fails when cannot convert entities", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("")).ThatDoesntExpectCommit()
		expected := systemauth.Collection{
			systemauth.Entity{
				AppID: repo.NewValidNullableString("app-id"),
				Value: repo.NewValidNullableString("{\"Credential\":\"invalid\"}"),
			},
		}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while converting entities")
	})

	t.Run("fails when cannot commit transaction", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error during transaction commit")).ThatFailsOnCommit()
		expected := systemauth.Collection{}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while database transaction commit")
	})

	t.Run("won't update client when object type is invalid", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		expected := systemauth.Collection{
			systemauth.Entity{
				Value: repo.NewValidNullableString("{\"Credential\":{\"Oauth\":{\"ClientID\":\"client-id\"}}}"),
			},
		}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})

	t.Run("won't update client when getting client credentials scopes fails", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{}, nil)
		oauthSvc.On("GetClientCredentialScopes", model.ApplicationReference).Return(nil, errors.New("error while getting scopes"))
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		expected := systemauth.Collection{
			systemauth.Entity{
				AppID: repo.NewValidNullableString("app-id"),
				Value: repo.NewValidNullableString("{\"Credential\":{\"Oauth\":{\"ClientID\":\"client-id\"}}}"),
			},
		}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})

	t.Run("won't update scopes if not needed", func(t *testing.T) {
		// GIVEN
		expectedScopes := []string{"first", "second"}
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{
			{
				ClientID: "client-id",
				Scopes:   "first second",
			},
		}, nil)
		oauthSvc.On("GetClientCredentialScopes", model.ApplicationReference).Return(expectedScopes, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatSucceeds()
		expected := systemauth.Collection{
			systemauth.Entity{
				ID:    "id",
				AppID: repo.NewValidNullableString("app-id"),
				Value: repo.NewValidNullableString("{\"Credential\":{\"Oauth\":{\"ClientID\":\"client-id\"}}}"),
			},
		}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})

	t.Run("will update scopes if needed", func(t *testing.T) {
		// GIVEN
		expectedScopes := []string{"first", "second"}
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients", context.TODO()).Return([]oauth20.Client{
			{
				ClientID: "client-id",
				Scopes:   "first",
			},
		}, nil)
		oauthSvc.On("GetClientCredentialScopes", model.ApplicationReference).Return(expectedScopes, nil)
		oauthSvc.On("UpdateClientCredentials", context.TODO(), "client-id", model.ApplicationReference).Return(nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatSucceeds()
		expected := systemauth.Collection{
			systemauth.Entity{
				AppID: repo.NewValidNullableString("app-id"),
				Value: repo.NewValidNullableString("{\"Credential\":{\"Oauth\":{\"ClientID\":\"client-id\"}}}"),
			},
		}
		mockedTx.On("Select", mock.Anything, selectQuery).Run(GenerateTestCollection(t, expected)).Return(nil).Once()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner)
		// WHEN
		err := scopeSyncSvc.UpdateClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})
}

func GenerateTestCollection(t *testing.T, generated systemauth.Collection) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		arg, ok := args.Get(0).(*systemauth.Collection)
		require.True(t, ok)
		*arg = generated
	}
}
