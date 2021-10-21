package ord_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/accessstrategy/automock"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const testAccessStrategy = "accessStrategy"

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

var successfulRoundTripFunc = func(t *testing.T) func(req *http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		var data []byte
		var err error
		var reqBody []byte
		if req.Body != nil {
			reqBody, err = ioutil.ReadAll(req.Body)
			require.NoError(t, err)
		}
		statusCode := http.StatusOK
		if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
			data, err = json.Marshal(fixWellKnownConfig())
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), ordDocURI) {
			data, err = json.Marshal(fixORDDocument())
			require.NoError(t, err)
		} else if strings.Contains(string(reqBody), "client_secret") {
			statusCode = http.StatusOK
			data, err = json.Marshal(struct {
				AccessToken string `json:"access_token"`
			}{
				AccessToken: "test-tkn",
			})
			require.NoError(t, err)
		} else if strings.Contains(string(reqBody), "grant_type=client_credentials") {
			statusCode = http.StatusOK
		} else {
			statusCode = http.StatusNotFound
		}
		return &http.Response{
			StatusCode: statusCode,
			Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
		}
	}
}

func TestClient_FetchOpenResourceDiscoveryDocuments(t *testing.T) {
	testErr := errors.New("test")

	testCases := []struct {
		Name                 string
		Credentials          *model.Auth
		AccessStrategy       string
		RoundTripFunc        func(req *http.Request) *http.Response
		ExecutorProviderFunc func() accessstrategy.ExecutorProvider
		ExpectedResult       ord.Documents
		ExpectedErr          error
	}{
		{
			Name:          "Success",
			RoundTripFunc: successfulRoundTripFunc(t),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
		},
		{
			Name: "Success with secured system type configured and basic credentials",
			Credentials: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "user",
						Password: "pass",
					},
				},
			},
			RoundTripFunc: successfulRoundTripFunc(t),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
		},
		{
			Name: "Success with secured system type configured and oauth credentials",
			Credentials: &model.Auth{
				Credential: model.CredentialData{
					Oauth: &model.OAuthCredentialData{
						ClientID:     "client-id",
						ClientSecret: "client-secret",
						URL:          "url",
					},
				},
			},
			RoundTripFunc: successfulRoundTripFunc(t),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
		},
		{
			Name: "Well-known config success fetch with access strategy",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				data, err := json.Marshal(fixWellKnownConfig())
				require.NoError(t, err)

				executor := &automock.Executor{}
				executor.On("Execute", mock.Anything, mock.Anything, baseURL+ord.WellKnownEndpoint).Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}, nil).Once()

				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				executorProvider.On("Provide", accessstrategy.OpenAccessStrategy).Return(accessstrategy.NewOpenAccessStrategyExecutor(), nil).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
		},
		{
			Name: "Well-known config fetch with access strategy fails when access strategy provider returns error",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(nil, testErr).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t),
			ExpectedErr:    errors.Errorf("cannot find executor for access strategy %q as part of webhook processing: test", testAccessStrategy),
		},
		{
			Name: "Well-known config fetch with access strategy fails when access strategy executor returns error",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executor := &automock.Executor{}
				executor.On("Execute", mock.Anything, mock.Anything, baseURL+ord.WellKnownEndpoint).Return(nil, testErr).Once()

				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t),
			ExpectedErr:    errors.Errorf("error while fetching open resource discovery well-known configuration with access strategy %q: test", testAccessStrategy),
		},
		{
			Name:        "Error fetching well-known config due to missing basic credentials",
			ExpectedErr: errors.New("error while fetching open resource discovery well-known configuration with webhook credentials: Invalid data [reason=Credentials not provided]"),
			Credentials: &model.Auth{
				Credential: model.CredentialData{
					Basic: nil,
				},
			},
		},
		{
			Name:        "Error fetching well-known config due to missing oauth credentials",
			ExpectedErr: errors.New("error while fetching open resource discovery well-known configuration with webhook credentials: Invalid data [reason=Credentials not provided]"),
			Credentials: &model.Auth{
				Credential: model.CredentialData{
					Oauth: nil,
				},
			},
		},
		{
			Name: "Error fetching well-known config due to invalid credentials",
			Credentials: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "user",
						Password: "pass",
					},
				},
			},
			RoundTripFunc: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       nil,
				}
			},
			ExpectedErr: errors.New("error while fetching open resource discovery well-known configuration: status code 401"),
		},
		{
			Name: "Error fetching well-known config",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				statusCode := http.StatusNotFound
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					statusCode = http.StatusInternalServerError
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedErr: errors.New("error while fetching open resource discovery well-known configuration: status code 500"),
		},
		{
			Name: "Error when well-known config is not proper json",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				statusCode := http.StatusNotFound
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					statusCode = http.StatusOK
					data = []byte("test")
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedErr: errors.New("error unmarshaling json body"),
		},
		{
			Name: "Document with unsupported access strategy is skipped",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					config := fixWellKnownConfig()
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = "custom"
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = "test"
					data, err = json.Marshal(config)
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					require.FailNow(t, "document with unsupported access strategy should not be fetched")
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
		},
		{
			Name: "Error fetching document",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					data, err = json.Marshal(fixWellKnownConfig())
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					statusCode = http.StatusInternalServerError
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
			ExpectedErr:    errors.Errorf("error while fetching open resource discovery document %q: status code %d", baseURL+ordDocURI, 500),
		},
		{
			Name: "Error when document is not proper json",
			RoundTripFunc: func(req *http.Request) *http.Response {
				var data []byte
				var err error
				statusCode := http.StatusOK
				if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
					data, err = json.Marshal(fixWellKnownConfig())
					require.NoError(t, err)
				} else if strings.Contains(req.URL.String(), ordDocURI) {
					data = []byte("test")
				} else {
					statusCode = http.StatusNotFound
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}
			},
			ExpectedResult: ord.Documents{},
			ExpectedErr:    errors.New("error unmarshaling document"),
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			testHTTPClient := NewTestClient(test.RoundTripFunc)

			var executorProviderMock accessstrategy.ExecutorProvider = accessstrategy.NewDefaultExecutorProvider()
			if test.ExecutorProviderFunc != nil {
				executorProviderMock = test.ExecutorProviderFunc()
			}

			client := ord.NewClient(testHTTPClient, executorProviderMock)

			testApp := fixApplicationPage().Data[0]
			testWebhook := fixWebhooks()[0]

			testWebhook.Auth = test.Credentials

			if len(test.AccessStrategy) > 0 {
				if testWebhook.Auth == nil {
					testWebhook.Auth = &model.Auth{}
				}
				testWebhook.Auth.AccessStrategy = &test.AccessStrategy
			}

			docs, actualBaseURL, err := client.FetchOpenResourceDiscoveryDocuments(context.TODO(), testApp, testWebhook)

			if test.ExpectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Len(t, docs, len(test.ExpectedResult))
				require.Equal(t, baseURL, actualBaseURL)
				require.Equal(t, test.ExpectedResult, docs)
			}

			if test.ExecutorProviderFunc != nil {
				mock.AssertExpectationsForObjects(t, executorProviderMock)
			}
		})
	}
}
