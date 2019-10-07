package oauth20

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

const clientCredentialScopesPrefix = "__clientCredentialsRegistrationScopes"
const applicationJSONType = "application/json"

var defaultGrantTypes = []string{"client_credentials"}

type ScopeCfgProvider interface {
	GetRequiredScopes(path string) ([]string, error)
}

type service struct {
	clientCredsRegistrationEndpoint string
	scopeCfgProvider                ScopeCfgProvider
	httpCli                         *http.Client
}

func NewService(scopeCfgProvider ScopeCfgProvider, clientCredentialsRegistrationEndpoint string) *service {
	return &service{scopeCfgProvider: scopeCfgProvider, clientCredsRegistrationEndpoint: clientCredentialsRegistrationEndpoint, httpCli: &http.Client{}}
}

func (s *service) GenerateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) (*model.OAuthCredentialDataInput, error) {
	// generate credentials
	credentialData := &model.OAuthCredentialDataInput{
		ClientID: "",
		ClientSecret: "",
		URL: ""
	}



	return credentialData, nil
}

func (s *service) RegisterClientCredentials(ctx context.Context, dataInput *model.OAuthCredentialDataInput, objectType model.SystemAuthReferenceObjectType) error {
	scopes, err := s.getClientCredentialScopes(objectType)
	if err != nil {
		return err
	}

	err = s.registerClientCredentials(dataInput, scopes)
	if err != nil {
		return err
	}

	return nil
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
	ClientSecret string   `json:"client_secret"`
	Scope        string   `json:"scope"`
}

func (s *service) registerClientCredentials(input *model.OAuthCredentialDataInput, scopes []string) error {
	reqBody := &clientCredentialsRegistrationBody{
		GrantTypes:   defaultGrantTypes,
		ClientID:     input.ClientID,
		ClientSecret: input.ClientSecret,
		Scope:        strings.Join(scopes, " "),
	}

	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(&reqBody)
	if err != nil {
		return errors.Wrap(err, "while encoding body")
	}

	req, err := http.NewRequest(http.MethodPost, s.clientCredsRegistrationEndpoint, buffer)
	if err != nil {
		return errors.Wrap(err, "while creating new request")
	}

	req.Header.Set("Content-Type", applicationJSONType)
	req.Header.Set("Accept", applicationJSONType)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while doing request to %s", s.clientCredsRegistrationEndpoint)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("invalid HTTP status code: received: %d,  expected %d", resp.StatusCode, http.StatusCreated)
	}
}

func (s *service) buildPath(objType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(objType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s/%s", clientCredentialScopesPrefix, transformedObjType)
}
