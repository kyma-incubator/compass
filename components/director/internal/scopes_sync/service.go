package scopes_sync

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

type SyncService interface {
	SynchronizeClientScopes(context.Context) error
}

//go:generate mockery --name=SystemAuthRepo --output=automock --outpkg=automock --case=underscore
type SystemAuthRepo interface {
	ListGlobalWithConditions(ctx context.Context, conditions repo.Conditions) ([]model.SystemAuth, error)
}

//go:generate mockery --name=OAuthService --output=automock --outpkg=automock --case=underscore
type OAuthService interface {
	ListClients(ctx context.Context) ([]oauth20.Client, error)
	GetClientCredentialScopes(objType model.SystemAuthReferenceObjectType) ([]string, error)
	UpdateClientScopes(ctx context.Context, clientID string, objectType model.SystemAuthReferenceObjectType) error
}

type service struct {
	oAuth20Svc OAuthService
	transact   persistence.Transactioner
	repo       SystemAuthRepo
}

func NewService(oAuth20Svc OAuthService, transact persistence.Transactioner, repo SystemAuthRepo) SyncService {
	return &service{
		oAuth20Svc: oAuth20Svc,
		transact:   transact,
		repo:       repo,
	}
}

func (s *service) SynchronizeClientScopes(ctx context.Context) error {
	clientsFromHydra, err := s.oAuth20Svc.ListClients(ctx)
	if err != nil {
		return errors.Wrap(err, "while listing clients from hydra")
	}
	clientScopes := convertScopesToMap(clientsFromHydra)

	auths, err := s.systemAuthsWithOAuth(ctx)
	if err != nil {
		return err
	}

	for _, auth := range auths {
		clientID := auth.Value.Credential.Oauth.ClientID

		objType, err := auth.GetReferenceObjectType()
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
			continue
		}

		expectedScopes, err := s.oAuth20Svc.GetClientCredentialScopes(objType)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error while getting client credentials scopes for client with ID %s: %v", clientID, err)
			continue
		}

		scopesFromHydra, ok := clientScopes[clientID]
		if !ok {
			log.C(ctx).Errorf("Client with ID %s is not present in Hydra", clientID)
			continue
		}
		if str.Matches(scopesFromHydra, expectedScopes) {
			log.C(ctx).Infof("Scopes for client with ID %s and type %s are in sync", clientID, objType)
			continue
		}

		if err = s.oAuth20Svc.UpdateClientScopes(ctx, clientID, objType); err != nil {
			log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
		}
	}

	log.C(ctx).Info("Finished synchronization of Hydra scopes")
	return nil
}

func convertScopesToMap(clientsFromHydra []oauth20.Client) map[string][]string {
	clientScopes := make(map[string][]string)
	for _, s := range clientsFromHydra {
		clientScopes[s.ClientID] = strings.Split(s.Scopes, " ")
	}
	return clientScopes
}

func (s *service) systemAuthsWithOAuth(ctx context.Context) ([]model.SystemAuth, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening database transaction")
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	conditions := repo.Conditions{
		repo.NewNotNullCondition("(value -> 'Credential' -> 'Oauth')"),
	}
	auths, err := s.repo.ListGlobalWithConditions(ctx, conditions)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while database transaction commit")
	}

	return auths, nil
}
