package fetchrequest

import (
	"context"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	client http.Client
	logger *log.Logger
}

func NewService(client http.Client, logger *log.Logger) *service {
	return &service{
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
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fr.Status.Condition = model.FetchRequestStatusConditionFailed
		log.Errorf("While reading API Spec: %s", err.Error())
		return nil
	}

	spec := string(body)
	fr.Status.Condition = model.FetchRequestStatusConditionSucceeded
	return &spec

}
