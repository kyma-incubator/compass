package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupBackupStepHappyPath_Run(t *testing.T) {
	// given
	inputCreatorMock := &automock.ProvisionInputCreator{}
	expCredentialsValues := []*gqlschema.ConfigEntryInput{
		{Key: "configuration.provider", Value: "azure", Secret: ptr.Bool(true)},
	}
	inputCreatorMock.On("AppendOverrides", "backup-init", expCredentialsValues).
		Return(nil).Once()

	operation := internal.ProvisioningOperation{
		InputCreator: inputCreatorMock,
	}

	memoryStorage := storage.NewMemoryStorage()
	step := NewSetupBackupStep(memoryStorage.Operations())

	// when
	gotOperation, retryTime, err := step.Run(operation, NewLogDummy())

	require.NoError(t, err)

	assert.Zero(t, retryTime)
	assert.Equal(t, operation, gotOperation)
	inputCreatorMock.AssertExpectations(t)
}
