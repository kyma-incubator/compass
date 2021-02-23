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
	"encoding/json"
	"net/http"
	"net/url"
	"text/template"

	"github.com/pkg/errors"
)

type Mode string

const (
	ModeSync  Mode = "SYNC"
	ModeAsync Mode = "ASYNC"
)

// Resource is used to identify entities which can be part of a webhook's request data
type Resource interface {
	Sentinel()
}

// RequestData struct contains parts of request that might be needed for later processing of a Webhook request
type RequestData struct {
	Application Resource
	TenantID    string
	Headers     map[string]string
}

// ResponseData struct contains parts of response that might be needed for later processing of Webhook response
type ResponseData struct {
	Body    map[string]string
	Headers map[string]string
}

type URL struct {
	Method *string `json:"method"`
	Path   *string `json:"path"`
}

// Response defines the schema for Webhook output templates
type Response struct {
	Location          *string `json:"location"`
	SuccessStatusCode *int    `json:"success_status_code"`
	Error             *string `json:"error"`
}

// ResponseStatus defines the schema for Webhook status templates when dealing with async webhooks
type ResponseStatus struct {
	Status                     *string `json:"status"`
	SuccessStatusCode          *int    `json:"success_status_code"`
	SuccessStatusIdentifier    *string `json:"success_status_identifier"`
	InProgressStatusIdentifier *string `json:"in_progress_status_identifier"`
	FailedStatusIdentifier     *string `json:"failed_status_identifier"`
	Error                      *string `json:"error"`
}

func (ot *URL) Validate() error {
	if ot.Method == nil {
		return errors.New("missing URL Template method field")
	}

	if ot.Path == nil {
		return errors.New("missing URL Template path field")
	}

	_, err := url.ParseRequestURI(*ot.Path)
	if err != nil {
		return errors.Wrap(err, "failed to parse URL Template path field")
	}

	return nil
}

func (ot *Response) Validate() error {
	if ot.Location == nil {
		return errors.New("missing Output Template location field")
	}

	if ot.SuccessStatusCode == nil {
		return errors.New("missing Output Template success status code field")
	}

	if ot.Error == nil {
		return errors.New("missing Output Template error field")
	}

	return nil
}

func (st *ResponseStatus) Validate() error {
	if st.Status == nil {
		return errors.New("missing Status Template status field")
	}

	if st.SuccessStatusCode == nil {
		return errors.New("missing Status Template success status code field")
	}

	if st.SuccessStatusIdentifier == nil {
		return errors.New("missing Status Template success status identifier field")
	}

	if st.InProgressStatusIdentifier == nil {
		return errors.New("missing Status Template in progress status identifier field")
	}

	if st.FailedStatusIdentifier == nil {
		return errors.New("missing Status Template failed status identifier field")
	}

	if st.Error == nil {
		return errors.New("missing Status Template error field")
	}

	return nil
}

func ParseURLTemplate(tmpl *string, reqData RequestData) (*URL, error) {
	if tmpl == nil {
		return nil, nil
	}

	urlTemplate, err := template.New("url").Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	err = urlTemplate.Execute(result, reqData)
	if err != nil {
		return nil, err
	}

	var url URL
	if err := json.Unmarshal(result.Bytes(), &url); err != nil {
		return nil, err
	}

	return &url, url.Validate()
}

func ParseInputTemplate(tmpl *string, reqData RequestData) ([]byte, error) {
	if tmpl == nil {
		return nil, nil
	}

	inputTemplate, err := template.New("input").Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	if err := inputTemplate.Execute(result, reqData); err != nil {
		return nil, err
	}

	res := json.RawMessage{}
	if err := json.Unmarshal(result.Bytes(), &res); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func ParseHeadersTemplate(tmpl *string, reqData RequestData) (http.Header, error) {
	if tmpl == nil {
		return nil, nil
	}

	headerTemplate, err := template.New("header").Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	if err := headerTemplate.Execute(result, reqData); err != nil {
		return nil, err
	}

	if result.Len() == 0 {
		return http.Header{}, nil
	}

	var headers http.Header
	if err := json.Unmarshal(result.Bytes(), &headers); err != nil {
		return nil, err
	}

	return headers, nil
}

func ParseOutputTemplate(inputTmpl, outputTmpl *string, respData ResponseData) (*Response, error) {
	if outputTmpl == nil && inputTmpl != nil {
		return nil, errors.New("missing webhook output template")
	}

	if outputTmpl == nil {
		return nil, nil
	}

	outputTemplate, err := template.New("output").Option("missingkey=zero").Parse(*outputTmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	if err := outputTemplate.Execute(result, respData); err != nil {
		return nil, err
	}

	var outputTmplResp Response
	if err := json.Unmarshal(result.Bytes(), &outputTmplResp); err != nil {
		return nil, err
	}

	return &outputTmplResp, outputTmplResp.Validate()
}

func ParseStatusTemplate(tmpl *string, respData ResponseData) (*ResponseStatus, error) {
	if tmpl == nil {
		return nil, nil
	}

	statusTemplate, err := template.New("status").Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return nil, err
	}

	result := new(bytes.Buffer)
	if err := statusTemplate.Execute(result, respData); err != nil {
		return nil, err
	}

	var statusTmpl ResponseStatus
	if err := json.Unmarshal(result.Bytes(), &statusTmpl); err != nil {
		return nil, err
	}

	return &statusTmpl, statusTmpl.Validate()
}
