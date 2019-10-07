package oauth20

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

const clientCredentialScopesPrefix = "__clientCredentialsRegistrationScopes"
const applicationJSONType = "application/json"

var defaultGrantTypes = []string{"client_credentials"}

//go:generate mockery -name=ScopeCfgProvider -output=automock -outpkg=automock -case=underscore
type ScopeCfgProvider interface {
	GetRequiredScopes(path string) ([]string, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	clientCreationEndpoint    string
	clientDeletionEndpoint    string
	publicAccessTokenEndpoint string
	scopeCfgProvider          ScopeCfgProvider
	httpCli                   *http.Client
	uidService                UIDService
}

func NewService(scopeCfgProvider ScopeCfgProvider, uidService UIDService, cfg Config) *service {
	return &service{
		scopeCfgProvider:          scopeCfgProvider,
		clientCreationEndpoint:    cfg.ClientCreationEndpoint,
		clientDeletionEndpoint:    cfg.ClientDeletionEndpoint,
		publicAccessTokenEndpoint: cfg.PublicAccessTokenEndpoint,
		httpCli:                   &http.Client{},
		uidService:                uidService,
	}
}

func (s *service) CreateClient(ctx context.Context, objectType model.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error) {
	scopes, err := s.getClientCredentialScopes(objectType)
	if err != nil {
		return nil, err
	}

	clientID := s.uidService.Generate()
	clientSecret, err := s.registerClient(clientID, scopes)
	if err != nil {
		return nil, err
	}

	credentialData := &model.OAuthCredentialDataInput{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          s.publicAccessTokenEndpoint,
	}

	return credentialData, nil
}

func (s *service) DeleteClient(ctx context.Context, clientID string) error {
	// TODO: DeleteClient
}

func (s *service) getClientCredentialScopes(objType model.SystemAuthReferenceObjectType) ([]string, error) {
	scopes, err := s.scopeCfgProvider.GetRequiredScopes(s.buildPath(objType))
	if err != nil {
		return nil, errors.Wrap(err, "while getting scopes for registering Client Credentials")
	}

	return scopes, nil
}

type clientCredentialsRegistrationBody struct {
	GrantTypes   []string `json:"grant_types"`
	ClientID     string   `json:"client_id"`
	Scope        string   `json:"scope"`
}

type clientCredentialsRegistrationResponse struct {
	ClientSecret     string   `json:"client_secret"`
}

func (s *service) registerClient(clientID string, scopes []string) (string, error) {
	reqBody := &clientCredentialsRegistrationBody{
		GrantTypes:   defaultGrantTypes,
		ClientID:     clientID,
		Scope:        strings.Join(scopes, " "),
	}

	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(&reqBody)
	if err != nil {
		return "", errors.Wrap(err, "while encoding body")
	}

	req, err := http.NewRequest(http.MethodPost, s.clientCreationEndpoint, buffer)
	if err != nil {
		return "", errors.Wrap(err, "while creating new request")
	}

	req.Header.Set("Content-Type", applicationJSONType)
	req.Header.Set("Accept", applicationJSONType)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return "", errors.Wrapf(err, "while doing request to %s", s.clientCreationEndpoint)
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			logrus.Error(err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("invalid HTTP status code: received: %d,  expected %d", resp.StatusCode, http.StatusCreated)
	}

	var registrationResp clientCredentialsRegistrationResponse
	err = json.NewDecoder(resp.Body).Decode(&registrationResp)
	if err != nil {
		return "", errors.Wrap(err, "while decoding response body")
	}

	return registrationResp.ClientSecret, nil
}

func (s *service) buildPath(objType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(objType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s/%s", clientCredentialScopesPrefix, transformedObjType)
}
