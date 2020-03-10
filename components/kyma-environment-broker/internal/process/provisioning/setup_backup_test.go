package provisioning

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	hyperscalerMocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetupBackupStepHappyPath_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()

	operation := fixOperationRuntimeStatus(t)
	err := memoryStorage.Operations().InsertProvisioningOperation(operation)
	assert.NoError(t, err)

	instance := fixInstanceRuntimeStatus()
	err = memoryStorage.Instances().Insert(instance)
	assert.NoError(t, err)

	accountProviderMock := &hyperscalerMocks.AccountProvider{}
	
	accountProviderMock.On("GardenerCredentials", hyperscaler.GCP, statusGlobalAccountID).Return(hyperscaler.Credentials{
		CredentialName:  "gardener-secret-gcp",
		HyperscalerType: "gcp",
		TenantName:      statusGlobalAccountID,
		CredentialData:  map[string][]byte{},
	}, nil)
	step := NewSetupBackupStep(memoryStorage.Operations(), accountProviderMock)

	// when
	operation, repeat , err := step.Run(operation, log)
	assert.NoError(t, err)

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Empty(t, operation.State)
	//require.NotNil(t, pp.Parameters.TargetSecret)
	//assert.Equal(t, "gardener-secret-gcp", *pp.Parameters.TargetSecret)
}
