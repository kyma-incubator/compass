package graphql

import (
	"net/url"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

var emptyApplicationLifecycleWebhookRequestObject = &webhook.ApplicationLifecycleWebhookRequestObject{
	Application: &Application{
		BaseEntity: &BaseEntity{},
	},
}

var emptyFormationConfigurationChangeInput = &webhook.FormationConfigurationChangeInput{
	Assignment:        &webhook.FormationAssignment{},
	ReverseAssignment: &webhook.FormationAssignment{},
	ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: &model.ApplicationTemplate{},
		Labels:              map[string]interface{}{},
	},
	Application: &webhook.ApplicationWithLabels{
		Application: &model.Application{
			BaseEntity: &model.BaseEntity{},
		},
		Labels: map[string]interface{}{},
	},
	Runtime: &webhook.RuntimeWithLabels{
		Runtime: &model.Runtime{},
		Labels:  map[string]interface{}{},
	},
	RuntimeContext: &webhook.RuntimeContextWithLabels{
		RuntimeContext: &model.RuntimeContext{},
		Labels:         map[string]interface{}{},
	},
}

var emptyApplicationTenantMappingInput = &webhook.ApplicationTenantMappingInput{
	Assignment:        &webhook.FormationAssignment{},
	ReverseAssignment: &webhook.FormationAssignment{},
	SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: &model.ApplicationTemplate{},
		Labels:              map[string]interface{}{},
	},
	SourceApplication: &webhook.ApplicationWithLabels{
		Application: &model.Application{
			BaseEntity: &model.BaseEntity{},
		},
		Labels: map[string]interface{}{},
	},
	TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
		ApplicationTemplate: &model.ApplicationTemplate{},
		Labels:              map[string]interface{}{},
	},
	TargetApplication: &webhook.ApplicationWithLabels{
		Application: &model.Application{
			BaseEntity: &model.BaseEntity{},
		},
		Labels: map[string]interface{}{},
	},
}

var webhookTemplateInputByType = map[WebhookType]webhook.TemplateInput{
	WebhookTypeRegisterApplication:      emptyApplicationLifecycleWebhookRequestObject,
	WebhookTypeUnregisterApplication:    emptyApplicationLifecycleWebhookRequestObject,
	WebhookTypeConfigurationChanged:     emptyFormationConfigurationChangeInput,
	WebhookTypeApplicationTenantMapping: emptyApplicationTenantMappingInput,
}

// Validate missing godoc
func (i WebhookInput) Validate() error {
	if err := validation.ValidateStruct(&i,
		validation.Field(&i.Type, validation.Required, validation.In(WebhookTypeConfigurationChanged, WebhookTypeApplicationTenantMapping, WebhookTypeRegisterApplication, WebhookTypeUnregisterApplication, WebhookTypeOpenResourceDiscovery)),
		validation.Field(&i.URL, is.URL, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.CorrelationIDKey, validation.RuneLength(0, longStringLengthLimit)),
		validation.Field(&i.Mode, validation.In(WebhookModeSync, WebhookModeAsync, WebhookModeAsyncCallback), validation.When(i.Type == WebhookTypeConfigurationChanged || i.Type == WebhookTypeApplicationTenantMapping, validation.In(WebhookModeSync, WebhookModeAsyncCallback)).Else(validation.NotIn(WebhookModeAsyncCallback))),
		validation.Field(&i.RetryInterval, validation.Min(0)),
		validation.Field(&i.Timeout, validation.Min(0)),
		validation.Field(&i.Auth),
	); err != nil {
		return err
	}

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
		if requestObject == nil {
			return apperrors.NewInvalidDataError("missing template input for type: %s", i.Type)
		}
		if _, err := requestObject.ParseURLTemplate(i.URLTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook url template: %s", err)
		}
	}

	if i.InputTemplate != nil {
		if requestObject == nil {
			return apperrors.NewInvalidDataError("missing template input for type: %s", i.Type)
		}
		if _, err := requestObject.ParseInputTemplate(i.InputTemplate); err != nil {
			return apperrors.NewInvalidDataError("failed to parse webhook input template: %s", err)
		}
	}

	if i.HeaderTemplate != nil {
		if requestObject == nil {
			return apperrors.NewInvalidDataError("missing template input for type: %s", i.Type)
		}
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

	return nil
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
