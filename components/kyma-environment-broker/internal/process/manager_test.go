package process

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	operationIDSuccess = "5b954fa8-fc34-4164-96e9-49e3b6741278"
	operationIDFailed  = "69b8ee2b-5c21-4997-9070-4fd356b24c46"
	operationIDRepeat  = "ca317a1e-ddab-44d2-b2ba-7bbd9df9066f"
)

func TestManager_Execute(t *testing.T) {
	for name, tc := range map[string]struct {
		operationID    string
		expectedError  bool
		expectedRepeat time.Duration
		expectedDesc   string
	}{
		"operation successful": {
			operationID:    operationIDSuccess,
			expectedError:  false,
			expectedRepeat: time.Duration(0),
			expectedDesc:   "init one two final",
		},
		"operation failed": {
			operationID:   operationIDFailed,
			expectedError: true,
		},
		"operation repeated": {
			operationID:    operationIDRepeat,
			expectedError:  false,
			expectedRepeat: time.Duration(10),
			expectedDesc:   "init",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// given
			log := logrus.New()
			memoryStorage := storage.NewMemoryStorage()
			err := memoryStorage.Operations().InsertProvisioningOperation(fixOperation(tc.operationID))
			assert.NoError(t, err)

			sInit := testStep{name: "init", storage: memoryStorage.Operations()}
			s1 := testStep{name: "one", storage: memoryStorage.Operations()}
			s2 := testStep{name: "two", storage: memoryStorage.Operations()}
			sFinal := testStep{name: "final", storage: memoryStorage.Operations()}

			manager := NewManager(memoryStorage.Operations(), log)
			manager.InitStep(&sInit)

			manager.AddStep(2, &sFinal)
			manager.AddStep(1, &s1)
			manager.AddStep(1, &s2)

			// when
			repeat, err := manager.Execute(tc.operationID)

			// then
			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRepeat, repeat)

				operation, err := memoryStorage.Operations().GetOperation(tc.operationID)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDesc, strings.Trim(operation.Description, " "))
			}
		})
	}
}

func fixOperation(ID string) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:          ID,
			State:       domain.InProgress,
			InstanceID:  "fea2c1a1-139d-43f6-910a-a618828a79d5",
			Description: "",
		},
	}
}

type testStep struct {
	t       *testing.T
	name    string
	storage storage.Operations
}

func (ts *testStep) Name() string {
	return ts.name
}

func (ts *testStep) Run(operation internal.ProvisioningOperation, logger logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	logger.Infof("inside %s step", ts.name)

	operation.Description = fmt.Sprintf("%s %s", operation.Description, ts.name)
	updated, err := ts.storage.UpdateProvisioningOperation(operation)
	if err != nil {
		ts.t.Errorf("cannot update operation: %s", err)
	}

	switch operation.ID {
	case operationIDFailed:
		return *updated, 0, fmt.Errorf("operation %s failed", operation.ID)
	case operationIDRepeat:
		return *updated, time.Duration(10), nil
	default:
		return *updated, 0, nil
	}
}
