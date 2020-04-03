package steps

// TODO: consider renaming package to install

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/sirupsen/logrus"
	"time"
)

// TODO: maybe move this to Runtime
// TODO: where to put processing result

const (
	defaultDelay = 2 * time.Second
)

func NewStepsExecutor(session dbsession.ReadWriteSession, operation model.OperationType, steps map[model.OperationStage]installation.Step) *Executor {
	return &Executor{
		dbSession:         session,
		installationSteps: steps,
		operation:         operation,
		log:               logrus.WithFields(logrus.Fields{"Component": "StepsExecutor", "OperationType": operation}),
	}
}

type Executor struct {
	dbSession         dbsession.ReadWriteSession
	installationSteps map[model.OperationStage]installation.Step
	operation         model.OperationType

	log logrus.FieldLogger
}

func (e *Executor) Execute(operationID string) installation.ProcessingResult {

	log := e.log.WithField("OperationId", operationID)

	// Get Operation
	operation, err := e.dbSession.GetOperation(operationID)
	if err != nil {
		log.Errorf("error getting operation while processing it: %s", err.Error())
		return installation.ProcessingResult{Requeue: true, Delay: defaultDelay}
	}

	log = log.WithField("RuntimeId", operation.ClusterID)

	if operation.State != model.InProgress {
		log.Infof("Operation not InProgress. State: %s", operation.State)
		return installation.ProcessingResult{Requeue: false}
	}

	cluster, err := e.dbSession.GetCluster(operation.ClusterID)
	if err != nil {
		log.Errorf("error getting cluster while processing operation: %s", err.Error())
		return installation.ProcessingResult{Requeue: true, Delay: defaultDelay}
	}

	if operation.Type == e.operation {
		requeue, delay, err := e.processInstallation(operation, cluster, log)
		if err != nil {
			nonRecoverable := installation.NonRecoverableError{}
			if errors.As(err, &nonRecoverable) {
				log.Errorf("unrecoverable error occurred while processing operation: %s", err.Error())
				err := e.dbSession.UpdateOperationState(operation.ID, nonRecoverable.Error(), model.Failed, time.Now()) // TODO: Align message
				if err != nil {
					panic(err) // TODO handle
				}
			}

			return installation.ProcessingResult{Requeue: true, Delay: defaultDelay}
		}

		return installation.ProcessingResult{Requeue: requeue, Delay: delay}
	}

	return installation.ProcessingResult{
		Requeue: false,
		Delay:   0,
	}
}

func (e *Executor) processInstallation(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (bool, time.Duration, error) {

	step, found := e.installationSteps[operation.Stage]
	if !found {
		return false, 0, installation.NewNonRecoverableError(fmt.Errorf("error: step %s not found in installation steps", operation.Stage))
	}

	for operation.Stage != model.FinishedStep {
		log := logger.WithField("Step", step.Name())
		log.Infof("Starting step")

		if e.timeoutReached(operation, step.TimeLimit()) {
			log.Errorf("Timeout reached for operation")
			return false, 0, installation.NewNonRecoverableError(fmt.Errorf("error: timeout while processing operation"))
		}

		result, err := step.Run(operation, cluster, log)
		if err != nil {
			log.Errorf("Step failed: %s", err.Error())
			return false, 0, err
		}

		if result.Step == model.FinishedStep {
			log.Infof("Finished processing operation. Setting operation to succeeded")
			err := e.dbSession.TransitionOperation(operation.ID, "Provisioning steps finished", model.FinishedStep, time.Now())
			if err != nil {
				panic(err) // TODO handle
			}
			break
		}

		if result.Step != step.Name() {
			transitionTime := time.Now()
			err = e.dbSession.TransitionOperation(operation.ID, "Operation in progress", result.Step, transitionTime) // TODO: Align message
			if err != nil {
				panic(err) // TODO handle
			}
			step = e.installationSteps[result.Step]
			operation.Stage = result.Step
			operation.LastTransition = &transitionTime
			log.Infof("Step finished")
		}

		if result.Delay > 0 {
			return true, result.Delay, nil
		}
	}

	// TODO: are we gaurenteed that it finished? Consider doing it in if statement above
	err := e.dbSession.UpdateOperationState(operation.ID, "Operation succeeded", model.Succeeded, time.Now()) // TODO: Align message
	if err != nil {
		panic(err) // TODO: handle
	}

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
