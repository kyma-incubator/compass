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

// RetrieveFormationAssignmentPointer converts the provided uninterpreted memory address in form of an integer back to the model.FormationAssignment pointer structure
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
