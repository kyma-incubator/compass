package fetchrequest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"

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
	retryHTTPFuncExecutor          *retry.HTTPExecutor
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

// NewServiceWithRetry creates a FetchRequest service which is able to retry failed HTTP requests
func NewServiceWithRetry(repo FetchRequestRepository, client *http.Client, executorProvider accessstrategy.ExecutorProvider, retryHTTPExecutor *retry.HTTPExecutor) *service {
	return &service{
		repo:                           repo,
		client:                         client,
		timestampGen:                   timestamp.DefaultGenerator,
		accessStrategyExecutorProvider: executorProvider,
		retryHTTPFuncExecutor:          retryHTTPExecutor,
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
	data, fr.Status = s.FetchSpec(ctx, fr)

	if err := s.repo.Update(ctx, tnt, fr); err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while updating fetch request status: %v", err)
		return nil
	}

	return data
}

// Update is identical to HandleSpec with the difference that the fetch request is only updated in DB without being re-executed
func (s *service) Update(ctx context.Context, fr *model.FetchRequest) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while getting tenant: %v", err)
		return err
	}

	if err := s.repo.Update(ctx, tnt, fr); err != nil {
		log.C(ctx).WithError(err).Errorf("An error has occurred while updating fetch request: %v", err)
		return err
	}

	return nil
}

func (s *service) FetchSpec(ctx context.Context, fr *model.FetchRequest) (*string, *model.FetchRequestStatus) {
	err := s.validateFetchRequest(fr)
	if err != nil {
		log.C(ctx).WithError(err).Error()
		return nil, FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr(err.Error()), s.timestampGen())
	}

	localTenantID, err := tenant.LoadLocalTenantIDFromContext(ctx)
	if err != nil {
		return nil, FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr(err.Error()), s.timestampGen())
	}
	var doRequest retry.ExecutableHTTPFunc
	if fr.Auth != nil && fr.Auth.AccessStrategy != nil && len(*fr.Auth.AccessStrategy) > 0 {
		log.C(ctx).Infof("Fetch Request with id %s is configured with %s access strategy.", fr.ID, *fr.Auth.AccessStrategy)
		var executor accessstrategy.Executor
		executor, err = s.accessStrategyExecutorProvider.Provide(accessstrategy.Type(*fr.Auth.AccessStrategy))
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Cannot find executor for access strategy %q as part of fetch request %s processing: %v", *fr.Auth.AccessStrategy, fr.ID, err)
			return nil, FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec: %s", err.Error())), s.timestampGen())
		}

		doRequest = func() (*http.Response, error) {
			return executor.Execute(ctx, s.client, fr.URL, localTenantID)
		}
	} else if fr.Auth != nil {
		doRequest = func() (*http.Response, error) {
			return httputil.GetRequestWithCredentials(ctx, s.client, fr.URL, localTenantID, fr.Auth)
		}
	} else {
		doRequest = func() (*http.Response, error) {
			return httputil.GetRequestWithoutCredentials(s.client, fr.URL, localTenantID)
		}
	}

	var resp *http.Response
	if s.retryHTTPFuncExecutor != nil {
		resp, err = s.retryHTTPFuncExecutor.Execute(doRequest)
	} else {
		resp, err = doRequest()
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

	body, err := io.ReadAll(resp.Body)
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
