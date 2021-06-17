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
	"fmt"
	"net/http"
	"net/url"
	"text/template"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/pkg/errors"
)

var allowedMethods = []string{"GET", "POST", "PUT", "DELETE"}

type Mode string

// Resource is used to identify entities which can be part of a webhook's request data
type Resource interface {
	Sentinel()
}

// RequestObject struct contains parts of request that might be needed for later processing of a Webhook request
type RequestObject struct {
	Application Resource
	TenantID    string
	Headers     map[string]string
}

// ResponseObject struct contains parts of response that might be needed for later processing of Webhook response
type ResponseObject struct {
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
	GoneStatusCode    *int    `json:"gone_status_code"`
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

func (u *URL) Validate() error {
	if u.Method == nil {
		return errors.New("missing URL Template method field")
	}

	if !isAllowedHTTPMethod(*u.Method) {
		return errors.New(fmt.Sprint("http method not allowed, allowed methods: ", allowedMethods))
	}

	if u.Path == nil {
		return errors.New("missing URL Template path field")
	}

	_, err := url.ParseRequestURI(*u.Path)
	if err != nil {
		return errors.Wrap(err, "failed to parse URL Template path field")
	}

	return nil
}

func (r *Response) Validate() error {
	if r.Location == nil {
		return errors.New("missing Output Template location field")
	}

	if r.SuccessStatusCode == nil {
		return errors.New("missing Output Template success status code field")
	}

	if r.Error == nil {
		return errors.New("missing Output Template error field")
	}

	return nil
}

func (rs *ResponseStatus) Validate() error {
	if rs.Status == nil {
		return errors.New("missing Status Template status field")
	}

	if rs.SuccessStatusCode == nil {
		return errors.New("missing Status Template success status code field")
	}

	if rs.SuccessStatusIdentifier == nil {
		return errors.New("missing Status Template success status identifier field")
	}

	if rs.InProgressStatusIdentifier == nil {
		return errors.New("missing Status Template in progress status identifier field")
	}

	if rs.FailedStatusIdentifier == nil {
		return errors.New("missing Status Template failed status identifier field")
	}

	if rs.Error == nil {
		return errors.New("missing Status Template error field")
	}

	return nil
}

func (rd *RequestObject) ParseURLTemplate(tmpl *string) (*URL, error) {
	var url URL
	return &url, parseTemplate(tmpl, *rd, &url)
}

func (rd *RequestObject) ParseInputTemplate(tmpl *string) ([]byte, error) {
	res := json.RawMessage{}
	return res, parseTemplate(tmpl, *rd, &res)
}

func (rd *RequestObject) ParseHeadersTemplate(tmpl *string) (http.Header, error) {
	var headers http.Header
	return headers, parseTemplate(tmpl, *rd, &headers)
}

func (rd *ResponseObject) ParseOutputTemplate(tmpl *string) (*Response, error) {
	var resp Response
	return &resp, parseTemplate(tmpl, *rd, &resp)
}

func (rd *ResponseObject) ParseStatusTemplate(tmpl *string) (*ResponseStatus, error) {
	var respStatus ResponseStatus
	return &respStatus, parseTemplate(tmpl, *rd, &respStatus)
}

func parseTemplate(tmpl *string, data interface{}, dest interface{}) error {
	t, err := template.New("").Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return err
	}

	if err = json.Unmarshal(res.Bytes(), dest); err != nil {
		return err
	}

	if validatable, ok := dest.(inputvalidation.Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func isAllowedHTTPMethod(method string) bool {
	for _, m := range allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}
