package provisioning

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	hyperscalerMocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestResolveCredentialsStepHappyPath_Run(t *testing.T) {
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

	assert.NoError(t, err)

	pp, err := operation.GetProvisioningParameters()

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Empty(t, operation.State)
	require.NotNil(t, pp.Parameters.TargetSecret)
	assert.Equal(t, "gardener-secret-gcp", *pp.Parameters.TargetSecret)
}

func TestResolveCredentialsStepFailedAfterRetry_Run(t *testing.T) {
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

	accountProviderMock.On("GardenerCredentials", hyperscaler.GCP, statusGlobalAccountID).Return(hyperscaler.Credentials{}, errors.New("Failed!"))

	step := NewResolveCredentialsStep(memoryStorage.Operations(), accountProviderMock)

	operation.UpdatedAt = time.Now()

	// when
	operation, repeat, err := step.Run(operation, log)

	assert.NoError(t, err)

	pp, err := operation.GetProvisioningParameters()
	assert.NoError(t, err)

	// then
	assert.NoError(t, err)
	assert.Equal(t, 2*time.Second, repeat)
	assert.Nil(t, pp.Parameters.TargetSecret)
	assert.Empty(t, operation.State)

	time.Sleep(repeat)
	operation, repeat, err = step.Run(operation, log)

	pp, err = operation.GetProvisioningParameters()
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, 2*time.Second, repeat)
	assert.Empty(t, operation.State)
	assert.Nil(t, pp.Parameters.TargetSecret)

	time.Sleep(repeat)
	operation, repeat, err = step.Run(operation, log)
	assert.Error(t, err)

	pp, err = operation.GetProvisioningParameters()
	assert.NoError(t, err)

	assert.Equal(t, domain.Failed, operation.State)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Nil(t, pp.Parameters.TargetSecret)
}
