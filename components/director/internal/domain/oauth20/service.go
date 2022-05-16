package oauth20

import (
	"context"
	"fmt"
	"strings"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/ory/hydra-client-go/models"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/ory/hydra-client-go/client/admin"
	"github.com/pkg/errors"
)

const (
	scopesPerConsumerTypePrefix      = "scopesPerConsumerType"
	clientCredentialGrantTypesPrefix = "clientCredentialsRegistrationGrantTypes"
)

// ClientDetailsConfigProvider missing godoc
//go:generate mockery --name=ClientDetailsConfigProvider --output=automock --outpkg=automock --case=underscore --disable-version-string
type ClientDetailsConfigProvider interface {
	GetRequiredScopes(path string) ([]string, error)
	GetRequiredGrantTypes(path string) ([]string, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// OryHydraService missing godoc
//go:generate mockery --name=OryHydraService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OryHydraService interface {
	ListOAuth2Clients(params *admin.ListOAuth2ClientsParams, opts ...admin.ClientOption) (*admin.ListOAuth2ClientsOK, error)
	CreateOAuth2Client(params *admin.CreateOAuth2ClientParams, opts ...admin.ClientOption) (*admin.CreateOAuth2ClientCreated, error)
	UpdateOAuth2Client(params *admin.UpdateOAuth2ClientParams, opts ...admin.ClientOption) (*admin.UpdateOAuth2ClientOK, error)
	DeleteOAuth2Client(params *admin.DeleteOAuth2ClientParams, opts ...admin.ClientOption) (*admin.DeleteOAuth2ClientNoContent, error)
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
	hydraCLi                  OryHydraService
}

// NewService missing godoc
func NewService(scopeCfgProvider ClientDetailsConfigProvider, uidService UIDService, publicAccessTokenEndpoint string, hydraCLi OryHydraService) *service {
	return &service{
		scopeCfgProvider:          scopeCfgProvider,
		publicAccessTokenEndpoint: publicAccessTokenEndpoint,
		uidService:                uidService,
		hydraCLi:                  hydraCLi,
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

	_, err := s.hydraCLi.DeleteOAuth2Client(admin.NewDeleteOAuth2ClientParams().WithID(clientID))
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
func (s *service) ListClients() ([]*models.OAuth2Client, error) {
	listClientsOK, err := s.hydraCLi.ListOAuth2Clients(admin.NewListOAuth2ClientsParams())
	if err != nil {
		return nil, err
	}
	return listClientsOK.Payload, nil
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

	created, err := s.hydraCLi.CreateOAuth2Client(admin.NewCreateOAuth2ClientParams().WithBody(&models.OAuth2Client{
		ClientID:   clientID,
		GrantTypes: details.GrantTypes,
		Scope:      strings.Join(details.Scopes, " "),
	}))

	if err != nil {
		return "", err
	}
	log.C(ctx).Debugf("client_id %s and client_secret successfully registered in Hydra", clientID)
	return created.Payload.ClientSecret, nil
}

func (s *service) updateClient(ctx context.Context, clientID string, details *ClientDetails) error {
	_, err := s.hydraCLi.UpdateOAuth2Client(admin.NewUpdateOAuth2ClientParams().WithID(clientID).WithBody(&models.OAuth2Client{
		ClientID:   clientID,
		GrantTypes: details.GrantTypes,
		Scope:      strings.Join(details.Scopes, " "),
	}))
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
