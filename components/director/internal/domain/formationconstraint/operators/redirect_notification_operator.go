package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	// RedirectNotificationOperator represents the redirect notification operator
	RedirectNotificationOperator = "RedirectNotification"
)

// NewRedirectNotificationInput is input constructor for DestinationCreatorOperator. It returns empty OperatorInput
func NewRedirectNotificationInput() OperatorInput {
	return &formationconstraint.RedirectNotificationInput{}
}

func (e *ConstraintEngine) RedirectNotification(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Starting executing operator: %s", RedirectNotificationOperator)

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panic("recovered panic")
			debug.PrintStack()
		}
	}()

	ri, ok := input.(*formationconstraint.RedirectNotificationInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %s", RedirectNotificationOperator)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", ri.ResourceType, ri.ResourceSubtype, ri.Location.ConstraintType, ri.Location.OperationName, ri.Operation)

	w, err := RetrieveEntityPointerFromMemoryAddress[*model.Webhook](ctx, &model.Webhook{}, ri.WebhookMemoryAddress)
	if err != nil {
		return false, err
	}

	if ri.Condition {
		if w.URLTemplate != nil {
			log.C(ctx).Infof("Current webhook URL template is: '%s'", *w.URLTemplate)
			var urlData *webhook.URL
			if err := json.Unmarshal([]byte(*w.URLTemplate), urlData); err != nil {
				return false, errors.Wrapf(err, "while unmarhalling webhook URL template")
			}

			if err := urlData.Validate(); err != nil {
				return false, err
			}

			*urlData.Path = ri.URL
			urlDataBytes, err := json.Marshal(urlData)
			if err != nil {
				return false, err
			}

			modifiedURL := string(urlDataBytes)
			log.C(ctx).Infof("Changing the URL template to: '%s'", modifiedURL)
			w.URLTemplate = &modifiedURL
		}

		if w.URL != nil {
			log.C(ctx).Infof("Current webhook URL is: '%s'. Changing it to: '%s'", *w.URL, ri.URL)
			w.URL = &ri.URL
		}
	}

	log.C(ctx).Infof("Finished executing operator: %s", RedirectNotificationOperator)
	return true, nil
}
