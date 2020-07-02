package deprovisioning

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"

	"context"
	"sync"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/event"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	operationIDSuccess = "5b954fa8-fc34-4164-96e9-49e3b6741278"
	operationIDFailed  = "69b8ee2b-5c21-4997-9070-4fd356b24c46"
	operationIDRepeat  = "ca317a1e-ddab-44d2-b2ba-7bbd9df9066f"
)

func TestManager_Execute(t *testing.T) {
	for name, tc := range map[string]struct {
		operationID            string
		expectedError          bool
		expectedRepeat         time.Duration
		expectedDesc           string
		expectedNumberOfEvents int
	}{
		"operation successful": {
			operationID:            operationIDSuccess,
			expectedError:          false,
			expectedRepeat:         time.Duration(0),
			expectedDesc:           "init one two final",
			expectedNumberOfEvents: 4,
		},
		"operation failed": {
			operationID:            operationIDFailed,
			expectedError:          true,
			expectedNumberOfEvents: 1,
		},
		"operation repeated": {
			operationID:            operationIDRepeat,
			expectedError:          false,
			expectedRepeat:         time.Duration(10),
			expectedDesc:           "init",
			expectedNumberOfEvents: 1,
		},
	} {
		t.Run(name, func(t *testing.T) {
			// given
			log := logrus.New()
			memoryStorage := storage.NewMemoryStorage()
			operations := memoryStorage.Operations()
			err := operations.InsertDeprovisioningOperation(fixOperation(tc.operationID))
			assert.NoError(t, err)

			sInit := testStep{t: t, name: "init", storage: operations}
			s1 := testStep{t: t, name: "one", storage: operations}
			s2 := testStep{t: t, name: "two", storage: operations}
			sFinal := testStep{t: t, name: "final", storage: operations}

			eventBroker := event.NewPubSub()
			eventCollector := &collectingEventHandler{}
			eventBroker.Subscribe(process.DeprovisioningStepProcessed{}, eventCollector.OnEvent)

			manager := NewManager(operations, eventBroker, log)
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

				operation, err := operations.GetOperationByID(tc.operationID)
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDesc, strings.Trim(operation.Description, " "))
			}
			assert.NoError(t, wait.PollImmediate(20*time.Millisecond, 2*time.Second, func() (bool, error) {
				return len(eventCollector.Events) == tc.expectedNumberOfEvents, nil
			}))
		})
	}
}

func fixOperation(ID string) internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
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

func (ts *testStep) Run(operation internal.DeprovisioningOperation, logger logrus.FieldLogger) (internal.DeprovisioningOperation, time.Duration, error) {
	logger.Infof("inside %s step", ts.name)

	operation.Description = fmt.Sprintf("%s %s", operation.Description, ts.name)
	updated, err := ts.storage.UpdateDeprovisioningOperation(operation)
	if err != nil {
		ts.t.Error(err)
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

type collectingEventHandler struct {
	mu     sync.Mutex
	Events []interface{}
}

func (h *collectingEventHandler) OnEvent(ctx context.Context, ev interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Events = append(h.Events, ev)
	return nil
}
