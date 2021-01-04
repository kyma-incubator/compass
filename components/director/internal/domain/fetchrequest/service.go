package fetchrequest

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const WellKnown = "/.well-known/openid-configuration"

type service struct {
	repo         FetchRequestRepository
	client       *http.Client
	timestampGen timestamp.Generator
	issuers      *issuers
}

type OpenIDMetadata struct {
	TokenEndpoint string `json:"token_endpoint"`
}

type issuers struct {
	knownURLs map[string]*OpenIDMetadata
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Update(ctx context.Context, item *model.FetchRequest) error
}

func NewService(repo FetchRequestRepository, client *http.Client) *service {
	return &service{
		repo:         repo,
		client:       client,
		timestampGen: timestamp.DefaultGenerator(),
		issuers: &issuers{
			knownURLs: make(map[string]*OpenIDMetadata),
		},
	}
}

func (s *service) HandleSpec(ctx context.Context, fr *model.FetchRequest) *string {
	var data *string
	data, fr.Status = s.fetchAPISpec(ctx, fr)

	err := s.repo.Update(ctx, fr)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while updating fetch request status.")
		return nil
	}

	return data
}

func (s *service) fetchAPISpec(ctx context.Context, fr *model.FetchRequest) (*string, *model.FetchRequestStatus) {

	err := s.validateFetchRequest(fr)
	if err != nil {
		log.C(ctx).WithError(err).Error()
		return nil, s.fixStatus(model.FetchRequestStatusConditionInitial, str.Ptr(err.Error()))
	}

	var resp *http.Response

	if fr.Auth != nil {
		resp, err = s.requestWithCredentials(ctx, fr)
	} else {
		resp, err = s.requestWithoutCredentials(fr)
	}

	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while fetching Spec.")
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching API Spec: %s", err.Error())))
	}

	defer func() {
		if resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body.")
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("While fetching API Spec status code: %d", resp.StatusCode)
		log.C(ctx).Errorf(errMsg)
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching API Spec status code: %d", resp.StatusCode)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading Spec.")
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While reading API Spec: %s", err.Error())))
	}

	spec := string(body)
	return &spec, s.fixStatus(model.FetchRequestStatusConditionSucceeded, nil)
}

func (s *service) validateFetchRequest(fr *model.FetchRequest) error {
	if fr.Mode != model.FetchModeSingle {
		return apperrors.NewInvalidDataError("Unsupported fetch mode: %s", fr.Mode)
	}

	if fr.Filter != nil {
		return apperrors.NewInvalidDataError("Filter for Fetch Request was provided, currently it's unsupported")
	}

	return nil
}

func (s *service) requestWithCredentials(ctx context.Context, fr *model.FetchRequest) (*http.Response, error) {
	if fr.Auth.Credential.Basic == nil && fr.Auth.Credential.Oauth == nil{
		return nil, apperrors.NewInvalidDataError("Credentials not provided")
	}

	var err error
	var resp *http.Response
	if fr.Auth.Credential.Basic != nil {
		resp, err = s.requestWithBasicCredentials(fr)
		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
	}

	if fr.Auth.Credential.Oauth != nil {
		resp, err = s.requestWithOauth(ctx, fr)
	}

	return resp, err
}

func (s *service) requestWithBasicCredentials(fr *model.FetchRequest) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fr.URL, nil)
	if err != nil {
		return nil, err
	}

	basicCred := fr.Auth.Credential.Basic
	req.SetBasicAuth(basicCred.Username, basicCred.Password)
	return s.client.Do(req)
}

func (s *service) requestWithOauth(ctx context.Context, fr *model.FetchRequest) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fr.URL, nil)
	if err != nil {
		return nil, err
	}

	token, err := s.getToken(ctx, fr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return s.client.Do(req)
}

func (s *service) getToken(ctx context.Context, fr *model.FetchRequest) (*oauth2.Token, error) {
	tokenEndpoint, err := s.issuers.getTokenEndpoint(s.client, fr.Auth.Credential.Oauth.URL)
	if err != nil {
		return nil, err
	}

	conf := &clientcredentials.Config{
		ClientID:     fr.Auth.Credential.Oauth.ClientID,
		ClientSecret: fr.Auth.Credential.Oauth.ClientSecret,
		TokenURL:     tokenEndpoint,
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.client)

	return conf.Token(ctx)
}

func (i *issuers) getTokenEndpoint(client *http.Client, issuerUrl string) (string, error) {
	rwMutex := &sync.RWMutex{}
	rwMutex.RLock()
	tokenEndpoint, found := i.knownURLs[issuerUrl]
	rwMutex.RUnlock()
	var err error
	if !found {
		tokenEndpoint, err = fetchTokenEndpoint(client, issuerUrl)
		if err != nil {
			return "", err
		}
		rwMutex.Lock()
		i.knownURLs[issuerUrl] = tokenEndpoint
		rwMutex.Unlock()
	}
	return tokenEndpoint.TokenEndpoint, nil
}

func fetchTokenEndpoint(client *http.Client, URL string) (*OpenIDMetadata, error) {
	request, err := http.NewRequest(http.MethodGet, URL+WellKnown, nil)
	if err != nil {
		return nil, err
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	t := &OpenIDMetadata{}

	err = json.NewDecoder(response.Body).Decode(t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *service) requestWithoutCredentials(fr *model.FetchRequest) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fr.URL, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req)
}

func (s *service) fixStatus(condition model.FetchRequestStatusCondition, message *string) *model.FetchRequestStatus {
	return &model.FetchRequestStatus{
		Condition: condition,
		Message:   message,
		Timestamp: s.timestampGen(),
	}
}
