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

func TestLoggingOverridesStepHappyPath_Run(t *testing.T) {
	// given
	inputCreatorMock := &automock.ProvisionInputCreator{}
	expCredentialsValues := []*gqlschema.ConfigEntryInput{
		{
			Key: "conf.Input.Kubernetes_loki.exclude.namespaces",
			Value: "kube-node-lease,kube-public,kube-system,kyma-system,istio-system,kyma-installer,kyma-integration,knative-serving,knative-eventing",
			Secret: ptr.Bool(true)},
	}
	inputCreatorMock.On("AppendOverrides", "logging", expCredentialsValues).
		Return(nil).Once()

	operation := internal.ProvisioningOperation{
		InputCreator: inputCreatorMock,
	}

	memoryStorage := storage.NewMemoryStorage()
	step := NewLoggingOverrides(memoryStorage.Operations())

	// when
	gotOperation, retryTime, err := step.Run(operation, NewLogDummy())

	// then
	require.NoError(t, err)

	assert.Zero(t, retryTime)
	assert.Equal(t, operation, gotOperation)
	inputCreatorMock.AssertExpectations(t)
}
