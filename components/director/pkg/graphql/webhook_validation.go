package graphql

import (
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

func (i WebhookInput) Validate() error {
	if i.URL == nil && i.URLTemplate == nil {
		return apperrors.NewInvalidDataError("missing webhook url")
	}

	if i.URL != nil && i.URLTemplate != nil {
		return apperrors.NewInvalidDataError("cannot provide both webhook url and url template")
	}

	if i.URL != nil {
		_, err := url.ParseRequestURI(*i.URL)
		if err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook url")
		}
	}

	reqData := webhook.RequestData{Application: &Application{BaseEntity: &BaseEntity{}}}
	if err := webhook.ValidateURLTemplate(i.URLTemplate, reqData); err != nil {
		return err
	}

	if err := webhook.ValidateInputTemplate(i.InputTemplate, reqData); err != nil {
		return err
	}

	if err := webhook.ValidateHeadersTemplate(i.HeaderTemplate, reqData); err != nil {
		return err
	}

	webhookMode := webhook.ModeSync
	if i.Mode != nil {
		webhookMode = webhook.Mode(*i.Mode)
	}

	var respData webhook.ResponseData
	if err := webhook.ValidateOutputTemplate(i.InputTemplate, i.OutputTemplate, webhookMode, respData); err != nil {
		return err
	}

	if i.Mode != nil && *i.Mode == WebhookModeAsync {
		if err := webhook.ValidateStatusTemplate(i.StatusTemplate, respData); err != nil {
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
