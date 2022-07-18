/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webhook_client_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"io/ioutil"
	"net/http"
	"testing"

	accessstrategy2 "github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/stretchr/testify/require"
)

var (
	invalidTemplate   = "invalidTemplate"
	emptyTemplate     = "{}"
	mockedError       = "mocked error"
	mockedLocationURL = "https://test-domain.com/operation"
	webhookAsyncMode  = model.WebhookModeAsync
)

func TestClient_Do_WhenUrlTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &invalidTemplate,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook URL")
}

func TestClient_Do_WhenUrlTemplateIsNil_ShouldReturnError(t *testing.T) {
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    nil,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing webhook url")
}

func TestClient_Do_WhenParseInputTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	invalidInputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"group\": \"{{.Application.Group}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &invalidInputTemplate,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook input body")
}

func TestClient_Do_WhenHeadersTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &invalidTemplate,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook headers")
}

func TestClient_Do_WhenCreatingRequestFails_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(nil, webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "nil Context")
}

func TestClient_Do_WhenAuthFlowCannotBeDetermined_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &emptyTemplate,
			Auth:           &model.Auth{AccessStrategy: str.Ptr("wrong")},
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "could not determine auth flow")
}

func TestClient_Do_WhenExecutingRequestFails_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &emptyTemplate,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{err: errors.New(mockedError)},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)

}

func TestClient_Do_WhenWebhookResponseDoesNotContainLocationURL_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				StatusCode: http.StatusAccepted,
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing location url after executing async webhook")
}

func TestClient_Do_WhenWebhookResponseBodyContainsError_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("{\"error\": \"%s\"}", mockedError)))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)
	require.Contains(t, err.Error(), "received error while polling external system")
}

func TestClient_Do_WhenWebhookResponseBodyContainsErrorWithJSONObjects_ShouldParseErrorSuccessfully(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	mockedJSONObjectError := "{\"code\":\"401\",\"message\":\"Unauthorized\",\"correlationId\":\"12345678-e89b-12d3-a456-556642440000\"}"

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("{\"error\": %s}", mockedJSONObjectError)))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Unauthorized")
	require.Contains(t, err.Error(), "received error while polling external system")
}

func TestClient_Do_WhenWebhookResponseStatusCodeIsGoneAndGoneStatusISDefined_ShouldReturnWebhookStatusGoneError(t *testing.T) {
	goneCodeString := "404"
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := fmt.Sprintf("{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"gone_status_code\": %s,\"error\": \"{{.Body.error}}\"}", goneCodeString)
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusNotFound,
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.IsType(t, webhook_client.WebhookStatusGoneErr{}, err)
	require.Contains(t, err.Error(), goneCodeString)
}

func TestClient_Do_WhenWebhookResponseStatusCodeIsNotSuccess_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusInternalServerError,
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response success status code was not met")
}

func TestClient_Do_WhenSuccessfulBasicAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	username, password := "user", "pass"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
			Auth: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: username,
						Password: password,
					},
				},
			},
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
			roundTripExpectations: func(r *http.Request) {
				credentials, err := auth.LoadFromContext(r.Context())
				require.NoError(t, err)
				basicCreds, ok := credentials.(*auth.BasicCredentials)
				require.True(t, ok)
				require.Equal(t, username, basicCreds.Username)
				require.Equal(t, password, basicCreds.Password)
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Do_WhenSuccessfulOAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	clientID, clientSecret, tokenURL := "client-id", "client-secret", "https://test-domain.com/oauth/token"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
			Auth: &model.Auth{
				Credential: model.CredentialData{
					Oauth: &model.OAuthCredentialData{
						ClientID:     clientID,
						ClientSecret: clientSecret,
						URL:          tokenURL,
					},
				},
			},
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
			roundTripExpectations: func(r *http.Request) {
				credentials, err := auth.LoadFromContext(r.Context())
				require.NoError(t, err)
				oAuthCredentials, ok := credentials.(*auth.OAuthCredentials)
				require.True(t, ok)
				require.Equal(t, clientID, oAuthCredentials.ClientID)
				require.Equal(t, clientSecret, oAuthCredentials.ClientSecret)
				require.Equal(t, tokenURL, oAuthCredentials.TokenURL)
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Do_WhenSuccessfulMTLSWebhook_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	mtlsCalled := false
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			URLTemplate:    &URLTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
			Auth: &model.Auth{
				AccessStrategy: str.Ptr(string(accessstrategy2.CMPmTLSAccessStrategy)),
			},
		},
		Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
	}

	mtlsClient := &http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
			roundTripExpectations: func(r *http.Request) {
				mtlsCalled = true
			},
		},
	}

	client := webhook_client.NewClient(nil, mtlsClient)

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
	require.True(t, mtlsCalled)
}

