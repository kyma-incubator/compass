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

	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/pkg/errors"
)

var allowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

// TemplateInput is an interface that unions all structs that can act as a template input for a webhook
type TemplateInput interface {
	ParseURLTemplate(tmpl *string) (*URL, error)
	ParseInputTemplate(tmpl *string) ([]byte, error)
	ParseHeadersTemplate(tmpl *string) (http.Header, error)
}

// FormationAssignmentTemplateInput is an interface that unions all structs that can act as a template input for a webhook
type FormationAssignmentTemplateInput interface {
	TemplateInput
	GetParticipantsIDs() []string
	SetAssignment(*model.FormationAssignment)
	SetReverseAssignment(*model.FormationAssignment)
	Clone() FormationAssignmentTemplateInput
}

// Mode is an enum for the mode of the webhook (sync or async)
type Mode string

// ResponseObject struct contains parts of response that might be needed for later processing of Webhook response
type ResponseObject struct {
	Body    map[string]string
	Headers map[string]string
}

// URL missing godoc
type URL struct {
	Method *string `json:"method"`
	Path   *string `json:"path"`
}

// Response defines the schema for Webhook output templates
type Response struct {
	Config               *string `json:"config"`
	Location             *string `json:"location"`
	State                *string `json:"state"`
	SuccessStatusCode    *int    `json:"success_status_code"`
	IncompleteStatusCode *int    `json:"incomplete_status_code"`
	ActualStatusCode     *int    `json:"-"`
	GoneStatusCode       *int    `json:"gone_status_code"`
	Error                *string `json:"error"`
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

// Validate missing godoc
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

// Validate missing godoc
func (r *Response) Validate() error {
	if r.SuccessStatusCode == nil {
		return errors.New("missing Output Template success status code field")
	}

	if r.Error == nil {
		return errors.New("missing Output Template error field")
	}

	return nil
}

// Validate missing godoc
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

// ParseOutputTemplate missing godoc
func (rd *ResponseObject) ParseOutputTemplate(tmpl *string) (*Response, error) {
	var resp Response
	return &resp, parseTemplate(tmpl, *rd, &resp)
}

// ParseStatusTemplate missing godoc
func (rd *ResponseObject) ParseStatusTemplate(tmpl *string) (*ResponseStatus, error) {
	var respStatus ResponseStatus
	return &respStatus, parseTemplate(tmpl, *rd, &respStatus)
}

func parseTemplate(tmpl *string, data interface{}, dest interface{}) error {
	t, err := template.New("").Funcs(templatehelper.GetFuncMap()).Option("missingkey=zero").Parse(*tmpl)
	if err != nil {
		return err
	}

	res := new(bytes.Buffer)
	if err = t.Execute(res, data); err != nil {
		return err
	}

	// <nil> comes after parsing the template with a go field that is a nil pointer
	// As we are expecting the resulting object to be valid JSON object, the <nil> value on its own is misleading in the following contexts
	// If we are working with a *string value we add quotes in the template around it.
	// If the *string is nil, it would result in "<nil>", which is not what we want,
	// but rather an empty string as it is the default value for an empty string in JSON.
	// This is in order to remove the <nil> that comes in templates that are surrounded by quotes that come from the template.
	resBytes := bytes.ReplaceAll(res.Bytes(), []byte(`"<nil>"`), []byte(`""`))
	// In other cases, we do not add quotes around the template, in such cases the value should be null,
	// as it is the correct default value for null JSON objects
	resBytes = bytes.ReplaceAll(resBytes, []byte(`<nil>`), []byte(`null`))
	if err = json.Unmarshal(resBytes, dest); err != nil {
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
