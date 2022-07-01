package fetchrequest

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type service struct {
	repo                           FetchRequestRepository
	client                         *http.Client
	timestampGen                   timestamp.Generator
	accessStrategyExecutorProvider accessstrategy.ExecutorProvider
}

// FetchRequestRepository missing godoc
//go:generate mockery --name=FetchRequestRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FetchRequestRepository interface {
	Update(ctx context.Context, tenant string, item *model.FetchRequest) error
}

// NewService missing godoc
func NewService(repo FetchRequestRepository, client *http.Client, executorProvider accessstrategy.ExecutorProvider) *service {
	return &service{
		repo:                           repo,
		client:                         client,
		timestampGen:                   timestamp.DefaultGenerator,
		accessStrategyExecutorProvider: executorProvider,
	}
}

// HandleSpec missing godoc
func (s *service) HandleSpec(ctx context.Context, fr *model.FetchRequest) *string {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while getting tenant: %v", err)
		return nil
	}

	var data *string
	data, fr.Status = s.fetchSpec(ctx, fr)

	if err := s.repo.Update(ctx, tnt, fr); err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while updating fetch request status: %v", err)
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
	if fr.Auth != nil && fr.Auth.AccessStrategy != nil && len(*fr.Auth.AccessStrategy) > 0 {
		log.C(ctx).Infof("Fetch Request with id %s is configured with %s access strategy.", fr.ID, *fr.Auth.AccessStrategy)
		var executor accessstrategy.Executor
		executor, err = s.accessStrategyExecutorProvider.Provide(accessstrategy.Type(*fr.Auth.AccessStrategy))
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Cannot find executor for access strategy %q as part of fetch request %s processing: %v", *fr.Auth.AccessStrategy, fr.ID, err)
			return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec: %s", err.Error())), s.timestampGen())
		}
		resp, err = executor.Execute(ctx, s.client, fr.URL, "")
	} else if fr.Auth != nil {
		resp, err = httputil.GetRequestWithCredentials(ctx, s.client, fr.URL, "", fr.Auth)
	} else {
		resp, err = httputil.GetRequestWithoutCredentials(s.client, fr.URL, "")
	}

	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while fetching Spec: %v", err)
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec: %s", err.Error())), s.timestampGen())
	}

	defer func() {
		if resp.Body != nil {
			err := resp.Body.Close()
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while closing response body: %v", err)
			}
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while reading Spec: %v", err)
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While reading Spec: %s", err.Error())), s.timestampGen())
	}

	if resp.StatusCode != http.StatusOK {
		log.C(ctx).Errorf("Failed to execute fetch request for %s with id %q: status code: %d body: %s", fr.ObjectType, fr.ObjectID, resp.StatusCode, string(body))
		return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec status code: %d", resp.StatusCode)), s.timestampGen())
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

// FixStatus missing godoc
func FixStatus(condition model.FetchRequestStatusCondition, message *string, timestamp time.Time) *model.FetchRequestStatus {
	return &model.FetchRequestStatus{
		Condition: condition,
		Message:   message,
		Timestamp: timestamp,
	}
}