func TestClient_Do_WhenMissingCorrelationID_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	correlationIDKey := "X-Correlation-Id"
	correlationID := "abc"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: model.Webhook{
			CorrelationIDKey: &correlationIDKey,
			URLTemplate:      &URLTemplate,
			InputTemplate:    &inputTemplate,
			HeaderTemplate:   &headersTemplate,
			OutputTemplate:   &outputTemplate,
			Mode:             &webhookAsyncMode,
		},
		Object:        &webhook.ApplicationLifecycleWebhookRequestObject{Application: app, Headers: map[string]string{}},
		CorrelationID: correlationID,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusAccepted,
			},
			roundTripExpectations: func(r *http.Request) {
				headers := correlation.HeadersFromContext(r.Context())
				correlationIDAttached := false
				for headerKey, headerValue := range headers {
					if headerKey == correlationIDKey && headerValue == correlationID {
						correlationIDAttached = true
						break
					}
				}
				require.True(t, correlationIDAttached)
			},
		},
	}, nil)

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Poll_WhenHeadersTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &invalidTemplate,
				StatusTemplate: &emptyTemplate,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook headers")
}

func TestClient_Poll_WhenCreatingRequestFails_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &emptyTemplate,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Poll(nil, webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "nil Context")
}

func TestClient_Poll_WhenAuthFlowCannotBeDetermined_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{

		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &emptyTemplate,
				Auth:           &model.Auth{AccessStrategy: str.Ptr("wrong")},
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(http.DefaultClient, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "could not determine auth flow")
}

func TestClient_Poll_WhenExecutingRequestFails_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &emptyTemplate,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{err: errors.New(mockedError)},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)

}

func TestClient_Poll_WhenParseStatusTemplateFails_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("{}")))},
		},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing Status Template success status code field")
}

func TestClient_Poll_WhenWebhookResponseBodyContainsError_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("{\"error\": \"%s\"}", mockedError)))),
				StatusCode: http.StatusOK,
			},
		},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)
	require.Contains(t, err.Error(), "received error while polling external system")
}

func TestClient_Poll_WhenWebhookResponseStatusCodeIsNotSuccess_ShouldReturnError(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				StatusCode: http.StatusInternalServerError,
			},
		},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response success status code was not met")
}

