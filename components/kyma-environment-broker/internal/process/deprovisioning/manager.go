package deprovisioning

import (
	"sort"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
)

type Step interface {
	Name() string
	Run(operation internal.DeprovisioningOperation, logger logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error)
}

type Manager struct {
	log              *logrus.Logger
	steps            map[int][]Step
	operationStorage storage.Operations
}

func NewManager(storage storage.Operations, logger *logrus.Logger) *Manager {
	return &Manager{
		log:              logger,
		operationStorage: storage,
		steps:            make(map[int][]Step, 0),
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

func (m *Manager) Execute(operationID string) (time.Duration, error) {
	op, err := m.operationStorage.GetDeprovisioningOperationByID(operationID)
	if err != nil {
		m.log.Errorf("Cannot fetch operation from storage: %s", err)
		return 3 * time.Second, nil
	}
	operation := *op

	var when time.Duration
	logOperation := m.log.WithFields(logrus.Fields{"operation": operationID, "instanceID": operation.InstanceID})

	logOperation.Info("Start process operation steps")
	for _, weightStep := range m.sortWeight() {
		steps := m.steps[weightStep]
		for _, step := range steps {
			logStep := logOperation.WithField("step", step.Name())
			logStep.Infof("Start step")

			operation, when, err = step.Run(operation, logStep)
			if err != nil {
				logStep.Errorf("Process operation failed: %s", err)
				return 0, err
			}
			if operation.State != domain.InProgress {
				logStep.Infof("Operation %q got status %s. Process finished.", operation.ID, operation.State)
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

	logrus.Infof("Operation %q got status %s. Process finished.", operation.ID, operation.State)
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
