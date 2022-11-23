package fetchrequest_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	accessstrategyautomock "github.com/kyma-incubator/compass/components/director/pkg/accessstrategy/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/stretchr/testify/assert"
)

const (
	externalClientCertSecretName = "resource-name1"
	extSvcClientCertSecretName   = "resource-name2"
)

var testErr = errors.New("test")

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	response := f(req)
	if response.StatusCode == http.StatusBadRequest {
		return nil, errors.New("error")
	}
	return response, nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func TestService_Update(t *testing.T) {
	fetchReq := &model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	testCases := []struct {
		Name                 string
		Context              context.Context
		FetchRequest         *model.FetchRequest
		FetchRequestRepoMock *automock.FetchRequestRepository
		ExpectedError        error
	}{

		{
			Name:         "Success",
			Context:      tenant.SaveToContext(context.TODO(), tenantID, tenantID),
			FetchRequest: fetchReq,
			FetchRequestRepoMock: func() *automock.FetchRequestRepository {
				fetchReqRepoMock := automock.FetchRequestRepository{}
				fetchReqRepoMock.On("Update", mock.Anything, tenantID, fetchReq).Return(nil).Once()
				return &fetchReqRepoMock
			}(),
			ExpectedError: nil,
		},
		{
			Name:                 "Fails when tenant is missing in context",
			Context:              context.TODO(),
			FetchRequest:         fetchReq,
			FetchRequestRepoMock: &automock.FetchRequestRepository{},
			ExpectedError:        apperrors.NewCannotReadTenantError(),
		},
		{
			Name:         "Fails when repo update fails",
			Context:      tenant.SaveToContext(context.TODO(), tenantID, tenantID),
			FetchRequest: fetchReq,
			FetchRequestRepoMock: func() *automock.FetchRequestRepository {
				fetchReqRepoMock := automock.FetchRequestRepository{}
				fetchReqRepoMock.On("Update", mock.Anything, tenantID, fetchReq).Return(testErr).Once()
				return &fetchReqRepoMock
			}(),
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.FetchRequestRepoMock
			svc := fetchrequest.NewService(repo, nil, nil)
			resultErr := svc.Update(testCase.Context, testCase.FetchRequest)
			assert.Equal(t, testCase.ExpectedError, resultErr)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_HandleSpec(t *testing.T) {
	const username = "username"
	const password = "password"
	const clientID = "clId"
	const secret = "clSecret"
	const url = "mocked-url/oauth/token"

	var testAccessStrategy = "testAccessStrategy"

	mockSpec := "spec"
	timestamp := time.Now()

	modelInput := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	modelInputBundle := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeBundle,
	}

	modelInputFilter := model.FetchRequest{
		ID:     "test",
		Mode:   model.FetchModeSingle,
		Filter: str.Ptr("filter"),
	}

	modelInputAccessStrategy := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
		URL:  "http://test.com",
		Auth: &model.Auth{AccessStrategy: &testAccessStrategy},
	}

	modelInputBasicCredentials := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: username,
					Password: password,
				},
			},
		},
		Mode: model.FetchModeSingle,
	}

	modelInputMissingCredentials := model.FetchRequest{
		ID: "test",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: nil,
			},
		},
		Mode: model.FetchModeSingle,
	}

	modelInputOauth := model.FetchRequest{
		ID:  "test",
		URL: "http://dummy.url.sth",
		Auth: &model.Auth{
			Credential: model.CredentialData{
				Basic: nil,
				Oauth: &model.OAuthCredentialData{
					ClientID:     clientID,
					ClientSecret: secret,
					URL:          url,
				},
			},
		},
		Mode: model.FetchModeSingle,
	}

	testCases := []struct {
		Name                 string
		Client               func(t *testing.T) *http.Client
		InputFr              model.FetchRequest
		ExecutorProviderFunc func() accessstrategy.ExecutorProvider
		ExpectedResult       *string
		ExpectedStatus       *model.FetchRequestStatus
	}{

		{
			Name: "Success without authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInput,
			ExpectedResult: &mockSpec,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp),
		},
		{
			Name: "Nil when fetch request validation fails due to mode Bundle",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{}
				})
			},

			InputFr:        modelInputBundle,
			ExpectedResult: nil,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr("Invalid data [reason=Unsupported fetch mode: BUNDLE]"), timestamp),
		},
		{
			Name: "Nil when fetch request validation fails due to provided filter",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{}
				})
			},

			InputFr:        modelInputFilter,
			ExpectedResult: nil,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr("Invalid data [reason=Filter for Fetch Request was provided, currently it's unsupported]"), timestamp),
		},
		{
			Name: "Success with access strategy",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executor := &accessstrategyautomock.Executor{}
				executor.On("Execute", mock.Anything, mock.Anything, modelInputAccessStrategy.URL, "").Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
				}, nil).Once()

				executorProvider := &accessstrategyautomock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				return executorProvider
			},
			Client: func(t *testing.T) *http.Client {
				return nil
			},
			InputFr:        modelInputAccessStrategy,
			ExpectedResult: &mockSpec,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp),
		},
		{
			Name: "Fails when access strategy is unknown",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executorProvider := &accessstrategyautomock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(nil, testErr).Once()
				return executorProvider
			},
			Client: func(t *testing.T) *http.Client {
				return nil
			},
			InputFr:        modelInputAccessStrategy,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec: test"), timestamp),
		},
		{
			Name: "Fails when access strategy execution fail",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executor := &accessstrategyautomock.Executor{}
				executor.On("Execute", mock.Anything, mock.Anything, modelInputAccessStrategy.URL, "").Return(nil, testErr).Once()

				executorProvider := &accessstrategyautomock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				return executorProvider
			},
			Client: func(t *testing.T) *http.Client {
				return nil
			},
			InputFr:        modelInputAccessStrategy,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec: test"), timestamp),
		},
		{
			Name: "Success with basic authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					actualUsername, actualPassword, ok := req.BasicAuth()
					assert.True(t, ok)
					assert.Equal(t, username, actualUsername)
					assert.Equal(t, password, actualPassword)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInputBasicCredentials,
			ExpectedResult: &mockSpec,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp),
		},
		{
			Name: "Fails to execute the request with basic authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					actualUsername, actualPassword, ok := req.BasicAuth()
					assert.True(t, ok)
					assert.Equal(t, username, actualUsername)
					assert.Equal(t, password, actualPassword)
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}
				})
			},
			InputFr:        modelInputBasicCredentials,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec status code: 500"), timestamp),
		},
		{
			Name: "Nil when auth without credentials is provided",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{}
				})
			},

			InputFr:        modelInputMissingCredentials,
			ExpectedResult: nil,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec: Invalid data [reason=Credentials not provided]"), timestamp),
		},
		{
			Name: "Success with oauth authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if req.URL.String() == url {
						actualClientID, actualSecret, ok := req.BasicAuth()
						assert.True(t, ok)
						assert.Equal(t, clientID, actualClientID)
						assert.Equal(t, secret, actualSecret)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"token"}`)),
						}
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInputOauth,
			ExpectedResult: &mockSpec,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp),
		},
		{
			Name: "Fails to fetch oauth token with oauth authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					actualClientID, actualSecret, ok := req.BasicAuth()
					if ok {
						assert.Equal(t, clientID, actualClientID)
						assert.Equal(t, secret, actualSecret)
					} else {
						credentials, err := io.ReadAll(req.Body)
						assert.NoError(t, err)
						assert.Contains(t, string(credentials), fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, secret))
					}
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}
				})
			},
			InputFr:        modelInputOauth,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec: Get \"http://dummy.url.sth\": oauth2: cannot fetch token: \nResponse: "), timestamp),
		},
		{
			Name: "Fails to execute the request with oauth authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if req.URL.String() == url {
						actualClientID, actualSecret, ok := req.BasicAuth()
						assert.True(t, ok)
						assert.Equal(t, clientID, actualClientID)
						assert.Equal(t, secret, actualSecret)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(bytes.NewBufferString(`{"access_token":"token"}`)),
						}
					}

					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInputOauth,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching Spec status code: 500"), timestamp),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			certCache := certloader.NewCertificateCache()
			var executorProviderMock accessstrategy.ExecutorProvider = accessstrategy.NewDefaultExecutorProvider(certCache, externalClientCertSecretName, extSvcClientCertSecretName)
			if testCase.ExecutorProviderFunc != nil {
				executorProviderMock = testCase.ExecutorProviderFunc()
			}

			ctx := context.TODO()
			ctx = tenant.SaveToContext(ctx, tenantID, tenantID)

			frRepo := &automock.FetchRequestRepository{}
			frRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()

			svc := fetchrequest.NewService(frRepo, testCase.Client(t), executorProviderMock)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			result := svc.HandleSpec(ctx, &testCase.InputFr)

			assert.Equal(t, testCase.ExpectedStatus, testCase.InputFr.Status)
			assert.Equal(t, testCase.ExpectedResult, result)

			if testCase.ExecutorProviderFunc != nil {
				mock.AssertExpectationsForObjects(t, executorProviderMock)
			}
		})
	}
}

func TestService_HandleSpec_FailedToUpdateStatusAfterFetching(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, tenantID)

	timestamp := time.Now()
	frRepo := &automock.FetchRequestRepository{}
	frRepo.On("Update", ctx, tenantID, mock.Anything).Return(errors.New("error")).Once()

	certCache := certloader.NewCertificateCache()
	svc := fetchrequest.NewService(frRepo, NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString("spec")),
		}
	}), accessstrategy.NewDefaultExecutorProvider(certCache, externalClientCertSecretName, extSvcClientCertSecretName))
	svc.SetTimestampGen(func() time.Time { return timestamp })

	modelInput := &model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	result := svc.HandleSpec(ctx, modelInput)
	expectedStatus := fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp)

	assert.Equal(t, expectedStatus, modelInput.Status)
	assert.Nil(t, result)
}

func TestService_HandleSpec_SucceedsAfterRetryMechanismIsLeveraged(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, tenantID)

	timestamp := time.Now()
	frRepo := &automock.FetchRequestRepository{}
	frRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()

	certCache := certloader.NewCertificateCache()
	retryConfig := &retry.Config{
		Attempts: 3,
		Delay:    100 * time.Millisecond,
	}

	mockSpec := "spec"

	invocations := 0
	svc := fetchrequest.NewServiceWithRetry(frRepo, NewTestClient(func(req *http.Request) *http.Response {
		defer func() {
			invocations++
		}()

		if invocations != int(retryConfig.Attempts)-1 {
			return &http.Response{StatusCode: http.StatusInternalServerError}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(mockSpec)),
		}
	}), accessstrategy.NewDefaultExecutorProvider(certCache, externalClientCertSecretName, extSvcClientCertSecretName), retry.NewHTTPExecutor(retryConfig))
	svc.SetTimestampGen(func() time.Time { return timestamp })

	modelInput := &model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	result := svc.HandleSpec(ctx, modelInput)
	expectedStatus := fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp)

	assert.Equal(t, expectedStatus, modelInput.Status)
	assert.Equal(t, mockSpec, *result)
	assert.Equal(t, int(retryConfig.Attempts), invocations)
}

func TestService_HandleSpec_FailsAfterRetryMechanismIsExhausted(t *testing.T) {
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, tenantID)

	timestamp := time.Now()
	frRepo := &automock.FetchRequestRepository{}
	frRepo.On("Update", ctx, tenantID, mock.Anything).Return(nil).Once()

	certCache := certloader.NewCertificateCache()
	retryConfig := &retry.Config{
		Attempts: 3,
		Delay:    100 * time.Millisecond,
	}

	invocations := 0
	svc := fetchrequest.NewServiceWithRetry(frRepo, NewTestClient(func(req *http.Request) *http.Response {
		defer func() {
			invocations++
		}()

		return &http.Response{StatusCode: http.StatusInternalServerError}
	}), accessstrategy.NewDefaultExecutorProvider(certCache, externalClientCertSecretName, extSvcClientCertSecretName), retry.NewHTTPExecutor(retryConfig))
	svc.SetTimestampGen(func() time.Time { return timestamp })

	modelInput := &model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	result := svc.HandleSpec(ctx, modelInput)
	respStatusCodeErr := fmt.Sprintf("unexpected status code: %d", http.StatusInternalServerError)
	expectedErr := fmt.Sprintf("All attempts fail:\n#1: %s\n#2: %s\n#3: %s", respStatusCodeErr, respStatusCodeErr, respStatusCodeErr)
	expectedStatus := fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr(fmt.Sprintf("While fetching Spec: %s", expectedErr)), timestamp)

	assert.Equal(t, expectedStatus, modelInput.Status)
	assert.Nil(t, result)
	assert.Equal(t, int(retryConfig.Attempts), invocations)
}