func TestClient_Poll_WhenSuccessfulBasicAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	username, password := "user", "pass"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
				Auth: &model.Auth{
					Credential: model.CredentialData{
						Basic: &model.BasicCredentialData{
							Username: username,
							Password: password,
						},
					},
				},
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				StatusCode: http.StatusOK,
			},
			roundTripExpectations: func(r *http.Request) {
				credentials, err := auth.LoadFromContext(r.Context())
				require.NoError(t, err)
				basicCreds, ok := credentials.(*auth.BasicCredentials)
				require.True(t, ok)
				require.Equal(t, username, basicCreds.Username)
				require.Equal(t, password, basicCreds.Password)
			},
		},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Poll_WhenSuccessfulOAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	clientID, clientSecret, tokenURL := "client-id", "client-secret", "https://test-domain.com/oauth/token"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
				Auth: &model.Auth{
					Credential: model.CredentialData{
						Oauth: &model.OAuthCredentialData{
							ClientID:     "client-id",
							ClientSecret: "client-secret",
							URL:          "https://test-domain.com/oauth/token",
						},
					},
				},
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				StatusCode: http.StatusOK,
			},
			roundTripExpectations: func(r *http.Request) {
				credentials, err := auth.LoadFromContext(r.Context())
				require.NoError(t, err)
				oAuthCredentials, ok := credentials.(*auth.OAuthCredentials)
				require.True(t, ok)
				require.Equal(t, clientID, oAuthCredentials.ClientID)
				require.Equal(t, clientSecret, oAuthCredentials.ClientSecret)
				require.Equal(t, tokenURL, oAuthCredentials.TokenURL)
			},
		},
	}, nil)
	_, err := client.Poll(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Poll_WhenSuccessfulMTLSWebhook_ShouldBeSuccessful(t *testing.T) {
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	mtlsCalled := false

	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	pollRequest := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				OutputTemplate: &outputTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
				Auth: &model.Auth{
					AccessStrategy: str.Ptr(string(accessstrategy2.CMPmTLSAccessStrategy)),
				},
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: "https://test-domain.com/poll/",
	}

	mtlsClient := &http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				Header:     http.Header{"Location": []string{mockedLocationURL}},
				StatusCode: http.StatusOK,
			},
			roundTripExpectations: func(r *http.Request) {
				mtlsCalled = true
			},
		},
	}

	client := webhook_client.NewClient(nil, mtlsClient)

	_, err := client.Poll(context.Background(), pollRequest)

	require.NoError(t, err)
	require.True(t, mtlsCalled)
}

func TestClient_Poll_WhenSuccessfulWebhookPollResponseContainsNullErrorField_ShouldBeSuccessful(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"error\":null}"))),
				StatusCode: http.StatusOK,
			},
		},
	}, nil)
	_, err := client.Poll(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Poll_WhenSuccessfulWebhookPollResponseContainsEmptyErrorField_ShouldBeSuccessful(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				HeaderTemplate: &headersTemplate,
				StatusTemplate: &statusTemplate,
				Mode:           &webhookAsyncMode,
			},
			Object: &webhook.ApplicationLifecycleWebhookRequestObject{Application: app},
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"error\":\"\"}"))),
				StatusCode: http.StatusOK,
			},
		},
	}, nil)
	_, err := client.Poll(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Poll_WhenMissingCorrelationID_ShouldBeSuccessful(t *testing.T) {
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	statusTemplate := "{\"status\":\"{{.Body.status}}\",\"success_status_code\": 200,\"success_status_identifier\":\"SUCCEEDED\",\"in_progress_status_identifier\":\"IN_PROGRESS\",\"failed_status_identifier\":\"FAILED\",\"error\": \"{{.Body.error}}\"}"
	correlationIDKey := "X-Correlation-Id"
	correlationID := "abc"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.PollRequest{
		Request: &webhook.Request{
			Webhook: model.Webhook{
				CorrelationIDKey: &correlationIDKey,
				HeaderTemplate:   &headersTemplate,
				StatusTemplate:   &statusTemplate,
				Mode:             &webhookAsyncMode,
			},
			Object:        &webhook.ApplicationLifecycleWebhookRequestObject{Application: app, Headers: map[string]string{}},
			CorrelationID: correlationID,
		},
		PollURL: mockedLocationURL,
	}

	client := webhook_client.NewClient(&http.Client{
		Transport: mockedTransport{
			resp: &http.Response{
				Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
				StatusCode: http.StatusOK,
			},
			roundTripExpectations: func(r *http.Request) {
				headers := correlation.HeadersFromContext(r.Context())
				correlationIDAttached := false
				for headerKey, headerValue := range headers {
					if headerKey == correlationIDKey && headerValue == correlationID {
						correlationIDAttached = true
						break
					}
				}
				require.True(t, correlationIDAttached)
			},
		},
	}, nil)

	_, err := client.Poll(context.Background(), webhookReq)

	require.NoError(t, err)
}

type mockedTransport struct {
	resp                  *http.Response
	err                   error
	roundTripExpectations func(r *http.Request)
}

func (m mockedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if m.roundTripExpectations != nil {
		m.roundTripExpectations(r)
	}
	return m.resp, m.err
}
