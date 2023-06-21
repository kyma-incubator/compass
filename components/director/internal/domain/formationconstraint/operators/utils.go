package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// RetrieveFormationAssignmentPointer converts the provided uninterpreted memory address in form of an integer back to the model.FormationAssignment pointer structure
func RetrieveFormationAssignmentPointer(ctx context.Context, jointPointDetailsAssignmentMemoryAddress uintptr) (*model.FormationAssignment, error) {
	if jointPointDetailsAssignmentMemoryAddress == 0 { // the default value of uintptr is 0
		return nil, errors.New("The joint point details' assignment memory address cannot be 0")
	}

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panicf("A panic occurred while converting joint point details' assignment address: %d to type: %T", jointPointDetailsAssignmentMemoryAddress, &model.FormationAssignment{})
			debug.PrintStack()
		}
	}()
	jointPointAssignmentPointer := (*model.FormationAssignment)(unsafe.Pointer(jointPointDetailsAssignmentMemoryAddress))

	return jointPointAssignmentPointer, nil
}

func (e *ConstraintEngine) transaction(ctx context.Context, dbCall func(ctxWithTransact context.Context) error) error {
	tx, err := e.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to begin DB transaction")
		return err
	}
	defer e.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = dbCall(ctx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit database transaction")
		return err
	}
	return nil
}

func (e *ConstraintEngine) setAssignmentToErrorState(ctx context.Context, assignment *model.FormationAssignment, errorMessage string) error {
	assignment.State = string(model.CreateErrorAssignmentState)
	assignmentError := formationassignment.AssignmentErrorWrapper{Error: formationassignment.AssignmentError{
		Message:   errorMessage,
		ErrorCode: formationassignment.ClientError,
	}}
	marshaled, err := json.Marshal(assignmentError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for assignment with ID %q", assignment.ID)
	}
	assignment.Value = marshaled
	if err := e.formationAssignmentRepo.Update(ctx, assignment); err != nil {
		return errors.Wrapf(err, "While updating formation assignment with id %q", assignment.ID)
	}
	log.C(ctx).Infof("Assignment with ID %s set to state %s", assignment.ID, assignment.State)
	return nil
}
