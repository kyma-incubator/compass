package operations

import (
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	retry "github.com/avast/retry-go"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/sirupsen/logrus"
)

const (
	defaultDelay         = 2 * time.Second
	backOffDirectorDelay = 1 * time.Second
)

func NewExecutor(
	session dbsession.ReadWriteSession,
	operation model.OperationType,
	stages map[model.OperationStage]Step,
	failureHandler FailureHandler,
	directorClient director.DirectorClient) *Executor {

	return &Executor{
		dbSession:      session,
		stages:         stages,
		operation:      operation,
		failureHandler: failureHandler,
		log:            logrus.WithFields(logrus.Fields{"Component": "Executor", "OperationType": operation}),
		directorClient: directorClient,
	}
}

type Executor struct {
	dbSession      dbsession.ReadWriteSession
	stages         map[model.OperationStage]Step
	operation      model.OperationType
	failureHandler FailureHandler
	directorClient director.DirectorClient

	log logrus.FieldLogger
}

func (e *Executor) Execute(operationID string) ProcessingResult {

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
		requeue, delay, err := e.process(operation, cluster, log)
		if err != nil {
			nonRecoverable := NonRecoverableError{}
			if errors.As(err, &nonRecoverable) {
				log.Errorf("unrecoverable error occurred while processing operation: %s", err.Error())
				e.handleOperationFailure(operation, cluster, log)
				e.updateOperationStatus(log, operation.ID, nonRecoverable.Error(), model.Failed, time.Now())
				e.setRuntimeStatusCondition(log, cluster.ID, cluster.Tenant)

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

func (e *Executor) process(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (bool, time.Duration, error) {

	step, found := e.stages[operation.Stage]
	if !found {
		return false, 0, NewNonRecoverableError(fmt.Errorf("error: step %s not found in installation stages", operation.Stage))
	}

	for operation.Stage != model.FinishedStage {
		log := logger.WithField("Stage", step.Name())
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

		if result.Stage != step.Name() {
			transitionTime := time.Now()
			e.updateOperationStage(log, operation.ID, fmt.Sprintf("Operation in progress. Stage %s", result.Stage), result.Stage, transitionTime)
			step = e.stages[result.Stage]
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

func (e *Executor) timeoutReached(operation model.Operation, timeout time.Duration) bool {

	lastTimestamp := operation.StartTimestamp
	if operation.LastTransition != nil {
		lastTimestamp = *operation.LastTransition
	}

	timePassed := time.Now().Sub(lastTimestamp)

	return timePassed > timeout
}

func (e *Executor) handleOperationFailure(operation model.Operation, cluster model.Cluster, log logrus.FieldLogger) {
	err := retry.Do(func() error {
		return e.failureHandler.HandleFailure(operation, cluster)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("error handling operation failure operation failure: %s", err.Error())
	}
}

func (e *Executor) updateOperationStatus(log logrus.FieldLogger, id, message string, state model.OperationState, t time.Time) {
	err := retry.Do(func() error {
		return e.dbSession.UpdateOperationState(id, message, state, t)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("Failed to set operation status to %s: %s", state, err.Error())
	}
}

func (e *Executor) setRuntimeStatusCondition(log logrus.FieldLogger, id, tenant string) {
	err := retry.Do(func() error {
		return e.directorClient.SetRuntimeStatusCondition(id, graphql.RuntimeStatusConditionFailed, tenant)
	}, retry.Attempts(5), retry.Delay(backOffDirectorDelay), retry.DelayType(retry.BackOffDelay))
	if err != nil {
		log.Errorf("failed to set runtime %s status condition: %s", graphql.RuntimeStatusConditionFailed.String(), err.Error())
	}
}

func (e *Executor) updateOperationStage(log logrus.FieldLogger, id, message string, stage model.OperationStage, t time.Time) {
	err := retry.Do(func() error {
		return e.dbSession.TransitionOperation(id, message, stage, t)
	}, retry.Attempts(5))
	if err != nil {
		log.Errorf("Failed to modify operation stage to %s: %s", stage, err.Error())
	}
}
