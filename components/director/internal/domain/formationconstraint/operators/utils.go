package operators

import (
	"context"
	"runtime/debug"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type entities interface {
	*model.FormationAssignment | *model.Webhook
}

// RetrieveEntityPointerFromMemoryAddress converts the provided memory address, in the form of an integer, to a provided entity type pointer structure.
// It's important the provided memory address to stores information about entity that matches the provided entity type, otherwise the result could be very abnormal
func RetrieveEntityPointerFromMemoryAddress[entity entities](ctx context.Context, e entity, memoryAddress uintptr) (entity, error) {
	if memoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting memory address: %d", memoryAddress)
			debug.PrintStack()
		}
	}()
	entityPointer := (e)(unsafe.Pointer(memoryAddress))

	return entityPointer, nil
}
