package fetchrequest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	repo         FetchRequestRepository
	client       *http.Client
	timestampGen timestamp.Generator
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
	}
}

func (s *service) HandleSpec(ctx context.Context, fr *model.FetchRequest) *string {
	var data *string
	data, fr.Status = s.fetchSpec(ctx, fr)

	err := s.repo.Update(ctx, fr)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while updating fetch request status.")
		return nil
	}

	return data
}

func (s *service) fetchSpec(ctx context.Context, fr *model.FetchRequest) (*string, *model.FetchRequestStatus) {
	err := s.validateFetchRequest(fr)
	if err != nil {
		log.C(ctx).WithError(err).Error()
		return nil, FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr(err.Error()), s.timestampGen())
	}

	var resp *http.Response
	if fr.Auth != nil {
		resp, err = s.requestWithCredentials(ctx, fr)
	} else {
		resp, err = s.requestWithoutCredentials(fr)
	}

	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while fetching Spec.")
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec: %s", err.Error())), s.timestampGen())
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
		errMsg := fmt.Sprintf("While fetching Spec status code: %d", resp.StatusCode)
		log.C(ctx).Errorf(errMsg)
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec status code: %d", resp.StatusCode)), s.timestampGen())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading Spec.")
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While reading Spec: %s", err.Error())), s.timestampGen())
	}

	spec := string(body)
	return &spec, FixStatus(model.FetchRequestStatusConditionSucceeded, nil, s.timestampGen())
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
	if fr.Auth.Credential.Basic == nil && fr.Auth.Credential.Oauth == nil {
		return nil, apperrors.NewInvalidDataError("Credentials not provided")
	}

	req, err := http.NewRequest(http.MethodGet, fr.URL, nil)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if fr.Auth.Credential.Basic != nil {
		req.SetBasicAuth(fr.Auth.Credential.Basic.Username, fr.Auth.Credential.Basic.Password)

		resp, err = s.client.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
	}

	if fr.Auth.Credential.Oauth != nil {
		resp, err = s.secureClient(ctx, fr).Do(req)
	}

	return resp, err
}

func (s *service) secureClient(ctx context.Context, fr *model.FetchRequest) *http.Client {
	conf := &clientcredentials.Config{
		ClientID:     fr.Auth.Credential.Oauth.ClientID,
		ClientSecret: fr.Auth.Credential.Oauth.ClientSecret,
		TokenURL:     fr.Auth.Credential.Oauth.URL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, s.client)
	securedClient := conf.Client(ctx)
	securedClient.Timeout = s.client.Timeout
	return securedClient
}

func (s *service) requestWithoutCredentials(fr *model.FetchRequest) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fr.URL, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req)
}

func FixStatus(condition model.FetchRequestStatusCondition, message *string, timestamp time.Time) *model.FetchRequestStatus {
	return &model.FetchRequestStatus{
		Condition: condition,
		Message:   message,
		Timestamp: timestamp,
	}
}
