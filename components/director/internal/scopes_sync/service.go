package scopes

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/ory/hydra-client-go/models"
	"github.com/pkg/errors"
)

// SyncService missing godoc
type SyncService interface {
	SynchronizeClientScopes(context.Context) error
}

// SystemAuthRepo missing godoc
//go:generate mockery --name=SystemAuthRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthRepo interface {
	ListGlobalWithConditions(ctx context.Context, conditions repo.Conditions) ([]model.SystemAuth, error)
}

// OAuthService missing godoc
//go:generate mockery --name=OAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OAuthService interface {
	ListClients() ([]*models.OAuth2Client, error)
	UpdateClient(ctx context.Context, clientID string, objectType model.SystemAuthReferenceObjectType) error
	GetClientDetails(objType model.SystemAuthReferenceObjectType) (*oauth20.ClientDetails, error)
}

type service struct {
	oAuth20Svc OAuthService
	transact   persistence.Transactioner
	repo       SystemAuthRepo
}

// NewService missing godoc
func NewService(oAuth20Svc OAuthService, transact persistence.Transactioner, repo SystemAuthRepo) SyncService {
	return &service{
		oAuth20Svc: oAuth20Svc,
		transact:   transact,
		repo:       repo,
	}
}

// SynchronizeClientScopes missing godoc
func (s *service) SynchronizeClientScopes(ctx context.Context) error {
	hydraClients, err := s.listHydraClients()
	if err != nil {
		return err
	}

	auths, err := s.systemAuthsWithOAuth(ctx)
	if err != nil {
		return err
	}

	areAllClientsUpdated := true
	for _, auth := range auths {
		log.C(ctx).Infof("Synchronizing oauth client of system auth with ID %s", auth.ID)
		if auth.Value == nil || auth.Value.Credential.Oauth == nil {
			log.C(ctx).Infof("System auth with ID %s does not have oauth client for update", auth.ID)
			continue
		}

		clientID := auth.Value.Credential.Oauth.ClientID

		objType, err := auth.GetReferenceObjectType()
		if err != nil {
			areAllClientsUpdated = false
			log.C(ctx).WithError(err).Errorf("Error while getting obj type of client with ID %s: %v", clientID, err)
			continue
		}

		requiredClientDetails, err := s.oAuth20Svc.GetClientDetails(objType)
		if err != nil {
			areAllClientsUpdated = false
			log.C(ctx).WithError(err).Errorf("Error while getting client credentials scopes for client with ID %s: %v", clientID, err)
			continue
		}

		clientDetails, ok := hydraClients[clientID]
		if !ok {
			log.C(ctx).Errorf("Client with ID %s is not present in Hydra", clientID)
			continue
		}
		if str.Matches(clientDetails.Scopes, requiredClientDetails.Scopes) && str.Matches(clientDetails.GrantTypes, requiredClientDetails.GrantTypes) {
			log.C(ctx).Infof("Scopes and grant types for client with ID %s and type %s are in sync", clientID, objType)
			continue
		}

		if err = s.oAuth20Svc.UpdateClient(ctx, clientID, objType); err != nil {
			areAllClientsUpdated = false
			log.C(ctx).WithError(err).Errorf("Error while updating scopes of client with ID %s: %v", clientID, err)
		}
	}
	if !areAllClientsUpdated {
		return errors.New("Not all clients were updated successfully")
	}

	log.C(ctx).Info("Finished synchronization of Hydra scopes")
	return nil
}

func (s *service) listHydraClients() (map[string]*oauth20.ClientDetails, error) {
	clients, err := s.oAuth20Svc.ListClients()
	if err != nil {
		return nil, errors.Wrap(err, "while listing clients from hydra")
	}

	clientsMap := make(map[string]*oauth20.ClientDetails)
	for _, c := range clients {
		clientsMap[c.ClientID] = &oauth20.ClientDetails{
			Scopes:     strings.Split(c.Scope, " "),
			GrantTypes: c.GrantTypes,
		}
	}
	return clientsMap, nil
}

func (s *service) systemAuthsWithOAuth(ctx context.Context) ([]model.SystemAuth, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening database transaction")
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	conditions := repo.Conditions{
		repo.NewNotEqualCondition("(value -> 'Credential' -> 'Oauth')", "null"),
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
