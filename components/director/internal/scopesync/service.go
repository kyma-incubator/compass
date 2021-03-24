package scopesync

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

type SyncService interface {
	UpdateClientScopes(context.Context) error
}

//go:generate mockery -name=OAuthService -output=automock -outpkg=automock -case=underscore
type OAuthService interface {
	ListClients(ctx context.Context) ([]oauth20.Client, error)
	GetClientCredentialScopes(objType model.SystemAuthReferenceObjectType) ([]string, error)
	UpdateClientScopes(ctx context.Context, clientID string, objectType model.SystemAuthReferenceObjectType) error
}

type service struct {
	oAuth20Svc OAuthService
	transact   persistence.Transactioner
}

func NewService(oAuth20Svc OAuthService, transact persistence.Transactioner) *service {
	return &service{
		oAuth20Svc: oAuth20Svc,
		transact:   transact,
	}
}

func (s *service) UpdateClientScopes(ctx context.Context) error {
	clientsFromHydra, err := s.oAuth20Svc.ListClients(ctx)
	if err != nil {
		return errors.Wrap(err, "while listing clients from hydra")
	}
	clientScopes := convertScopesToMap(clientsFromHydra)

	auths, err := systemAuthsWithOAuth(ctx, s.transact)
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
			log.C(ctx).Errorf("Client with ID %s not presents in Hydra", clientID)
		}
		if !str.Matches(scopesFromHydra, expectedScopes) {
			err = s.oAuth20Svc.UpdateClientScopes(ctx, clientID, objType)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
			}
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

func systemAuthsWithOAuth(ctx context.Context, transact persistence.Transactioner) ([]model.SystemAuth, error) {
	tx, err := transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening database transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)

	auths, err := listOauthAuths(tx)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while database transaction commit")
	}

	return auths, nil
}

func listOauthAuths(persist persistence.PersistenceOp) ([]model.SystemAuth, error) {
	var dest systemauth.Collection

	query := "select * from system_auths where (value -> 'Credential' -> 'Oauth') is not null"
	log.D().Debugf("Executing DB query: %s", query)
	err := persist.Select(&dest, query)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Oauth system auths")
	}

	auths, err := multipleFromEntities(dest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting entities")
	}
	return auths, nil
}

func multipleFromEntities(entities systemauth.Collection) ([]model.SystemAuth, error) {
	conv := systemauth.NewConverter(nil)
	var items []model.SystemAuth

	for _, ent := range entities {
		m, err := conv.FromEntity(ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating system auth model from entity")
		}

		items = append(items, m)
	}

	return items, nil
}
