package deprovisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	provisionerAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRemoveRuntimeStep_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationRemoveRuntime()
	err := memoryStorage.Operations().InsertDeprovisioningOperation(operation)
	assert.NoError(t, err)

	err = memoryStorage.Instances().Insert(fixInstanceRuntimeStatus())
	assert.NoError(t, err)

	provisionerClient := &provisionerAutomock.Client{}
	provisionerClient.On("DeprovisionRuntime", fixGlobalAccountID, fixRuntimeID).Return(fixProvisionerOperationID, nil)

	step := NewRemoveRuntimeStep(memoryStorage.Operations(), memoryStorage.Instances(), provisionerClient)

	// when
	entry := log.WithFields(logrus.Fields{"step": "TEST"})
	result, repeat, err := step.Run(operation, entry)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 1*time.Second, repeat)
	assert.Equal(t, fixProvisionerOperationID, result.ProvisionerOperationID)

	instance, err := memoryStorage.Instances().GetByID(result.InstanceID)
	assert.NoError(t, err)
	assert.Equal(t, instance.RuntimeID, fixRuntimeID)
}

func fixOperationRemoveRuntime() internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
		Operation: internal.Operation{
			ID:          fixOperationID,
			InstanceID:  fixInstanceID,
			Description: "",
			UpdatedAt:   time.Now(),
		},
	}
}
