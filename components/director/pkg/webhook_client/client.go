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

package webhookclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/pkg/errors"
)

const emptyBody = `{}`

type client struct {
	httpClient *http.Client
	mtlsClient *http.Client
}

// NewClient creates a new webhook client
func NewClient(httpClient *http.Client, mtlsClient *http.Client) *client {
	return &client{
		httpClient: httpClient,
		mtlsClient: mtlsClient,
	}
}

func (c *client) Do(ctx context.Context, request *Request) (*webhook.Response, error) {
	var err error
	webhook := request.Webhook

	if webhook.OutputTemplate == nil {
		return nil, errors.Errorf("missing output template")
	}

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
		return nil, errors.Errorf("missing webhook url")
	}

	body := []byte(emptyBody)
	if webhook.InputTemplate != nil {
		body, err = request.Object.ParseInputTemplate(webhook.InputTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook input body")
		}
	}

	fmt.Printf("ALEX 1. %+v \n", *webhook.HeaderTemplate)

	headers := http.Header{}
	if webhook.HeaderTemplate != nil {
		headers, err = request.Object.ParseHeadersTemplate(webhook.HeaderTemplate)

		fmt.Printf("ALEX 2. %+v \n", headers)

		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook headers")
		}
	}

	fmt.Printf("ALEX 3. %+v \n", headers)

	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, webhook.CorrelationIDKey, &request.CorrelationID)

	req, err := http.NewRequestWithContext(ctx, method, *url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := c.executeRequestWithCorrectClient(ctx, req, webhook)
	if err != nil {
		return nil, errors.Wrap(err, "while initially executing webhook")
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	responseObject, err := parseResponseObject(resp)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Info(fmt.Sprintf("Webhook response object: %v", *responseObject))

	response, err := responseObject.ParseOutputTemplate(webhook.OutputTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse response into webhook output template")
	}

	if err = checkForGoneStatus(resp, response.GoneStatusCode); err != nil {
		return response, err
	}

	isLocationEmpty := response.Location != nil && *response.Location == ""
	isAsyncWebhook := webhook.Mode != nil && *webhook.Mode == graphql.WebhookModeAsync

	if isLocationEmpty && isAsyncWebhook {
		return nil, errors.Errorf("missing location url after executing async webhook: HTTP response status %+v with body %s", resp.Status, responseObject.Body)
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func (c *client) Poll(ctx context.Context, request *PollRequest) (*webhook.ResponseStatus, error) {
	var err error
	webhook := request.Webhook

	if webhook.StatusTemplate == nil {
		return nil, errors.Errorf("missing status template")
	}

	headers := http.Header{}
	if webhook.HeaderTemplate != nil {
		headers, err = request.Object.ParseHeadersTemplate(webhook.HeaderTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "unable to parse webhook headers")
		}
	}

	ctx = correlation.SaveCorrelationIDHeaderToContext(ctx, webhook.CorrelationIDKey, &request.CorrelationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, request.PollURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := c.executeRequestWithCorrectClient(ctx, req, webhook)
	if err != nil {
		return nil, errors.Wrap(err, "while executing webhook for poll")
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(ctx).Error(err, "Failed to close HTTP response body")
		}
	}()

	responseObject, err := parseResponseObject(resp)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Info(fmt.Sprintf("Webhook response object: %v", *responseObject))

	response, err := responseObject.ParseStatusTemplate(webhook.StatusTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse response status into status template")
	}

	return response, checkForErr(resp, response.SuccessStatusCode, response.Error)
}

func (c *client) executeRequestWithCorrectClient(ctx context.Context, req *http.Request, webhook graphql.Webhook) (*http.Response, error) {
	if webhook.Auth != nil {
		if str.PtrStrToStr(webhook.Auth.AccessStrategy) == string(accessstrategy.CMPmTLSAccessStrategy) {
			return c.mtlsClient.Do(req)
		} else if webhook.Auth.Credential != nil {
			ctx = saveToContext(ctx, webhook.Auth.Credential)
			req = req.WithContext(ctx)
			return c.httpClient.Do(req)
		} else {
			return nil, errors.New("could not determine auth flow for webhook")
		}
	} else {
		return c.httpClient.Do(req)
	}
}

func parseResponseObject(resp *http.Response) (*webhook.ResponseObject, error) {
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := make(map[string]string, 0)
	if len(respBody) > 0 {
		tmpBody := make(map[string]interface{})
		if err := json.Unmarshal(respBody, &tmpBody); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("failed to unmarshall HTTP response with body %s", respBody))
		}

		for k, v := range tmpBody {
			if v == nil {
				continue
			}
			body[k] = fmt.Sprintf("%v", v)
		}
	}

	headers := make(map[string]string, 0)
	for key, value := range resp.Header {
		headers[key] = value[0]
	}

	return &webhook.ResponseObject{
		Headers: headers,
		Body:    body,
	}, nil
}

func checkForErr(resp *http.Response, successStatusCode *int, errorMessage *string) error {
	var errMsg string
	if *successStatusCode != resp.StatusCode {
		errMsg += fmt.Sprintf("response success status code was not met - expected %d, got %d; ", *successStatusCode, resp.StatusCode)
	}

	if errorMessage != nil && *errorMessage != "" {
		errMsg += fmt.Sprintf("received error while calling external system: %s", *errorMessage)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func checkForGoneStatus(resp *http.Response, goneStatusCode *int) error {
	if goneStatusCode != nil && resp.StatusCode == *goneStatusCode {
		return NewWebhookStatusGoneErr(*goneStatusCode)
	}
	return nil
}

func saveToContext(ctx context.Context, credentialData graphql.CredentialData) context.Context {
	var credentials auth.Credentials

	switch v := credentialData.(type) {
	case *graphql.BasicCredentialData:
		credentials = &auth.BasicCredentials{
			Username: v.Username,
			Password: v.Password,
		}
	case *graphql.OAuthCredentialData:
		credentials = &auth.OAuthCredentials{
			ClientID:     v.ClientID,
			ClientSecret: v.ClientSecret,
			TokenURL:     v.URL,
		}
	default:
		return ctx
	}

	return auth.SaveToContext(ctx, credentials)
}
