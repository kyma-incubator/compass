package fetchrequest

import (
	"context"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	repo   FetchRequestRepository
	client http.Client
	logger *log.Logger
}

//go:generate mockery -name=FetchRequestRepository -output=automock -outpkg=automock -case=underscore
type FetchRequestRepository interface {
	Update(ctx context.Context, item *model.FetchRequest) error
}

func NewService(repo FetchRequestRepository, client http.Client, logger *log.Logger) *service {
	return &service{
		repo:   repo,
		client: client,
		logger: logger,
	}
}

func (s *service) FetchAPISpec(ctx context.Context, fr *model.FetchRequest) *string {

	if fr.Mode != model.FetchModeSingle {
		return nil
	}

	resp, err := s.client.Get(fr.URL)
	if err != nil {
		fr.Status.Condition = model.FetchRequestStatusConditionFailed
		log.Errorf("While fetching API Spec: %s", err.Error())
		return nil
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.logger.Errorf("While closing body: %s", err.Error())
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fr.Status.Condition = model.FetchRequestStatusConditionFailed
		log.Errorf("While reading API Spec: %s", err.Error())
		return nil
	}

	spec := string(body)
	fr.Status.Condition = model.FetchRequestStatusConditionSucceeded
	err = s.repo.Update(ctx, fr)
	if err != nil {
		log.Errorf("While updating Fetch Request status: %s", err.Error())
		return nil
	}
	return &spec

}
