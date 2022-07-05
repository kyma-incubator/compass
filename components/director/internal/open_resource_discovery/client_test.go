package ord_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/mock"

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

var successfulRoundTripFunc = func(t *testing.T, bothBaseURLsProvided, noBaseURLsButDocWithAbsURL bool) func(req *http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		var data []byte
		var err error
		var reqBody []byte
		if req.Body != nil {
			reqBody, err = ioutil.ReadAll(req.Body)
			require.NoError(t, err)
		}
		statusCode := http.StatusOK
		if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) && bothBaseURLsProvided {
			config := fixWellKnownConfig()
			config.BaseURL = baseURL2
			data, err = json.Marshal(config)
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), ord.WellKnownEndpoint) {
			data, err = json.Marshal(fixWellKnownConfig())
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), customWebhookConfigURL) && noBaseURLsButDocWithAbsURL {
			config := fixWellKnownConfig()
			config.BaseURL = ""
			config.OpenResourceDiscoveryV1.Documents[0].URL = absoluteDocURL
			data, err = json.Marshal(config)
			require.NoError(t, err)
		} else if strings.Contains(req.URL.String(), customWebhookConfigURL) {
			config := fixWellKnownConfig()
			// WHEN webhookURL is not /well-known, a valid baseURL in the config must be provided
			config.BaseURL = baseURL2
			data, err = json.Marshal(config)
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
		ExpectedBaseURL      string
		ExpectedErr          error
		WebhookURL           string
	}{
		{
			Name:          "Success when webhookURL contains /well-known suffix",
			RoundTripFunc: successfulRoundTripFunc(t, false, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL,
		},
		{
			Name:          "Success when webhookURL is custom and does not contain /well-known suffix but config baseURL is provided",
			RoundTripFunc: successfulRoundTripFunc(t, false, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL2,
			WebhookURL:      customWebhookConfigURL,
		},
		{
			Name:          "Success when webhookURL is /well-known and config baseURL is set - config baseURL is chosen",
			RoundTripFunc: successfulRoundTripFunc(t, true, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL2,
		},
		{
			Name:          "Success when webhookURL isn't /well-known, no config baseURL is set but all documents in the config are with absolute url",
			RoundTripFunc: successfulRoundTripFunc(t, false, true),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: "",
			WebhookURL:      customWebhookConfigURL,
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
			RoundTripFunc: successfulRoundTripFunc(t, false, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL,
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
			RoundTripFunc: successfulRoundTripFunc(t, false, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL,
		},
		{
			Name: "Well-known config success fetch with access strategy",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				data, err := json.Marshal(fixWellKnownConfig())
				require.NoError(t, err)

				executor := &automock.Executor{}
				executor.On("Execute", context.TODO(), mock.Anything, baseURL+ord.WellKnownEndpoint, "").Return(&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBuffer(data)),
				}, nil).Once()

				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				executorProvider.On("Provide", accessstrategy.OpenAccessStrategy).Return(accessstrategy.NewOpenAccessStrategyExecutor(), nil).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t, false, false),
			ExpectedResult: ord.Documents{
				fixORDDocument(),
			},
			ExpectedBaseURL: baseURL,
		},
		{
			Name: "Well-known config fetch with access strategy fails when access strategy provider returns error",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(nil, testErr).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t, false, false),
			ExpectedErr:    errors.Errorf("cannot find executor for access strategy %q as part of webhook processing: test", testAccessStrategy),
		},
		{
			Name: "Well-known config fetch with access strategy fails when access strategy executor returns error",
			ExecutorProviderFunc: func() accessstrategy.ExecutorProvider {
				executor := &automock.Executor{}
				executor.On("Execute", context.TODO(), mock.Anything, baseURL+ord.WellKnownEndpoint, "").Return(nil, testErr).Once()

				executorProvider := &automock.ExecutorProvider{}
				executorProvider.On("Provide", accessstrategy.Type(testAccessStrategy)).Return(executor, nil).Once()
				return executorProvider
			},
			AccessStrategy: testAccessStrategy,
			RoundTripFunc:  successfulRoundTripFunc(t, false, false),
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
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].Type = accessstrategy.CustomAccessStrategy
					config.OpenResourceDiscoveryV1.Documents[0].AccessStrategies[0].CustomType = "foo.bar:test:v1"
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
			ExpectedResult:  ord.Documents{},
			ExpectedBaseURL: baseURL,
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

			certCache := certloader.NewCertificateCache()
			var executorProviderMock accessstrategy.ExecutorProvider = accessstrategy.NewDefaultExecutorProvider(certCache)
			if test.ExecutorProviderFunc != nil {
				executorProviderMock = test.ExecutorProviderFunc()
			}

			client := ord.NewClient(testHTTPClient, executorProviderMock)

			testApp := fixApplicationPage().Data[0]
			testWebhook := fixWebhooks()[0]

			testWebhook.Auth = test.Credentials

			if test.WebhookURL != "" {
				testWebhook.URL = str.Ptr(test.WebhookURL)
			} else {
				testWebhook.URL = str.Ptr(baseURL + ord.WellKnownEndpoint)
			}

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
				require.Equal(t, test.ExpectedBaseURL, actualBaseURL)
				require.Equal(t, test.ExpectedResult, docs)
			}

			if test.ExecutorProviderFunc != nil {
				mock.AssertExpectationsForObjects(t, executorProviderMock)
			}
		})
	}
}
