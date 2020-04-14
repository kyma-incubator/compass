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
	stagesSteps map[model.OperationStage]Step,
	failureHandler FailureHandler) *StepsExecutor {
	return &StepsExecutor{
		dbSession:      session,
		stagesSteps:    stagesSteps,
		operation:      operation,
		failureHandler: failureHandler,
		log:            logrus.WithFields(logrus.Fields{"Component": "StepsExecutor", "OperationType": operation}),
	}
}

type StepsExecutor struct {
	dbSession      dbsession.ReadWriteSession
	stagesSteps    map[model.OperationStage]Step
	operation      model.OperationType
	failureHandler FailureHandler

	log logrus.FieldLogger
}

func (e *StepsExecutor) Execute(operationID string) ProcessingResult {

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
				e.updateOperationStatus(log, operation.ID, nonRecoverable.Error(), model.Failed, time.Now())
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

func (e *StepsExecutor) processInstallation(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (bool, time.Duration, error) {

	step, found := e.stagesSteps[operation.Stage]
	if !found {
		return false, 0, NewNonRecoverableError(fmt.Errorf("error: step %s not found in installation steps", operation.Stage))
	}

	for operation.Stage != model.FinishedStage {
		log := logger.WithField("Stage", step.Stage())
		log.Infof("Starting processing")

		if e.timeoutReached(operation, step.TimeLimit()) {
			log.Errorf("Timeout reached for operation")
			return false, 0, NewNonRecoverableError(fmt.Errorf("error: timeout while processing operation"))
		}

		result, err := step.Run(cluster, operation, log)
		if err != nil {
			log.Errorf("error while processing operation, stage failed: %s", err.Error())
			return false, 0, err
		}

		if result.Stage == model.FinishedStage {
			log.Infof("Finished processing operation")
			e.updateOperationStage(log, operation.ID, "Provisioning steps finished", model.FinishedStage, time.Now())
			break
		}

		if result.Stage != step.Stage() {
			transitionTime := time.Now()
			e.updateOperationStage(log, operation.ID, fmt.Sprintf("Operation in progress. Stage %s", result.Stage), result.Stage, transitionTime)
			step = e.stagesSteps[result.Stage]
			operation.Stage = result.Stage
			operation.LastTransition = &transitionTime
			log.Infof("Stage completed")
		}

		if result.Delay > 0 {
			return true, result.Delay, nil
		}
	}

	logger.Infof("Setting operation to succeeded")
	e.updateOperationStatus(logger, operation.ID, "Operation succeeded", model.Succeeded, time.Now())

	return false, 0, nil
}

func (e *StepsExecutor) timeoutReached(operation model.Operation, timeout time.Duration) bool {

	lastTimestamp := operation.StartTimestamp
	if operation.LastTransition != nil {
		lastTimestamp = *operation.LastTransition
	}

	timePassed := time.Now().Sub(lastTimestamp)

	return timePassed > timeout
}

func (e *StepsExecutor) handleOperationFailure(operation model.Operation, cluster model.Cluster, log logrus.FieldLogger) {
	err := retry.Do(func() error {
		return e.failureHandler.HandleFailure(operation, cluster)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("error handling operation failure operation failure: %s", err.Error())
	}
}

func (e *StepsExecutor) updateOperationStatus(log logrus.FieldLogger, id, message string, state model.OperationState, t time.Time) {
	err := retry.Do(func() error {
		return e.dbSession.UpdateOperationState(id, message, state, t)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("Failed to set operation status to %s: %s", state, err.Error())
	}
}

func (e *StepsExecutor) updateOperationStage(log logrus.FieldLogger, id, message string, stage model.OperationStage, t time.Time) {
	err := retry.Do(func() error {
		return e.dbSession.TransitionOperation(id, message, stage, t)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("Failed to modify operation stage to %s: %s", stage, err.Error())
	}
}
