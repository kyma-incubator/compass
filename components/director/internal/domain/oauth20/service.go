package oauth20

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	clientCredentialScopesPrefix = "clientCredentialsRegistrationScopes"
	applicationJSONType          = "application/json"
)

var defaultGrantTypes = []string{"client_credentials"}

//go:generate mockery -name=ScopeCfgProvider -output=automock -outpkg=automock -case=underscore
type ScopeCfgProvider interface {
	GetRequiredScopes(path string) ([]string, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type clientCredentialsRegistrationBody struct {
	GrantTypes []string `json:"grant_types,omitempty"`
	ClientID   string   `json:"client_id,omitempty"`
	Scope      string   `json:"scope"`
}

type clientCredentialsRegistrationResponse struct {
	ClientSecret string `json:"client_secret"`
}

type Client struct {
	ClientID string `json:"client_id"`
	Scopes   string `json:"scope"`
}

type service struct {
	clientEndpoint            string
	publicAccessTokenEndpoint string
	scopeCfgProvider          ScopeCfgProvider
	httpCli                   *http.Client
	uidService                UIDService
}

func NewService(scopeCfgProvider ScopeCfgProvider, uidService UIDService, cfg Config, httpCli *http.Client) *service {
	return &service{
		scopeCfgProvider:          scopeCfgProvider,
		clientEndpoint:            cfg.ClientEndpoint,
		publicAccessTokenEndpoint: cfg.PublicAccessTokenEndpoint,
		httpCli:                   httpCli,
		uidService:                uidService,
	}
}

func (s *service) CreateClientCredentials(ctx context.Context, objectType model.SystemAuthReferenceObjectType) (*model.OAuthCredentialDataInput, error) {
	scopes, err := s.GetClientCredentialScopes(objectType)
	if err != nil {
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return nil, err
		}
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
		if !model.IsIntegrationSystemNoTenantFlow(err, objectType) {
			return err
		}
	}
	log.C(ctx).Debugf("Fetched Client credential scopes: %s for %s", scopes, objectType)

	if err := s.updateClient(ctx, clientID, scopes); err != nil {
		return errors.Wrapf(err, "while updating Client with ID %s in Hydra", clientID)
	}

	return nil
}

func (s *service) DeleteClientCredentials(ctx context.Context, clientID string) error {
	return s.unregisterClient(ctx, clientID)
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

func (s *service) ListClients(ctx context.Context) ([]Client, error) {
	return s.clientsFromHydra(ctx, s.clientEndpoint)
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
	reqBody := &clientCredentialsRegistrationBody{
		GrantTypes: defaultGrantTypes,
		ClientID:   clientID,
		Scope:      strings.Join(scopes, " "),
	}

	buffer := &bytes.Buffer{}
	err := json.NewEncoder(buffer).Encode(&reqBody)
	if err != nil {
		return "", errors.Wrap(err, "while encoding body")
	}

	resp, closeBody, err := s.doRequest(ctx, http.MethodPost, s.clientEndpoint, buffer)
	if err != nil {
		return "", err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("invalid HTTP status code: received: %d, expected %d", resp.StatusCode, http.StatusCreated)
	}

	var registrationResp clientCredentialsRegistrationResponse
	err = json.NewDecoder(resp.Body).Decode(&registrationResp)
	if err != nil {
		return "", errors.Wrap(err, "while decoding response body")
	}

	log.C(ctx).Debugf("client_id %s and client_secret successfully registered in Hydra", clientID)
	return registrationResp.ClientSecret, nil
}

func (s *service) updateClient(ctx context.Context, clientID string, scopes []string) error {
	log.C(ctx).Infof("Updating client with client_id %s in Hydra with scopes: %s", clientID, scopes)
	reqBody := &clientCredentialsRegistrationBody{
		Scope: strings.Join(scopes, " "),
	}

	buffer := &bytes.Buffer{}
	if err := json.NewEncoder(buffer).Encode(&reqBody); err != nil {
		return errors.Wrap(err, "while encoding body")
	}

	resp, closeBody, err := s.doRequest(ctx, http.MethodPut, fmt.Sprintf("%s/%s", s.clientEndpoint, clientID), buffer)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid HTTP status code: received: %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	log.C(ctx).Debugf("Client with client_id %s successfully updated in Hydra", clientID)
	return nil
}

func (s *service) unregisterClient(ctx context.Context, clientID string) error {
	log.C(ctx).Debugf("Unregistering client_id %s and client_secret in Hydra", clientID)
	endpoint := fmt.Sprintf("%s/%s", s.clientEndpoint, clientID)

	resp, closeBody, err := s.doRequest(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	defer closeBody(resp.Body)

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("invalid HTTP status code: received: %d, expected %d", resp.StatusCode, http.StatusNoContent)
	}

	log.C(ctx).Debugf("client_id %s and client_secret successfully unregistered in Hydra", clientID)
	return nil
}

func (s *service) doRequest(ctx context.Context, method string, endpoint string, body io.Reader) (*http.Response, func(body io.ReadCloser), error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating new request")
	}

	req.Header.Set("Accept", applicationJSONType)
	req.Header.Set("Content-Type", applicationJSONType)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "while doing request to %s", s.clientEndpoint)
	}

	closeBodyFn := func(body io.ReadCloser) {
		if body == nil {
			return
		}
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			log.C(ctx).WithError(err).Error("An error has occurred while copying response body.")
		}

		err := body.Close()
		if err != nil {
			log.C(ctx).WithError(err).Error("An error has occurred while closing body.")
		}
	}

	return resp, closeBodyFn, nil
}

func (s *service) clientsFromHydra(ctx context.Context, endpoint string) ([]Client, error) {
	var resultClients []Client
	nextLink := endpoint
	for nextLink != "" {
		log.C(ctx).Debugf("Calling %s", nextLink)
		resp, closeBody, err := s.doRequest(ctx, http.MethodGet, nextLink, nil)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("invalid HTTP status code: received: %d, expected %d", resp.StatusCode, http.StatusOK)
		}

		var clients []Client
		if err = json.NewDecoder(resp.Body).Decode(&clients); err != nil {
			return nil, fmt.Errorf("failed to decode response body from Hydra: %v", err)
		}
		closeBody(resp.Body)
		resultClients = append(resultClients, clients...)
		if nextLink = getNextLink(resp); nextLink != "" {
			nextLink = endpoint + nextLink
		}
	}

	return resultClients, nil
}

func getNextLink(resp *http.Response) string {
	linkHeader := resp.Header.Get("Link")
	if linkHeader == "" {
		return ""
	}

	lastLinkRgx := regexp.MustCompile(`</clients(.+)>; rel="last"`)
	lastLink := lastLinkRgx.FindStringSubmatch(linkHeader)
	if len(lastLink) == 0 {
		return ""
	}
	nextLinkRgx := regexp.MustCompile(`</clients(.+)>; rel="next"`)
	nextLink := nextLinkRgx.FindStringSubmatch(linkHeader)
	if len(nextLink) == 0 {
		return ""
	}

	nextLinkPath := nextLink[1]
	return nextLinkPath
}

func (s *service) buildPath(objType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(objType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}
