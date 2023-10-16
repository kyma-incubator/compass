package operators

import (
	"context"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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

// RedirectNotification is an operator that based on different condition could redirect the formation assignment notification
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

	//ww := &WebhookWrapper{webhook: &graphql.Webhook{}}
	//w, err := RetrieveEntityPointerFromMemoryAddress[*WebhookWrapper, *graphql.Webhook](ctx, ww, ri.WebhookMemoryAddress)
	//if err != nil {
	//	return false, err
	//}

	//entity, err := RetrieveEntityPointerFromMemoryAddress2(ctx, &graphql.Webhook{}, ri.WebhookMemoryAddress)
	//if err != nil {
	//	return false, err
	//}
	//w, ok := entity.(*graphql.Webhook)
	//if !ok {
	//	return false, errors.New("Failed to cast to webhook entity")
	//}

	w, err := RetrieveWebhookPointerFromMemoryAddress(ctx, ri.WebhookMemoryAddress)
	if err != nil {
		return false, err
	}

	if !ri.Condition {
		log.C(ctx).Infof("The condition for the redirect notification operator is not satisfied. Returning...")
		return true, nil
	}

	if w.URLTemplate != nil && ri.URLTemplate != "" {
		log.C(ctx).Infof("Current webhook URL template is: '%s', changing it to: '%s'", *w.URLTemplate, ri.URLTemplate)
		w.URLTemplate = &ri.URLTemplate
	}

	if w.URL != nil && ri.URL != "" {
		log.C(ctx).Infof("Current webhook URL is: '%s', changing it to: '%s'", *w.URL, ri.URL)
		w.URL = &ri.URL
	}

	log.C(ctx).Infof("Finished executing operator: %s", RedirectNotificationOperator)
	return true, nil
}
