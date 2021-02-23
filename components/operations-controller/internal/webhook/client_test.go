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

var invalidTemplate = "invalidTemplate"
var mockedError = "mocked error"
var webhookAsyncMode = graphql.WebhookModeAsync

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

func TestClient_Do_WhenWebhookResponseContainsError_ShouldReturnError(t *testing.T) {
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{\"application_id\": \"{{.Application.ID}}\",\"name\": \"{{.Application.Name}}\"}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
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
	outputTemplate := "{\"location\":\"{{.Headers.location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
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
	// TODO
	URLTemplate := "{\"method\": \"DELETE\",\"path\":\"https://test-domain.com/api/v1/applicaitons/{{.Application.ID}}\"}"
	inputTemplate := "{}"
	headersTemplate := "{\"user-identity\":[\"{{.Headers.Client_user}}\"]}"
	outputTemplate := "{\"location\":\"{{.Headers.location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}"
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

	require.NoError(t, err)
	//require.Contains(t, err.Error(), "response success status code was not met")
}

func TestClient_Do_WhenSuccessfulOAuthWebhook_ShouldBeSuccessful(t *testing.T) {}

func TestClient_Do_WhenMissingCorrelationID_ShouldBeSuccessful(t *testing.T) {}

type mockedTransport struct {
	resp *http.Response
	err  error
}

func (m mockedTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.resp, m.err
}
