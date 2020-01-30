package runtimes

import (
	"errors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/runtimes/clientbuilder/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestProvider_CreateConfigMapForRuntime(t *testing.T) {
	connectorURL := "https://kyma.cx/connector/graphql"
	runtimeID := "123-123-456"
	tenant := "tenant"
	token := "shdfv7123ygfbw832b"
	kubeconfig := "some Kubeconfig"

	namespace := "default"

	t.Run("Should create secret", func(t *testing.T) {
		//given
		config := RuntimeConfig{
			ConnectorURL: connectorURL,
			RuntimeID:    runtimeID,
			Tenant:       tenant,
			OneTimeToken: token,
		}

		builder := &mocks.ConfigMapClientBuilder{}

		configProvider := NewRuntimeConfigProvider(namespace, builder)

		configMapClient := fake.NewSimpleClientset().CoreV1().ConfigMaps(namespace)

		builder.On("CreateK8SConfigMapClient", kubeconfig, namespace).Return(configMapClient, nil)

		//when
		configMap, err := configProvider.CreateConfigMapForRuntime(config, kubeconfig)

		//then
		require.NoError(t, err)
		assert.Equal(t, connectorURL, configMap.Data["CONNECTOR_URL"])
		assert.Equal(t, runtimeID, configMap.Data["RUNTIME_ID"])
		assert.Equal(t, tenant, configMap.Data["TENANT"])
		assert.Equal(t, token, configMap.Data["TOKEN"])
	})

	t.Run("Should return error when ClientSet returns error", func(t *testing.T) {
		//given
		config := RuntimeConfig{
			ConnectorURL: connectorURL,
			RuntimeID:    runtimeID,
			Tenant:       tenant,
			OneTimeToken: token,
		}

		builder := &mocks.ConfigMapClientBuilder{}

		configProvider := NewRuntimeConfigProvider(namespace, builder)

		builder.On("CreateK8SConfigMapClient", kubeconfig, namespace).Return(nil, errors.New("Some bad bad error"))

		//when
		configMap, err := configProvider.CreateConfigMapForRuntime(config, kubeconfig)

		//then
		require.Error(t, err)
		require.Empty(t, configMap)
	})
}
