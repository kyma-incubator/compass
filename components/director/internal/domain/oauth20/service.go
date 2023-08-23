package oauth20

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	hydraClient "github.com/ory/hydra-client-go/v2"
)

const (
	scopesPerConsumerTypePrefix      = "scopesPerConsumerType"
	clientCredentialGrantTypesPrefix = "clientCredentialsRegistrationGrantTypes"
)

// ClientDetailsConfigProvider missing godoc
//
//go:generate mockery --name=ClientDetailsConfigProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientDetailsConfigProvider interface {
	GetRequiredScopes(path string) ([]string, error)
	GetRequiredGrantTypes(path string) ([]string, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// OryHydraService missing godoc
//
//go:generate mockery --name=OryHydraService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OryHydraService interface {
	ListOAuth2Clients(ctx context.Context) hydraClient.OAuth2ApiListOAuth2ClientsRequest
	ListOAuth2ClientsExecute(r hydraClient.OAuth2ApiListOAuth2ClientsRequest) ([]hydraClient.OAuth2Client, *http.Response, error)

	DeleteOAuth2Client(ctx context.Context, id string) hydraClient.OAuth2ApiDeleteOAuth2ClientRequest
	DeleteOAuth2ClientExecute(r hydraClient.OAuth2ApiDeleteOAuth2ClientRequest) (*http.Response, error)

	CreateOAuth2Client(ctx context.Context) hydraClient.OAuth2ApiCreateOAuth2ClientRequest
	CreateOAuth2ClientExecute(r hydraClient.OAuth2ApiCreateOAuth2ClientRequest) (*hydraClient.OAuth2Client, *http.Response, error)

	SetOAuth2Client(ctx context.Context, id string) hydraClient.OAuth2ApiSetOAuth2ClientRequest
	SetOAuth2ClientExecute(r hydraClient.OAuth2ApiSetOAuth2ClientRequest) (*hydraClient.OAuth2Client, *http.Response, error)
}

// ClientDetails missing godoc
type ClientDetails struct {
	Scopes     []string
	GrantTypes []string
}

type service struct {
	publicAccessTokenEndpoint string
	scopeCfgProvider          ClientDetailsConfigProvider
	uidService                UIDService
	hydraService              OryHydraService
}

// NewService missing godoc
func NewService(scopeCfgProvider ClientDetailsConfigProvider, uidService UIDService, publicAccessTokenEndpoint string, hydraService OryHydraService) *service {
	return &service{
		scopeCfgProvider:          scopeCfgProvider,
		publicAccessTokenEndpoint: publicAccessTokenEndpoint,
		uidService:                uidService,
		hydraService:              hydraService,
	}
}

// CreateClientCredentials missing godoc
func (s *service) CreateClientCredentials(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error) {
	details, err := s.GetClientDetails(objectType)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Debugf("Fetched client credential scopes: %s for %s", details.Scopes, objectType)

	clientID := s.uidService.Generate()
	clientSecret, err := s.registerClient(ctx, clientID, details)
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

// UpdateClient missing godoc
func (s *service) UpdateClient(ctx context.Context, clientID string, objectType pkgmodel.SystemAuthReferenceObjectType) error {
	details, err := s.GetClientDetails(objectType)
	if err != nil {
		return err
	}
	log.C(ctx).Debugf("Fetched Client credential scopes: %s for %s", details.Scopes, objectType)

	if err := s.updateClient(ctx, clientID, details); err != nil {
		return errors.Wrapf(err, "while updating Client with ID %s in Hydra", clientID)
	}

	return nil
}

// DeleteClientCredentials missing godoc
func (s *service) DeleteClientCredentials(ctx context.Context, clientID string) error {
	log.C(ctx).Debugf("Unregistering client_id %s and client_secret in Hydra", clientID)

	request := s.hydraService.DeleteOAuth2Client(ctx, clientID)
	_, err := s.hydraService.DeleteOAuth2ClientExecute(request)

	if err != nil {
		return err
	}

	log.C(ctx).Debugf("client_id %s and client_secret successfully unregistered in Hydra", clientID)
	return nil
}

// DeleteMultipleClientCredentials missing godoc
func (s *service) DeleteMultipleClientCredentials(ctx context.Context, auths []pkgmodel.SystemAuth) error {
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

// ListClients missing godoc
func (s *service) ListClients() ([]hydraClient.OAuth2Client, error) {
	request := s.hydraService.ListOAuth2Clients(context.Background())
	clients, _, err := s.hydraService.ListOAuth2ClientsExecute(request)

	if err != nil {
		return nil, err
	}

	return clients, nil
}

// GetClientDetails missing godoc
func (s *service) GetClientDetails(objType pkgmodel.SystemAuthReferenceObjectType) (*ClientDetails, error) {
	scopes, err := s.scopeCfgProvider.GetRequiredScopes(s.buildPath(objType))
	if err != nil {
		return nil, errors.Wrapf(err, "while getting scopes for registering Client Credentials for %s", objType)
	}

	grantTypes, err := s.scopeCfgProvider.GetRequiredGrantTypes(clientCredentialGrantTypesPrefix)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting grant_types for registering Client Credentials for %s", objType)
	}

	return &ClientDetails{
		Scopes:     scopes,
		GrantTypes: grantTypes,
	}, nil
}

func (s *service) registerClient(ctx context.Context, clientID string, details *ClientDetails) (string, error) {
	log.C(ctx).Debugf("Registering client_id %s and client_secret in Hydra with scopes: %s and grant_types %s", clientID, details.Scopes, details.GrantTypes)

	scopes := strings.Join(details.Scopes, " ")
	clientToCreate := hydraClient.OAuth2Client{
		ClientId:   &clientID,
		GrantTypes: details.GrantTypes,
		Scope:      &scopes,
	}

	request := s.hydraService.CreateOAuth2Client(ctx).OAuth2Client(clientToCreate)
	createdClient, _, err := s.hydraService.CreateOAuth2ClientExecute(request)

	if err != nil {
		return "", err
	}
	log.C(ctx).Debugf("client_id %s and client_secret successfully registered in Hydra", clientID)
	return *createdClient.ClientSecret, nil
}

func (s *service) updateClient(ctx context.Context, clientID string, details *ClientDetails) error {
	scopes := strings.Join(details.Scopes, " ")

	clientToUpgrade := hydraClient.OAuth2Client{
		ClientId:   &clientID,
		GrantTypes: details.GrantTypes,
		Scope:      &scopes,
	}

	request := s.hydraService.SetOAuth2Client(ctx, clientID).OAuth2Client(clientToUpgrade)
	_, _, err := s.hydraService.SetOAuth2ClientExecute(request)

	if err != nil {
		return err
	}
	log.C(ctx).Infof("Client with client_id %s successfully updated in Hydra", clientID)
	return nil
}

func (s *service) buildPath(objType pkgmodel.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(objType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", scopesPerConsumerTypePrefix, transformedObjType)
}
