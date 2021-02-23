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

package webhook_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"

	"github.com/stretchr/testify/require"
)

var (
	invalidTemplate   = "invalidTemplate"
	mockedError       = "mocked error"
	mockedLocationURL = "https://test-domain.com/operation"
	webhookAsyncMode  = graphql.WebhookModeAsync
)

func TestClient_Do_WhenUrlTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate: &invalidTemplate,
		},
		Data: web_hook.RequestData{},
	}

	client := webhook.DefaultClient{}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook URL")
}

func TestClient_Do_WhenUrlTemplateIsNil_ShouldReturnError(t *testing.T) {
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate: nil,
		},
		Data: web_hook.RequestData{},
	}

	client := webhook.DefaultClient{}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing webhook url")
}

func TestClient_Do_WhenParseInputTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	invalidInputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"group\": \"{{.Application.Group}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:   &URLTemplate,
			InputTemplate: &invalidInputTemplate,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse webhook input body")
}

func TestClient_Do_WhenHeadersTemplateIsInvalid_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &invalidTemplate,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{}

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
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{}

	_, err := client.Do(nil, webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "nil Context")
}

func TestClient_Do_WhenExecutingRequestFails_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{err: errors.New(mockedError)},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)

}

func TestClient_Do_WhenParseOutputTemplateFails_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			Mode:           &webhookAsyncMode,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("{}")))},
			},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "missing webhook output template")
}

func TestClient_Do_WhenWebhookResponseDoesNotContainLocationURL_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					StatusCode: http.StatusAccepted,
				},
			},
		},
	}

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
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf("{\"error\": \"%s\"}", mockedError)))),
					Header:     http.Header{"Location": []string{mockedLocationURL}},
					StatusCode: http.StatusAccepted,
				},
			},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), mockedError)
	require.Contains(t, err.Error(), "received error while polling external system")
}

func TestClient_Do_WhenWebhookResponseStatusCodeIsNotSuccess_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					Header:     http.Header{"Location": []string{mockedLocationURL}},
					StatusCode: http.StatusInternalServerError,
				},
			},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.Error(t, err)
	require.Contains(t, err.Error(), "response success status code was not met")
}

func TestClient_Do_WhenSuccessfulBasicAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	username := "user"
	password := "pass"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
			Auth: &graphql.Auth{
				Credential: &graphql.BasicCredentialData{
					Username: username,
					Password: password,
				},
			},
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					Header:     http.Header{"Location": []string{mockedLocationURL}},
					StatusCode: http.StatusAccepted,
				},
				roundTripExpectations: func(r *http.Request) {
					auth := username + ":" + password
					base64Creds := base64.StdEncoding.EncodeToString([]byte(auth))
					require.NotEmpty(t, r.Header["Authorization"])
					require.Equal(t, r.Header["Authorization"][0], "Basic "+base64Creds)
				},
			},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
}

func TestClient_Do_WhenSuccessfulOAuthWebhook_ShouldBeSuccessful(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
	app := &graphql.Application{BaseEntity: &graphql.BaseEntity{ID: "appID"}}
	webhookReq := &webhook.Request{
		Webhook: graphql.Webhook{
			URLTemplate:    &URLTemplate,
			InputTemplate:  &inputTemplate,
			HeaderTemplate: &headersTemplate,
			OutputTemplate: &outputTemplate,
			Mode:           &webhookAsyncMode,
			Auth: &graphql.Auth{
				Credential: &graphql.OAuthCredentialData{
					ClientID:     "client-id",
					ClientSecret: "client-secret",
					URL:          "https://test-domain.com/oauth/token",
				},
			},
		},
		Data: web_hook.RequestData{Application: app},
	}

	client := webhook.DefaultClient{
		OAuthClientProviderFunc: func(_ context.Context, client http.Client, _ *graphql.OAuthCredentialData) *http.Client {
			client.Transport = mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					Header:     http.Header{"Location": []string{mockedLocationURL}},
					StatusCode: http.StatusAccepted,
				},
			}
			return &client
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

	require.NoError(t, err)
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
		Webhook: graphql.Webhook{
			CorrelationIDKey: &correlationIDKey,
			URLTemplate:      &URLTemplate,
			InputTemplate:    &inputTemplate,
			HeaderTemplate:   &headersTemplate,
			OutputTemplate:   &outputTemplate,
			Mode:             &webhookAsyncMode,
		},
		Data:          web_hook.RequestData{Application: app, Headers: map[string]string{}},
		CorrelationID: correlationID,
	}

	client := webhook.DefaultClient{
		HTTPClient: http.Client{
			Transport: mockedTransport{
				resp: &http.Response{
					Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
					Header:     http.Header{"Location": []string{mockedLocationURL}},
					StatusCode: http.StatusAccepted,
				},
				roundTripExpectations: func(r *http.Request) {
					require.NotEmpty(t, r.Header[correlationIDKey])
					require.Equal(t, r.Header[correlationIDKey][0], correlationID)
				},
			},
		},
	}

	_, err := client.Do(context.Background(), webhookReq)

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
