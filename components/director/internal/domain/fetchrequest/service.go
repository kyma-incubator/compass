package fetchrequest

import (
	"context"
	"errors"
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

func (s *service) FetchAPISpec(fr *model.FetchRequest) (*string, error) {

	if fr.Mode != model.FetchModeSingle || fr.Auth != nil {
		return nil, nil
	}

	resp, err := s.client.Get(fr.URL)
	if err != nil {
		fr.Status.Condition = model.FetchRequestStatusConditionFailed
		s.logger.Errorf("While fetching API Spec: %s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf("While fetching API Spec status code: %d", resp.StatusCode)
		return nil, errors.New("")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.logger.Errorf("While closing body: %s", err.Error())
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.logger.Errorf("While reading API Spec: %s", err.Error())
		return nil, err
	}
	fr.Status.Timestamp = s.timestampGen()
	spec := string(body)
	return &spec, nil

}
