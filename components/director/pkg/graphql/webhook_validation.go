package graphql

import (
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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
			log.D().Errorf("failed to parse URI: %s", err.Error())
			return apperrors.NewInvalidDataError("failed to parse webhook url")
		}
	}

	requestObject := webhook.RequestObject{Application: &Application{BaseEntity: &BaseEntity{}}}
	if i.URLTemplate != nil {
		if _, err := requestObject.ParseURLTemplate(i.URLTemplate); err != nil {
			log.D().Errorf("failed to parse URL Template: %s", err.Error())
			return apperrors.NewInvalidDataError("failed to parse webhook url template")
		}
	}

	if i.InputTemplate != nil {
		if _, err := requestObject.ParseInputTemplate(i.InputTemplate); err != nil {
			log.D().Errorf("failed to parse Input Template: %s", err.Error())
			return apperrors.NewInvalidDataError("failed to parse webhook input template")
		}
	}

	if i.HeaderTemplate != nil {
		if _, err := requestObject.ParseHeadersTemplate(i.HeaderTemplate); err != nil {
			log.D().Errorf("failed to parse Headers Template: %s", err.Error())
			return apperrors.NewInvalidDataError("failed to parse webhook headers template")
		}
	}

	if i.OutputTemplate == nil && isOutTemplateMandatory(i.Type) {
		return apperrors.NewInvalidDataError("outputTemplate is required for type: %v", i.Type)
	}

	var responseObject webhook.ResponseObject
	if i.OutputTemplate != nil {
		if _, err := responseObject.ParseOutputTemplate(i.OutputTemplate); err != nil {
			log.D().Errorf("failed to parse Output Template: %s", err.Error())
			return apperrors.NewInvalidDataError("failed to parse webhook output template")
		}
	}

	if i.Mode != nil && *i.Mode == WebhookModeAsync {
		if i.StatusTemplate != nil {
			if _, err := responseObject.ParseStatusTemplate(i.StatusTemplate); err != nil {
				log.D().Errorf("failed to parse Status Template: %s", err.Error())
				return apperrors.NewInvalidDataError("failed to parse webhook status template")
			}
		} else {
			return apperrors.NewInvalidDataError("missing webhook status template")
		}
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Type, validation.Required, validation.In(WebhookTypeConfigurationChanged, WebhookTypeRegisterApplication, WebhookTypeUnregisterApplication, WebhookTypeOpenResourceDiscovery)),
		validation.Field(&i.URL, is.URL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.CorrelationIDKey, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Mode, validation.In(WebhookModeSync, WebhookModeAsync)),
		validation.Field(&i.RetryInterval, validation.Min(0)),
		validation.Field(&i.Timeout, validation.Min(0)),
		validation.Field(&i.Auth),
	)
}

func isOutTemplateMandatory(webhookType WebhookType) bool {
	switch webhookType {
	case WebhookTypeRegisterApplication,
		WebhookTypeUnregisterApplication:
		return true
	default:
		return false
	}
}
