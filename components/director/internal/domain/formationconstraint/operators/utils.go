package operators

import (
	"context"
	"runtime/debug"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// reqBodyNameRegex is a regex defined by the destination creator API specifying what destination names are allowed
var reqBodyNameRegex = "[a-zA-Z0-9_-]{1,64}"

// RetrieveFormationAssignmentPointer converts the provided memory address in form of an integer back to the model.FormationAssignment pointer structure
// It's important the provided memory address to stores information about model.FormationAssignment entity, otherwise the result could be very abnormal
func RetrieveFormationAssignmentPointer(ctx context.Context, joinPointDetailsAssignmentMemoryAddress uintptr) (*model.FormationAssignment, error) {
	if joinPointDetailsAssignmentMemoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The join point details' assignment memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting join point details' assignment address: %d to type: %T", joinPointDetailsAssignmentMemoryAddress, &model.FormationAssignment{})
			debug.PrintStack()
		}
	}()
	joinPointAssignmentPointer := (*model.FormationAssignment)(unsafe.Pointer(joinPointDetailsAssignmentMemoryAddress))

	return joinPointAssignmentPointer, nil
}

// RetrieveWebhookPointerFromMemoryAddress converts the provided uninterpreted memory address in form of an integer back to the model.Webhook pointer structure
// It's important the provided memory address to stores information about model.Webhook entity, otherwise the result could be very abnormal
func RetrieveWebhookPointerFromMemoryAddress(ctx context.Context, webhookMemoryAddress uintptr) (*graphql.Webhook, error) {
	if webhookMemoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The webhook memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting join point details' assignment address: %d to type: %T", webhookMemoryAddress, &model.Webhook{})
			debug.PrintStack()
		}
	}()
	joinPointWebhookPointer := (*graphql.Webhook)(unsafe.Pointer(webhookMemoryAddress))

	return joinPointWebhookPointer, nil
}

// RetrieveNotificationStatusReportPointer converts the provided memory address in form of an integer back to the statusreport.NotificationStatusReport pointer structure
// It's important the provided memory address to stores information about model.FormationAssignment entity, otherwise the result could be very abnormal
func RetrieveNotificationStatusReportPointer(ctx context.Context, notificationStatusReportMemoryAddress uintptr) (*statusreport.NotificationStatusReport, error) {
	if notificationStatusReportMemoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The join point details' notification status report memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting join point details' notification status report memory address: %d to type: %T", notificationStatusReportMemoryAddress, &model.FormationAssignment{})
			debug.PrintStack()
		}
	}()
	notificationStatusReport := (*statusreport.NotificationStatusReport)(unsafe.Pointer(notificationStatusReportMemoryAddress))

	return notificationStatusReport, nil
}
