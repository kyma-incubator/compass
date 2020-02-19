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
	operationStorage storage.Operations
	steps            map[int]Step
}

func NewManager(logger *logrus.Logger) *Manager {
	return &Manager{
		log:   logger,
		steps: make(map[int]Step, 1),
	}
}

func (m *Manager) AddStep(order int, step Step) {
	m.steps[order] = step
}

func (m *Manager) Execute(operationID string) (time.Duration, error) {
	operation, err := m.operationStorage.GetProvisioningOperationByID(operationID)
	if err != nil {
		return 10 * time.Minute, nil
	}

	var when time.Duration
	processedOperation := *operation
	log := m.log.WithFields(logrus.Fields{"operation": operationID})

	var order []int
	for o := range m.steps {
		order = append(order, o)
	}
	sort.Ints(order)

	for _, orderStep := range order {
		step := m.steps[orderStep]
		log.WithFields(logrus.Fields{"step": step.Name()})
		log.Infof("Start step")

		processedOperation, when, err = step.Run(processedOperation, log)
		if err != nil {
			log.Errorf("Step failed: %s", err)
			return 0, err
		}
		if when == 0 {
			log.Info("Step successful")
			continue
		}

		log.Infof("Step will be repeated in %s ...", when)
		return when, nil
	}

	return 0, nil
}
