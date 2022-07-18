package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"net/url"
)

var emptyApplicationLifecycleWebhookRequestObject = &webhook.ApplicationLifecycleWebhookRequestObject{
	Application: &Application{
		BaseEntity: &BaseEntity{},
	},
}

var emptyFormationConfigurationChangeInput = &webhook.FormationConfigurationChangeInput{
	ApplicationTemplate: webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: model.ApplicationTemplate{},
		Labels:              map[string]interface{}{},
	},
	Application: webhook.ApplicationWithLabels{
		Application: model.Application{
			BaseEntity: &model.BaseEntity{},
		},
		Labels: map[string]interface{}{},
	},
	Runtime: webhook.RuntimeWithLabels{
		Runtime: model.Runtime{},
		Labels:  map[string]interface{}{},
	},
	RuntimeContext: webhook.RuntimeContextWithLabels{
		RuntimeContext: model.RuntimeContext{},
		Labels:         map[string]interface{}{},
	},
}

var webhookTemplateInputByType = map[WebhookType]webhook.TemplateInput{
	WebhookTypeRegisterApplication:   emptyApplicationLifecycleWebhookRequestObject,
	WebhookTypeUnregisterApplication: emptyApplicationLifecycleWebhookRequestObject,
	WebhookTypeUnpairApplication:     emptyApplicationLifecycleWebhookRequestObject,
	WebhookTypeConfigurationChanged:  emptyFormationConfigurationChangeInput,
}

// Validate missing godoc
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
			return apperrors.NewInvalidDataError("failed to parse webhook url: %s", err)
		}
	}

	requestObject := webhookTemplateInputByType[i.Type]
	if i.URLTemplate != nil {
		if _, err := requestObject.ParseURLTemplate(i.URLTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook url template: %s", err)
		}
	}

	if i.InputTemplate != nil {
		if _, err := requestObject.ParseInputTemplate(i.InputTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook input template: %s", err)
		}
	}

	if i.HeaderTemplate != nil {
		if _, err := requestObject.ParseHeadersTemplate(i.HeaderTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook headers template: %s", err)
		}
	}

	if i.OutputTemplate == nil && isOutTemplateMandatory(i.Type) {
		return apperrors.NewInvalidDataError("outputTemplate is required for type: %v", i.Type)
	}

	var responseObject webhook.ResponseObject
	if i.OutputTemplate != nil {
		if _, err := responseObject.ParseOutputTemplate(i.OutputTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook output template: %s", err)
		}
	}

	if i.Mode != nil && *i.Mode == WebhookModeAsync {
		if i.StatusTemplate != nil {
			if _, err := responseObject.ParseStatusTemplate(i.StatusTemplate); err != nil {
				return apperrors.NewInvalidDataError("failed to parse webhook status template: %s", err)
			}
		} else {
			return apperrors.NewInvalidDataError("missing webhook status template")
		}
	}

	return validation.ValidateStruct(&i,
		validation.Field(&i.Type, validation.Required, validation.In(WebhookTypeConfigurationChanged, WebhookTypeRegisterApplication, WebhookTypeUnregisterApplication, WebhookTypeOpenResourceDiscovery, WebhookTypeUnpairApplication)),
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
