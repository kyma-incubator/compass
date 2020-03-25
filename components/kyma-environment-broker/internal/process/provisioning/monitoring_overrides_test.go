package provisioning

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSetupMonitoringOverridesStepHappyPath_Run(t *testing.T) {
	// given
	inputCreatorMock := &automock.ProvisionInputCreator{}
	//expCredentialsValues := []*gqlschema.ConfigEntryInput{
	//	{
	//		Key: "resourceSelector.namespaces",
	//		Value: "kyma-system,istio-system,knative-eventing,knative-serving,kyma-integration,kube-system",
	//		Secret: ptr.Bool(true),
	//	},
	//	{
	//		Key:    "grafana.kyma.console.enabled",
	//		Value:  "false",
	//		Secret: ptr.Bool(true),
	//	},
	//	{
	//		Key:    "grafana.env.GF_USERS_AUTO_ASSIGN_ORG_ROLE",
	//		Value:  "Admin",
	//		Secret: ptr.Bool(true),
	//	},
	//	{
	//		Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_SCOPES",
	//		Value:  "openid email",
	//		Secret: ptr.Bool(true),
	//	},
	//	{
	//		Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_TOKEN_URL",
	//		Value:  "https://kyma.blah.com/oauth2/token",
	//		Secret: ptr.Bool(true),
	//	},
	//	{
	//		Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_AUTH_URL",
	//		Value:  "https://kyma.foo.com/oauth2/token",
	//		Secret: ptr.Bool(true),
	//	},{
	//		Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_API_URL",
	//		Value:  "https://kyma.bar.com/oauth2/token",
	//		Secret: ptr.Bool(true),
	//	},
	//}
	//inputCreatorMock.On("AppendOverrides", "monitoring", expCredentialsValues).
	//	Return(nil).Once()
	expCredentialsValues := []*gqlschema.ConfigEntryInput{
		{
			Key:    "test7",
			Value:  "test7abc",
			Secret: ptr.Bool(true),
		},
	}
	//inputCreatorMock.On("setupMonitoringOverride").Return(expCredentialsValues).Once()

	inputCreatorMock.On("AppendOverrides", "monitoring",expCredentialsValues ).Return(nil).Once()


	operation := internal.ProvisioningOperation{
		InputCreator: inputCreatorMock,
	}

	memoryStorage := storage.NewMemoryStorage()
	step := NewMonitoringOverrideStep(memoryStorage.Operations())

	// when
	_, retryTime, err := step.Run(operation, NewLogDummy())

	// then
	require.NoError(t, err)

	assert.Zero(t, retryTime)
	assert.Equal(t, time.Duration(0), retryTime)
	//inputCreatorMock.AssertExpectatixons(t)
}
