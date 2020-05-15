package provisioning

import (
	"sort"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"context"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type Step interface {
	Name() string
	Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error)
}

type Manager struct {
	log              logrus.FieldLogger
	steps            map[int][]Step
	operationStorage storage.Operations

	publisher event.Publisher
}

func NewManager(storage storage.Operations, pub event.Publisher, logger logrus.FieldLogger) *Manager {
	return &Manager{
		log:              logger,
		operationStorage: storage,
		steps:            make(map[int][]Step, 0),
		publisher:        pub,
	}
}

func (m *Manager) InitStep(step Step) {
	m.AddStep(0, step)
}

func (m *Manager) AddStep(weight int, step Step) {
	if weight <= 0 {
		weight = 1
	}
	m.steps[weight] = append(m.steps[weight], step)
}

func (m *Manager) runStep(step Step, operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	start := time.Now()
	processedOperation, when, err := step.Run(operation, logger)
	m.publisher.Publish(context.TODO(), process.ProvisioningStepProcessed{
		OldOperation: operation,
		Operation:    processedOperation,
		StepProcessed: process.StepProcessed{
			StepName: step.Name(),
			Duration: time.Since(start),
			When:     when,
			Error:    err,
		},
	})
	return processedOperation, when, err
}

func (m *Manager) Execute(operationID string) (time.Duration, error) {
	operation, err := m.operationStorage.GetProvisioningOperationByID(operationID)
	if err != nil {
		m.log.Errorf("Cannot fetch operation from storage: %s", err)
		return 3 * time.Second, nil
	}

	var when time.Duration
	processedOperation := *operation
	logOperation := m.log.WithFields(logrus.Fields{"operation": operationID, "instanceID": operation.InstanceID})

	logOperation.Info("Start process operation steps")
	for _, weightStep := range m.sortWeight() {
		steps := m.steps[weightStep]
		for _, step := range steps {
			logStep := logOperation.WithField("step", step.Name())
			logStep.Infof("Start step")

			processedOperation, when, err = m.runStep(step, processedOperation, logStep)
			if err != nil {
				logStep.Errorf("Process operation failed: %s", err)
				return 0, err
			}
			if processedOperation.State != domain.InProgress {
				logrus.Infof("Operation %q got status %s. Process finished.", operation.ID, processedOperation.State)
				return 0, nil
			}
			if when == 0 {
				logStep.Info("Process operation successful")
				continue
			}

			logStep.Infof("Process operation will be repeated in %s ...", when)
			return when, nil
		}
	}

	logrus.Infof("Operation %q got status %s. Process finished.", operation.ID, processedOperation.State)
	return 0, nil
}

func (m *Manager) sortWeight() []int {
	var weight []int
	for w := range m.steps {
		weight = append(weight, w)
	}
	sort.Ints(weight)

	return weight
}
