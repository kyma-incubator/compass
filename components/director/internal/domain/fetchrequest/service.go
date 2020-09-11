package fetchrequest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	repo         FetchRequestRepository
	client       *http.Client
	logger       *log.Logger
	timestampGen timestamp.Generator
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Update(ctx context.Context, item *model.FetchRequest) error
}

func NewService(repo FetchRequestRepository, client *http.Client, logger *log.Logger) *service {
	return &service{
		repo:         repo,
		client:       client,
		logger:       logger,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) HandleAPISpec(ctx context.Context, fr *model.FetchRequest) *string {
	var data *string
	data, fr.Status = s.fetchAPISpec(fr)

	err := s.repo.Update(ctx, fr)
	if err != nil {
		s.logger.Errorf("While updating fetch request status: %s", err)
		return nil
	}

	return data
}

func (s *service) fetchAPISpec(fr *model.FetchRequest) (*string, *model.FetchRequestStatus) {

	err := s.validateFetchRequest(fr)
	if err != nil {
		s.logger.Error(err)
		return nil, s.fixStatus(model.FetchRequestStatusConditionInitial, str.Ptr(err.Error()))
	}

	resp, err := s.client.Get(fr.URL)
	if err != nil {
		s.logger.Errorf("While fetching API Specs: %s", err.Error())
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching API Specs: %s", err.Error())))
	}

	defer func() {
		if resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				s.logger.Errorf("While closing body: %s", err.Error())
			}
		}
	}()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("While fetching API Specs status code: %d", resp.StatusCode)
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching API Specs status code: %d", resp.StatusCode)))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("While reading API Specs: %s", err.Error())
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While reading API Specs: %s", err.Error())))
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
