package fetchrequest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	client       *http.Client
	logger       *log.Logger
	timestampGen timestamp.Generator
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Update(ctx context.Context, item *model.FetchRequest) error
}

func NewService(client *http.Client, logger *log.Logger) *service {
	return &service{
		client:       client,
		logger:       logger,
		timestampGen: timestamp.DefaultGenerator(),
	}
}

func (s *service) SetTimestampGen(timestampGen func() time.Time) {
	s.timestampGen = timestampGen
}

func (s *service) FetchAPISpec(fr *model.FetchRequest) (*string, *model.FetchRequestStatus, error) {

	valid, msg := s.validateFetchRequest(fr)
	if !valid {
		return nil, s.fixStatus(model.FetchRequestStatusConditionInitial, msg), nil
	}

	resp, err := s.client.Get(fr.URL)
	if err != nil {
		s.logger.Errorf("While fetching API Spec: %s", err.Error())
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, fmt.Sprintf("While fetching API Spec: %s", err.Error())), err
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
		s.logger.Errorf("While fetching API Spec status code: %d", resp.StatusCode)
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, fmt.Sprintf("While fetching API Spec status code: %d", resp.StatusCode)), errors.New("")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("While reading API Spec: %s", err.Error())
		return nil, s.fixStatus(model.FetchRequestStatusConditionFailed, fmt.Sprintf("While reading API Spec: %s", err.Error())), err
	}

	spec := string(body)
	return &spec, s.fixStatus(model.FetchRequestStatusConditionSucceeded, ""), nil
}

func (s *service) validateFetchRequest(fr *model.FetchRequest) (bool, string) {
	if fr.Mode != model.FetchModeSingle {
		s.logger.Errorf("Unsupported fetch mode: %s", fr.Mode)
		return false, fmt.Sprintf("Unsupported fetch mode: %s", fr.Mode)
	}

	if fr.Auth != nil {
		s.logger.Error("Fetch Request Auth was provided")
		return false, "Fetch Request Auth was provided"
	}

	if fr.Filter != nil {
		s.logger.Error("Fetch Request Filter was provided")
		return false, "Fetch Request Filter was provided"
	}
	return true, ""
}

func (s *service) fixStatus(condition model.FetchRequestStatusCondition, message string) *model.FetchRequestStatus {
	return &model.FetchRequestStatus{
		Condition: condition,
		Message:   &message,
		Timestamp: s.timestampGen(),
	}
}
