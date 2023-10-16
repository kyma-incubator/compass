package operators

import (
	"context"
	"runtime/debug"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

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

// todo::: delete
type EntityType interface {
	*model.FormationAssignment | *graphql.Webhook
}

type Entity[ET EntityType] interface {
	GetEntity() ET
}

type WebhookWrapper struct {
	webhook *graphql.Webhook
}

func (ww *WebhookWrapper) GetEntity() *graphql.Webhook {
	return ww.webhook
}

//func RetrieveEntityPointerFromMemoryAddress[ETM Entity[ET], ET EntityType](ctx context.Context, e ETM, entityMemoryAddress uintptr) (ET, error) {
//	if entityMemoryAddress == 0 { // the default value of uintptr is 0
//		return nil, errors.New("The entity memory address cannot be 0")
//	}
//
//	entity := e.GetEntity()
//	fmt.Printf("entity type: %T", entity)
//	defer func() {
//		if err := recover(); err != nil {
//			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting join point details' assignment address: %d to type: %T", entityMemoryAddress, entity)
//			debug.PrintStack()
//		}
//	}()
//	entityPointer := (entity)(unsafe.Pointer(entityMemoryAddress))
//
//	return entityPointer, nil
//}

func RetrieveEntityPointerFromMemoryAddress2(ctx context.Context, entity interface{}, entityMemoryAddress uintptr) (interface{}, error) {
	if entityMemoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The entity memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting join point details' entity memory address: %d to pointer", entityMemoryAddress)
			debug.PrintStack()
		}
	}()

	var entityPointer interface{}
	switch entity.(type) {
	case *model.FormationAssignment:
		entityPointer = (*model.FormationAssignment)(unsafe.Pointer(entityMemoryAddress))
	case *graphql.Webhook:
		entityPointer = (*model.Webhook)(unsafe.Pointer(entityMemoryAddress))
	default:
		return nil, errors.Errorf("Coulnd't cast the entity neither to formation assignment nor webhook")
	}

	return entityPointer, nil
}
