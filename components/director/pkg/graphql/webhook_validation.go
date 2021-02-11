package graphql

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"text/template"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// RequestData struct contains parts of request that might be needed for later processing of a Webhook request
type RequestData struct {
	Application Application
	TenantID    string
	Headers     http.Header
}

// ResponseData struct contains parts of response that might be needed for later processing of Webhook response
type ResponseData struct {
	Body    map[string]interface{}
	Headers http.Header
}

// OutputTemplate defines the schema for Webhook output templates
type OutputTemplate struct {
	Location          *string `json:"location"`
	SuccessStatusCode *int    `json:"success_status_code"`
	Error             *string `json:"error"`
}

// StatusTemplate defines the schema for Webhook status templates when dealing with async webhooks
type StatusTemplate struct {
	Status            *string `json:"status"`
	SuccessStatusCode *int    `json:"success_status_code"`
	Error             *string `json:"error"`
}

func (ot *OutputTemplate) Validate(webhookMode *WebhookMode) error {
	if webhookMode != nil && *webhookMode == WebhookModeAsync && ot.Location == nil {
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

func (st *StatusTemplate) Validate() error {
	if st.Status == nil {
		return errors.New("missing Status Template status field")
	}

	if st.SuccessStatusCode == nil {
		return errors.New("missing Status Template success status code field")
	}

	if st.Error == nil {
		return errors.New("missing Status Template error field")
	}

	return nil
}

func (i WebhookInput) Validate() error {
	if i.URL == nil && i.URLTemplate == nil {
		return apperrors.NewInvalidDataError("missing webhook url")
	}

	if i.URL != nil && i.URLTemplate != nil {
		return apperrors.NewInvalidDataError("cannot provide both webhook url and url template")
	}

	if i.URLTemplate == nil {
		i.URLTemplate = i.URL
	}

	reqData := RequestData{Application: Application{BaseEntity: &BaseEntity{}}}
	if err := i.validateURLTemplate(reqData); err != nil {
		return err
	}

	if err := i.validateInputTemplate(reqData); err != nil {
		return err
	}

	if err := i.validateHeadersTemplate(reqData); err != nil {
		return err
	}

	var respData ResponseData
	if err := i.validateOutputTemplate(respData); err != nil {
		return err
	}

	if i.Mode != nil && *i.Mode == WebhookModeAsync {
		if err := i.validateStatusTemplate(respData); err != nil {
			return err
		}
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Type, validation.Required, validation.In(WebhookTypeConfigurationChanged, WebhookTypeRegisterApplication, WebhookTypeUnregisterApplication)),
		validation.Field(&i.URL, is.URL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.CorrelationIDKey, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Mode, validation.In(WebhookModeSync, WebhookModeAsync)),
		validation.Field(&i.RetryInterval, validation.Min(0)),
		validation.Field(&i.Timeout, validation.Min(0)),
		validation.Field(&i.Auth),
	)
}

func (i WebhookInput) validateURLTemplate(reqData RequestData) error {
	if i.URLTemplate == nil {
		return nil
	}

	urlTemplate, err := template.New("url").Parse(*i.URLTemplate)
	if err != nil {
		return err
	}

	finalURL := new(bytes.Buffer)
	err = urlTemplate.Execute(finalURL, reqData)
	if err != nil {
		return err
	}

	_, err = url.ParseRequestURI(finalURL.String())
	if err != nil {
		return err
	}

	return nil
}

func (i WebhookInput) validateInputTemplate(reqData RequestData) error {
	if i.InputTemplate == nil {
		return nil
	}

	inputTemplate, err := template.New("input").Parse(*i.InputTemplate)
	if err != nil {
		return err
	}

	inputBody := new(bytes.Buffer)
	if err := inputTemplate.Execute(inputBody, reqData); err != nil {
		return err
	}

	res := json.RawMessage{}
	return json.Unmarshal(inputBody.Bytes(), &res)
}

func (i WebhookInput) validateHeadersTemplate(reqData RequestData) error {
	if i.HeaderTemplate == nil {
		return nil
	}

	headerTemplate, err := template.New("header").Parse(*i.HeaderTemplate)
	if err != nil {
		return err
	}

	headers := new(bytes.Buffer)
	if err := headerTemplate.Execute(headers, reqData); err != nil {
		return err
	}

	if headers.Len() == 0 {
		return nil
	}

	h := map[string][]string{}
	return json.Unmarshal(headers.Bytes(), &h)
}

func (i WebhookInput) validateOutputTemplate(respData ResponseData) error {
	if i.OutputTemplate == nil && i.InputTemplate != nil {
		return errors.New("missing webhook output template")
	}

	if i.OutputTemplate == nil {
		return nil
	}

	outputTemplate, err := template.New("output").Parse(*i.OutputTemplate)
	if err != nil {
		return err
	}

	outputBody := new(bytes.Buffer)
	if err := outputTemplate.Execute(outputBody, respData); err != nil {
		return err
	}

	var outputTmpl OutputTemplate
	if err := json.Unmarshal(outputBody.Bytes(), &outputTmpl); err != nil {
		return err
	}

	return outputTmpl.Validate(i.Mode)
}

func (i WebhookInput) validateStatusTemplate(respData ResponseData) error {
	if i.StatusTemplate == nil {
		return nil
	}

	statusTemplate, err := template.New("status").Parse(*i.StatusTemplate)
	if err != nil {
		return err
	}

	statusBody := new(bytes.Buffer)
	if err := statusTemplate.Execute(statusBody, respData); err != nil {
		return err
	}

	var statusTmpl StatusTemplate
	if err := json.Unmarshal(statusBody.Bytes(), &statusTmpl); err != nil {
		return err
	}

	return statusTmpl.Validate()
}
