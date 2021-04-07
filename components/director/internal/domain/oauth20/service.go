package oauth20

import (
	"context"
	"fmt"
	"strings"

	"github.com/ory/hydra-client-go/models"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/pkg/errors"
)

const (
	clientCredentialScopesPrefix = "clientCredentialsRegistrationScopes"
)

var defaultGrantTypes = []string{"client_credentials"}

//go:generate mockery --name=ScopeCfgProvider --output=automock --outpkg=automock --case=underscore
type ScopeCfgProvider interface {
	GetRequiredScopes(path string) ([]string, error)
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

//go:generate mockery --name=OryHydraService --output=automock --outpkg=automock --case=underscore
type OryHydraService interface {
	ListOAuth2Clients(params *admin.ListOAuth2ClientsParams) (*admin.ListOAuth2ClientsOK, error)
	CreateOAuth2Client(params *admin.CreateOAuth2ClientParams) (*admin.CreateOAuth2ClientCreated, error)
	UpdateOAuth2Client(params *admin.UpdateOAuth2ClientParams) (*admin.UpdateOAuth2ClientOK, error)
	DeleteOAuth2Client(params *admin.DeleteOAuth2ClientParams) (*admin.DeleteOAuth2ClientNoContent, error)
}

type service struct {
	publicAccessTokenEndpoint string
	scopeCfgProvider          ScopeCfgProvider
	uidService                UIDService
	hydraCLi                  OryHydraService
}

func NewService(scopeCfgProvider ScopeCfgProvider, uidService UIDService, publicAccessTokenEndpoint string, hydraCLi OryHydraService) *service {
	return &service{
		scopeCfgProvider:          scopeCfgProvider,
		publicAccessTokenEndpoint: publicAccessTokenEndpoint,
		uidService:                uidService,
		hydraCLi:                  hydraCLi,
	}
}

func (s *service) CreateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error) {
	scopes, err := s.GetClientCredentialScopes(objectType)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Debugf("Fetched client credential scopes: %s for %s", scopes, objectType)

	clientID := s.uidService.Generate()
	clientSecret, err := s.registerClient(ctx, clientID, scopes)
	if err != nil {
		return nil, errors.Wrap(err, "while registering client credentials in Hydra")
	}

	credentialData := &model.OAuthCredentialDataInput{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          s.publicAccessTokenEndpoint,
	}

	return credentialData, nil
}

func (s *service) UpdateClientScopes(ctx context.Context, clientID string, objectType model.SystemAuthReferenceObjectType) error {
	scopes, err := s.GetClientCredentialScopes(objectType)
	if err != nil {
		return err
	}
	log.C(ctx).Debugf("Fetched Client credential scopes: %s for %s", scopes, objectType)

	if err := s.updateClient(ctx, clientID, scopes); err != nil {
		return errors.Wrapf(err, "while updating Client with ID %s in Hydra", clientID)
	}

	return nil
}

func (s *service) DeleteClientCredentials(ctx context.Context, clientID string) error {
	log.C(ctx).Debugf("Unregistering client_id %s and client_secret in Hydra", clientID)

	_, err := s.hydraCLi.DeleteOAuth2Client(admin.NewDeleteOAuth2ClientParams().WithID(clientID))
	if err != nil {
		return err
	}

	log.C(ctx).Debugf("client_id %s and client_secret successfully unregistered in Hydra", clientID)
	return nil
}

func (s *service) DeleteMultipleClientCredentials(ctx context.Context, auths []model.SystemAuth) error {
	for _, auth := range auths {
		if auth.Value == nil {
			continue
		}
		if auth.Value.Credential.Oauth == nil {
			continue
		}
		err := s.DeleteClientCredentials(ctx, auth.Value.Credential.Oauth.ClientID)
		if err != nil {
			return errors.Wrap(err, "while deleting OAuth 2.0 credentials")
		}
	}
	return nil
}

func (s *service) ListClients() ([]*models.OAuth2Client, error) {
	listClientsOK, err := s.hydraCLi.ListOAuth2Clients(admin.NewListOAuth2ClientsParams())
	if err != nil {
		return nil, err
	}
	return listClientsOK.Payload, nil
}

func (s *service) GetClientCredentialScopes(objType model.SystemAuthReferenceObjectType) ([]string, error) {
	scopes, err := s.scopeCfgProvider.GetRequiredScopes(s.buildPath(objType))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting scopes for registering Client Credentials for %s", objType)
	}

	return scopes, nil
}

func (s *service) registerClient(ctx context.Context, clientID string, scopes []string) (string, error) {
	log.C(ctx).Debugf("Registering client_id %s and client_secret in Hydra with scopes: %s", clientID, scopes)

	created, err := s.hydraCLi.CreateOAuth2Client(admin.NewCreateOAuth2ClientParams().WithBody(&models.OAuth2Client{
		ClientID:   clientID,
		GrantTypes: defaultGrantTypes,
		Scope:      strings.Join(scopes, " "),
	}))

	if err != nil {
		return "", err
	}
	log.C(ctx).Debugf("client_id %s and client_secret successfully registered in Hydra", clientID)
	return created.Payload.ClientSecret, nil
}

func (s *service) updateClient(ctx context.Context, clientID string, scopes []string) error {
	_, err := s.hydraCLi.UpdateOAuth2Client(admin.NewUpdateOAuth2ClientParams().WithID(clientID).WithBody(&models.OAuth2Client{
		Scope: strings.Join(scopes, " "),
	}))
	if err != nil {
		return err
	}
	log.C(ctx).Infof("Client with client_id %s successfully updated in Hydra", clientID)
	return nil
}

func (s *service) buildPath(objType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(objType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}
