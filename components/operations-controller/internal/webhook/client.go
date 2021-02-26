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

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	web_hook "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

type OAuthClientProviderFunc func(ctx context.Context, client http.Client, oauthCreds *graphql.OAuthCredentialData) *http.Client

type client struct {
	httpClient              http.Client
	oAuthClientProviderFunc OAuthClientProviderFunc
}

func NewClient(httpClient http.Client, oAuthClientProviderFunc OAuthClientProviderFunc) *client {
	return &client{
		httpClient:              httpClient,
		oAuthClientProviderFunc: oAuthClientProviderFunc,
	}
}

func (c *client) Do(ctx context.Context, request *Request) (*web_hook.Response, error) {
	var err error
	webhook := request.Webhook

	var method string
	url := webhook.URL
	if webhook.URLTemplate != nil {
		resultURL, err := request.Object.ParseURLTemplate(webhook.URLTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook URL")
		}
		url = resultURL.Path
		method = *resultURL.Method
	}

	if url == nil {
		return nil, errors.New("missing webhook url")
	}

	body := []byte("{}")
	if webhook.InputTemplate != nil {
		body, err = request.Object.ParseInputTemplate(webhook.InputTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook input body")
		}
	}

	headers := http.Header{}
	if webhook.HeaderTemplate != nil {
		headers, err = request.Object.ParseHeadersTemplate(webhook.HeaderTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook headers")
		}
	}

	if webhook.CorrelationIDKey != nil && request.CorrelationID != "" {
		headers.Add(*webhook.CorrelationIDKey, request.CorrelationID)
	}

	req, err := http.NewRequestWithContext(ctx, method, *url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	client := &c.httpClient
	if webhook.Auth != nil {
		basicCreds, isBasicAuth := webhook.Auth.Credential.(*graphql.BasicCredentialData)
		if isBasicAuth {
			req.SetBasicAuth(basicCreds.Username, basicCreds.Password)
		}

		oauthCreds, isOAuth := webhook.Auth.Credential.(*graphql.OAuthCredentialData)
		if isOAuth {
			client = c.oAuthClientProviderFunc(ctx, *client, oauthCreds)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	responseObject, err := parseResponseObject(resp)
	if err != nil {
		return nil, err
	}

	response, err := responseObject.ParseOutputTemplate(webhook.OutputTemplate)
	if err != nil {
		return nil, err
	}

	isLocationEmpty := response.Location != nil && *response.Location == ""
	isAsyncWebhook := webhook.Mode != nil && *webhook.Mode == graphql.WebhookModeAsync

	if isLocationEmpty && isAsyncWebhook {
		return nil, errors.New("missing location url after executing async webhook")
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func (c *client) Poll(ctx context.Context, request *PollRequest) (*web_hook.ResponseStatus, error) {
	var err error
	webhook := request.Webhook

	headers := http.Header{}
	if webhook.HeaderTemplate != nil {
		headers, err = request.Object.ParseHeadersTemplate(webhook.HeaderTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook headers")
		}
	}

	if webhook.CorrelationIDKey != nil && request.CorrelationID != "" {
		headers.Add(*webhook.CorrelationIDKey, request.CorrelationID)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, request.PollURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = headers

	client := &c.httpClient
	if webhook.Auth != nil {
		basicCreds, isBasicAuth := webhook.Auth.Credential.(*graphql.BasicCredentialData)
		if isBasicAuth {
			req.SetBasicAuth(basicCreds.Username, basicCreds.Password)
		}

		oauthCreds, isOAuth := webhook.Auth.Credential.(*graphql.OAuthCredentialData)
		if isOAuth {
			client = c.oAuthClientProviderFunc(ctx, *client, oauthCreds)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	responseObject, err := parseResponseObject(resp)
	if err != nil {
		return nil, err
	}

	response, err := responseObject.ParseStatusTemplate(webhook.StatusTemplate)
	if err != nil {
		return nil, err
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func parseResponseObject(resp *http.Response) (*web_hook.ResponseObject, error) {
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := make(map[string]string, 0)
	if len(bytes) > 0 {
		if err := json.Unmarshal(bytes, &body); err != nil {
			return nil, err
		}
	}

	headers := make(map[string]string, 0)
	for key, value := range resp.Header {
		headers[key] = value[0]
	}

	return &web_hook.ResponseObject{
		Headers: headers,
		Body:    body,
	}, nil
}

func checkForErr(resp *http.Response, successStatusCode *int, error *string) error {
	var errMsg string
	if *successStatusCode != resp.StatusCode {
		errMsg += fmt.Sprintf("response success status code was not met - expected %q, got %q; ", *successStatusCode, resp.StatusCode)
	}

	if error != nil && *error != "" {
		errMsg += fmt.Sprintf("received error while polling external system: %s", *error)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}
