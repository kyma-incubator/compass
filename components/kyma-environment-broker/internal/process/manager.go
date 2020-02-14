package process

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

type Step interface {
	Run(operation *internal.ProvisioningOperation) (error, time.Duration)
}

type Manager struct {
	operationStorage storage.Operations
	steps            []Step
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) AddStep(step Step) {
	m.steps = append(m.steps, step)
}

func (m *Manager) Execute(operationID string) (error, time.Duration) {
	operation, err := m.operationStorage.GetProvisioningOperationByID(operationID)
	if err != nil {
		return nil, 10 * time.Minute
	}

	for _, step := range m.steps {
		err, when := step.Run(operation)
		if err != nil {
			return err, 0
		}
		if when == 0 {
			continue
		}
		return nil, when
	}

	return nil, 0
}
