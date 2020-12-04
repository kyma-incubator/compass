package fetchrequest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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

func (s *service) HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string {
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

	resp, err := s.client.Get(fr.URL)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while fetching API Spec.")
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
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading API Spec.")
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While reading API Spec: %s", err.Error())))
	}

	spec := string(body)
	return &spec, s.fixStatus(model.FetchRequestStatusConditionSucceeded, nil)
}

func (s *service) validateFetchRequest(fr *model.FetchRequest) error {
	if fr.Mode != model.FetchModeSingle {
		return apperrors.NewInvalidDataError("Unsupported fetch mode: %s", fr.Mode)
	}

	if fr.Auth != nil {
		return apperrors.NewInvalidDataError("Auth for Fetch Request was provided, currently it's unsupported")
	}

	if fr.Filter != nil {
		return apperrors.NewInvalidDataError("Filter for Fetch Request was provided, currently it's unsupported")
	}
	return nil
}

func (s *service) fixStatus(condition model.FetchRequestStatusCondition, message *string) *model.FetchRequestStatus {
	return &model.FetchRequestStatus{
		Condition: condition,
		Message:   message,
		Timestamp: s.timestampGen(),
	}
}
