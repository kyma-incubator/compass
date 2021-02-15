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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"text/template"
)

// Client defines a general purpose Webhook executor
type Client interface {
	Do(ctx context.Context, webhook graphql.Webhook, reqData graphql.RequestData, correlationID string) (graphql.Response, error) // TODO: Move Response and other templates to a better package - maybe even in the Controller project itself
}

type DefaultClient struct {
}

func (d DefaultClient) Do(ctx context.Context, webhook graphql.Webhook, reqData graphql.RequestData, correlationID string) (*graphql.Response, error) {
	url := webhook.URL
	if webhook.URLTemplate != nil {
		resultURL, err := prepareURL(webhook.URLTemplate, reqData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to prepare webhook URL")
		}
		url = &resultURL
	}

	body, err := prepareBody(webhook.InputTemplate, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare webhook input body")
	}

	headers, err := prepareHeaders(webhook.HeaderTemplate, reqData)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare webhook headers")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, *url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header = headers

	resp, err := http.DefaultClient.Do(req) // TODO: Build custom client, do not rely on default one
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respBody map[string]interface{}
	if err := json.Unmarshal(bytes, &respBody); err != nil {
		return nil, err
	}

	responseData := graphql.ResponseData{
		Body:    respBody,
		Headers: resp.Header,
	}

	response, err := prepareResponse(webhook.OutputTemplate, responseData)
	if err != nil {
		return nil, err
	}

	if *response.SuccessStatusCode != resp.StatusCode {
		// TODO: check
		return nil, errors.Errorf("response success status code was not met - expected %q, got %q", response.SuccessStatusCode, resp.StatusCode)
	}

	// TODO: ...

	return nil, nil
}

func prepareURL(tmpl *string, reqData graphql.RequestData) (string, error) {
	if tmpl == nil {
		return "", errors.New("missing URL template")
	}

	urlTemplate, err := template.New("url").Parse(*tmpl)
	if err != nil {
		return "", err
	}

	result := new(bytes.Buffer)
	err = urlTemplate.Execute(result, reqData)
	if err != nil {
		return "", err
	}

	resultURL := result.String()
	_, err = url.ParseRequestURI(resultURL)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse URL")
	}

	return resultURL, nil
}

func prepareBody(tmpl *string, reqData graphql.RequestData) ([]byte, error) {
	if tmpl == nil {
		return nil, errors.New("missing input template")
	}

	inputTemplate, err := template.New("input").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	err = inputTemplate.Execute(result, reqData)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func prepareHeaders(tmpl *string, reqData graphql.RequestData) (http.Header, error) {
	if tmpl == nil {
		return nil, errors.New("missing headers template")
	}

	headersTemplate, err := template.New("headers").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	err = headersTemplate.Execute(result, reqData)
	if err != nil {
		return nil, err
	}

	var headers http.Header
	if err := json.Unmarshal(result.Bytes(), &headers); err != nil {
		return nil, err
	}

	return headers, nil
}

func prepareResponse(tmpl *string, respData graphql.ResponseData) (*graphql.Response, error) {
	if tmpl == nil {
		return nil, nil
	}

	outputTemplate, err := template.New("output").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	if err := outputTemplate.Execute(result, respData); err != nil {
		return nil, err
	}

	var outputTmpl graphql.Response
	if err := json.Unmarshal(result.Bytes(), &outputTmpl); err != nil {
		return nil, err
	}

	return &outputTmpl, err
}
