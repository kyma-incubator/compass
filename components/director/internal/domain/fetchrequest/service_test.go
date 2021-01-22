package fetchrequest_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

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

func TestService_HandleSpec(t *testing.T) {
	const username = "username"
	const password = "password"
	const clientId = "clId"
	const secret = "clSecret"
	const url = "mocked-url/oauth/token"
	mockSpec := "spec"
	timestamp := time.Now()

	modelInput := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModeSingle,
	}

	modelInputPackage := model.FetchRequest{
		ID:   "test",
		Mode: model.FetchModePackage,
	}

	modelInputFilter := model.FetchRequest{
		ID:     "test",
		Mode:   model.FetchModeSingle,
		Filter: str.Ptr("filter"),
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
					ClientID:     clientId,
					ClientSecret: secret,
					URL:          url,
				},
			},
		},
		Mode: model.FetchModeSingle,
	}

	testCases := []struct {
		Name           string
		Client         func(t *testing.T) *http.Client
		InputFr        model.FetchRequest
		ExpectedResult *string
		ExpectedStatus *model.FetchRequestStatus
	}{

		{
			Name: "Success without authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInput,
			ExpectedResult: &mockSpec,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionSucceeded, nil, timestamp),
		},
		{
			Name: "Nil when fetch request validation fails due to mode Package",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					return &http.Response{}
				})
			},

			InputFr:        modelInputPackage,
			ExpectedResult: nil,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionInitial, str.Ptr("Invalid data [reason=Unsupported fetch mode: PACKAGE]"), timestamp),
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
			Name: "Success with basic authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					actualUsername, actualPassword, ok := req.BasicAuth()
					assert.True(t, ok)
					assert.Equal(t, username, actualUsername)
					assert.Equal(t, password, actualPassword)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
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
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching API Spec status code: 500"), timestamp),
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
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching API Spec: Invalid data [reason=Credentials not provided]"), timestamp),
		},
		{
			Name: "Success with oauth authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if req.URL.String() == url {
						actualClientId, actualSecret, ok := req.BasicAuth()
						assert.True(t, ok)
						assert.Equal(t, clientId, actualClientId)
						assert.Equal(t, secret, actualSecret)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`{"access_token":"token"}`)),
						}
					}

					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
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
					actualClientId, actualSecret, ok := req.BasicAuth()
					if ok {
						assert.Equal(t, clientId, actualClientId)
						assert.Equal(t, secret, actualSecret)
					} else {
						credentials, err := ioutil.ReadAll(req.Body)
						assert.NoError(t, err)
						assert.Contains(t, string(credentials), fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials", clientId, secret))
					}
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
					}
				})
			},
			InputFr:        modelInputOauth,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching API Spec: Get \"http://dummy.url.sth\": oauth2: cannot fetch token: \nResponse: "), timestamp),
		},
		{
			Name: "Fails to execute the request with oauth authentication",
			Client: func(t *testing.T) *http.Client {
				return NewTestClient(func(req *http.Request) *http.Response {
					if req.URL.String() == url {
						actualClientId, actualSecret, ok := req.BasicAuth()
						assert.True(t, ok)
						assert.Equal(t, clientId, actualClientId)
						assert.Equal(t, secret, actualSecret)
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       ioutil.NopCloser(bytes.NewBufferString(`{"access_token":"token"}`)),
						}
					}

					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(bytes.NewBufferString(mockSpec)),
					}
				})
			},
			InputFr:        modelInputOauth,
			ExpectedStatus: fetchrequest.FixStatus(model.FetchRequestStatusConditionFailed, str.Ptr("While fetching API Spec status code: 500"), timestamp),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			frRepo := &automock.FetchRequestRepository{}
			frRepo.On("Update", ctx, mock.Anything).Return(nil).Once()

			svc := fetchrequest.NewService(frRepo, testCase.Client(t))
			svc.SetTimestampGen(func() time.Time { return timestamp })

			result := svc.HandleSpec(ctx, &testCase.InputFr)

			assert.Equal(t, testCase.ExpectedStatus, testCase.InputFr.Status)
			assert.Equal(t, testCase.ExpectedResult, result)

		})
	}
}

func TestService_HandleSpec_FailedToUpdateStatusAfterFetching(t *testing.T) {
	ctx := context.TODO()
	timestamp := time.Now()
	frRepo := &automock.FetchRequestRepository{}
	frRepo.On("Update", ctx, mock.Anything).Return(errors.New("error")).Once()

	svc := fetchrequest.NewService(frRepo, NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBufferString("spec")),
		}
	}))
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
