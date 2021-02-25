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
	webhook := request.Webhook

	var method string
	url := webhook.URL
	if webhook.URLTemplate != nil {
		resultURL, err := web_hook.ParseURLTemplate(webhook.URLTemplate, request.Data)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook URL")
		}
		url = resultURL.Path
		method = *resultURL.Method
	}

	if url == nil {
		return nil, errors.New("missing webhook url")
	}

	body, err := web_hook.ParseInputTemplate(webhook.InputTemplate, request.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse webhook input body")
	}

	headers, err := web_hook.ParseHeadersTemplate(webhook.HeaderTemplate, request.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse webhook headers")
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

	responseData, err := parseResponseData(resp)
	if err != nil {
		return nil, err
	}

	mode := graphql.WebhookModeSync
	if webhook.Mode != nil {
		mode = *webhook.Mode
	}
	response, err := web_hook.ParseOutputTemplate(webhook.InputTemplate, webhook.OutputTemplate, *responseData)
	if err != nil {
		return nil, err
	}

	if *response.Location == "" && mode == graphql.WebhookModeAsync {
		return nil, errors.New("missing location url after executing async webhook")
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func (c *client) Poll(ctx context.Context, request *PollRequest) (*web_hook.ResponseStatus, error) {
	webhook := request.Webhook

	headers, err := web_hook.ParseHeadersTemplate(webhook.HeaderTemplate, request.Data)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse webhook headers")
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

	responseData, err := parseResponseData(resp)
	if err != nil {
		return nil, err
	}

	response, err := web_hook.ParseStatusTemplate(webhook.StatusTemplate, *responseData)
	if err != nil {
		return nil, err
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func parseResponseData(resp *http.Response) (*web_hook.ResponseData, error) {
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

	return &web_hook.ResponseData{
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
