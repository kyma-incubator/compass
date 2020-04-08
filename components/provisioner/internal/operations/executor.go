package operations

import (
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/sirupsen/logrus"
)

const (
	defaultDelay = 2 * time.Second
)

func NewStepsExecutor(
	session dbsession.ReadWriteSession,
	operation model.OperationType,
	stages map[model.OperationStage]Stage,
	failureHandler FailureHandler) *StagesExecutor {
	return &StagesExecutor{
		dbSession:      session,
		stages:         stages,
		operation:      operation,
		failureHandler: failureHandler,
		log:            logrus.WithFields(logrus.Fields{"Component": "StepsExecutor", "OperationType": operation}),
	}
}

type StagesExecutor struct {
	dbSession      dbsession.ReadWriteSession
	stages         map[model.OperationStage]Stage
	operation      model.OperationType
	failureHandler FailureHandler

	log logrus.FieldLogger
}

func (e *StagesExecutor) Execute(operationID string) ProcessingResult {

	log := e.log.WithField("OperationId", operationID)

	// Get Operation
	operation, err := e.dbSession.GetOperation(operationID)
	if err != nil {
		log.Errorf("error getting operation while processing it: %s", err.Error())
		return ProcessingResult{Requeue: true, Delay: defaultDelay}
	}

	log = log.WithField("RuntimeId", operation.ClusterID)

	if operation.State != model.InProgress {
		log.Infof("Operation not InProgress. State: %s", operation.State)
		return ProcessingResult{Requeue: false}
	}

	cluster, err := e.dbSession.GetCluster(operation.ClusterID)
	if err != nil {
		log.Errorf("error getting cluster while processing operation: %s", err.Error())
		return ProcessingResult{Requeue: true, Delay: defaultDelay}
	}

	if operation.Type == e.operation {
		requeue, delay, err := e.processInstallation(operation, cluster, log)
		if err != nil {
			nonRecoverable := NonRecoverableError{}
			if errors.As(err, &nonRecoverable) {
				log.Errorf("unrecoverable error occurred while processing operation: %s", err.Error())
				e.handleOperationFailure(operation, cluster, log)
				e.updateOperationStatus(log, operation.ID, nonRecoverable.Error(), model.Failed, time.Now()) // TODO: what should be the massega?
				return ProcessingResult{Requeue: false}
			}

			return ProcessingResult{Requeue: true, Delay: defaultDelay}
		}

		return ProcessingResult{Requeue: requeue, Delay: delay}
	}

	return ProcessingResult{
		Requeue: false,
		Delay:   0,
	}
}

func (e *StagesExecutor) processInstallation(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (bool, time.Duration, error) {

	step, found := e.stages[operation.Stage]
	if !found {
		return false, 0, NewNonRecoverableError(fmt.Errorf("error: step %s not found in installation steps", operation.Stage))
	}

	for operation.Stage != model.FinishedStage {
		log := logger.WithField("Stage", step.Name())
		log.Infof("Starting step")

		if e.timeoutReached(operation, step.TimeLimit()) {
			log.Errorf("Timeout reached for operation")
			return false, 0, NewNonRecoverableError(fmt.Errorf("error: timeout while processing operation"))
		}

		result, err := step.Run(cluster, operation, log)
		if err != nil {
			log.Errorf("Stage failed: %s", err.Error())
			return false, 0, err
		}

		if result.Stage == model.FinishedStage {
			log.Infof("Finished processing operation. Setting operation to succeeded")
			err := e.dbSession.TransitionOperation(operation.ID, "Provisioning steps finished", model.FinishedStage, time.Now())
			if err != nil {
				panic(err) // TODO handle
			}
			break
		}

		if result.Stage != step.Name() {
			transitionTime := time.Now()
			err = e.dbSession.TransitionOperation(operation.ID, "Operation in progress", result.Stage, transitionTime) // TODO: Align message
			if err != nil {
				panic(err) // TODO handle
			}
			step = e.stages[result.Stage]
			operation.Stage = result.Stage
			operation.LastTransition = &transitionTime
			log.Infof("Stage finished")
		}

		if result.Delay > 0 {
			return true, result.Delay, nil
		}
	}

	// TODO: are we gaurenteed that it finished? Consider doing it in if statement above
	// TODO: update with retries

	e.updateOperationStatus(logger, operation.ID, "Operation succeeded", model.Succeeded, time.Now())
	return false, 0, nil
}

func (e *StagesExecutor) timeoutReached(operation model.Operation, timeout time.Duration) bool {

	lastTimestamp := operation.StartTimestamp
	if operation.LastTransition != nil {
		lastTimestamp = *operation.LastTransition
	}

	timePassed := time.Now().Sub(lastTimestamp)

	return timePassed > timeout
}

func (e *StagesExecutor) handleOperationFailure(operation model.Operation, cluster model.Cluster, log logrus.FieldLogger) {
	err := retry.Do(func() error {
		return e.failureHandler.HandleFailure(operation, cluster)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("error handling operation failure operation failure: %s", err.Error())
	}
}

func (e *StagesExecutor) updateOperationStatus(log logrus.FieldLogger, id, message string, state model.OperationState, t time.Time) {
	err := retry.Do(func() error {
		return e.dbSession.UpdateOperationState(id, message, state, t)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("Failed to set operation status to %s: %s", state, err.Error())
	}
}
