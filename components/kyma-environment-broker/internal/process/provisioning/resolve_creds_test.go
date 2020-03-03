package provisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	hyperscalerMocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/mocks"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestResolveCredentialsStep_Run(t *testing.T) {
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

	step := NewResolveCredentialsStep(memoryStorage.Operations(), accountProviderMock)

	// when
	operation, repeat, err := step.Run(operation, log)

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Equal(t, domain.Succeeded, operation.State)
	assert.Equal(t, "gardener-secret-gcp", operation.TargetSecret)
}
