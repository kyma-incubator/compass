package scopes

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/ory/hydra-client-go/models"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/scopes_sync/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSyncService_UpdateClientScopes(t *testing.T) {
	const clientID = "client-id"
	selectCondition := repo.Conditions{
		repo.NewNotEqualCondition("(value -> 'Credential' -> 'Oauth')", "null"),
	}

	t.Run("fails when oauth service cannot list clients", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients").Return(nil, errors.New("error"))
		scopeSyncSvc := NewService(oauthSvc, nil, &automock.SystemAuthRepo{})
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while listing clients from hydra")
	})

	t.Run("fails when cannot begin transaction", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatFailsOnBegin()
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, &automock.SystemAuthRepo{})
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while opening database transaction")
	})

	t.Run("fails when cannot list systemAuths", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error while transaction begin")).ThatDoesntExpectCommit()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return(nil, errors.New("error while listing systemAuths"))
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "error while listing systemAuths")
	})
	t.Run("fails when cannot commit transaction", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error during transaction commit")).ThatFailsOnCommit()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Error(t, err, "while database transaction commit")
	})

	t.Run("won't update client when object type is invalid", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.EqualError(t, err, "Not all clients were updated successfully")
	})

	t.Run("won't update client when getting client credentials scopes fails", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		oauthSvc.On("GetClientDetails", pkgmodel.ApplicationReference).Return(nil, errors.New("error while getting scopes"))
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.EqualError(t, err, "Not all clients were updated successfully")
	})

	t.Run("won't try to update the client when client is not present in hydra", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{}, nil)
		oauthSvc.On("GetClientDetails", pkgmodel.ApplicationReference).Return(&oauth20.ClientDetails{
			Scopes:     []string{},
			GrantTypes: []string{},
		}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.NoError(t, err)
	})

	t.Run("won't update scopes if not needed", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{
			{
				ClientID: clientID,
				Scope:    "scope",
			},
		}, nil)
		oauthSvc.On("GetClientDetails", pkgmodel.ApplicationReference).Return(&oauth20.ClientDetails{
			Scopes:     []string{"scope"},
			GrantTypes: []string{},
		}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})

	t.Run("won't update scopes when returned client is not for update", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{
			{
				ClientID: clientID,
				Scope:    "scope",
			},
		}, nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})

	t.Run("fails when client update in Hydra fails", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{
			{
				ClientID: clientID,
				Scope:    "first",
			},
		}, nil)
		oauthSvc.On("GetClientDetails", pkgmodel.ApplicationReference).Return(&oauth20.ClientDetails{
			Scopes:     []string{"scope"},
			GrantTypes: []string{},
		}, nil)
		oauthSvc.On("UpdateClient", mock.Anything, "client-id", pkgmodel.ApplicationReference).Return(errors.New("fail"))
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.EqualError(t, err, "Not all clients were updated successfully")
	})

	t.Run("will update scopes successfully", func(t *testing.T) {
		// GIVEN
		oauthSvc := &automock.OAuthService{}
		systemAuthRepo := &automock.SystemAuthRepo{}
		oauthSvc.On("ListClients").Return([]*models.OAuth2Client{
			{
				ClientID: clientID,
				Scope:    "first",
			},
		}, nil)
		oauthSvc.On("GetClientDetails", pkgmodel.ApplicationReference).Return(&oauth20.ClientDetails{
			Scopes:     []string{"scope"},
			GrantTypes: []string{},
		}, nil)
		oauthSvc.On("UpdateClient", mock.Anything, "client-id", pkgmodel.ApplicationReference).Return(nil)
		mockedTx, transactioner := txtest.NewTransactionContextGenerator(errors.New("error")).ThatSucceeds()
		systemAuthRepo.On("ListGlobalWithConditions", mock.Anything, selectCondition).Return([]pkgmodel.SystemAuth{
			{
				AppID: str.Ptr("app-id"),
				Value: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID: clientID,
						},
					},
				},
			},
		}, nil)
		defer mockedTx.AssertExpectations(t)
		defer transactioner.AssertExpectations(t)
		defer oauthSvc.AssertExpectations(t)
		scopeSyncSvc := NewService(oauthSvc, transactioner, systemAuthRepo)
		// WHEN
		err := scopeSyncSvc.SynchronizeClientScopes(context.TODO())
		// THEN
		assert.Nil(t, err)
	})
}
