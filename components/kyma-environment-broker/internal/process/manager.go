package process

import (
	"sort"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
)

type Step interface {
	Name() string
	Run(operation internal.ProvisioningOperation, logger *logrus.Entry) (internal.ProvisioningOperation, time.Duration, error)
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
	operation, err := m.operationStorage.GetProvisioningOperationByID(operationID)
	if err != nil {
		return 3 * time.Second, nil
	}

	var when time.Duration
	processedOperation := *operation
	logOperation := m.log.WithFields(logrus.Fields{"operation": operationID})

	logOperation.Info("Start steps")
	for _, weightStep := range m.sortWeight() {
		steps := m.steps[weightStep]
		for _, step := range steps {
			logStep := logOperation.WithFields(logrus.Fields{"step": step.Name()})
			logStep.Infof("Start step")

			processedOperation, when, err = step.Run(processedOperation, logStep)
			if err != nil {
				logStep.Errorf("Step failed: %s", err)
				return 0, err
			}
			if when == 0 {
				logStep.Info("Step successful")
				continue
			}

			logStep.Infof("Step will be repeated in %s ...", when)
			return when, nil
		}
	}

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
