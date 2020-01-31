package runtime

import (
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mocks2 "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime/clientbuilder/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProvider_CreateConfigMapForRuntime(t *testing.T) {
	connectorURL := "https://kyma.cx/connector/graphql"
	runtimeID := "123-123-456"
	tenant := "tenant"
	token := "shdfv7123ygfbw832b"
	kubeconfig := "some Kubeconfig"

	namespace := "compass-system"

	cluster := model.Cluster{
		ID:     runtimeID,
		Tenant: tenant,
		KymaConfig: model.KymaConfig{
			Components: []model.KymaComponentConfig{
				{
					Namespace: namespace,
					Component: runtimeAgentComponentName,
				},
			},
		},
	}

	oneTimeToken := graphql.OneTimeTokenForRuntimeExt{
		OneTimeTokenForRuntime: graphql.OneTimeTokenForRuntime{
			TokenWithURL: graphql.TokenWithURL{Token: token, ConnectorURL: connectorURL},
		},
	}

	t.Run("Should configure Runtime Agent", func(t *testing.T) {
		//given
		builder := &mocks.ConfigMapClientBuilder{}
		directorClient := &mocks2.DirectorClient{}

		configMapClient := fake.NewSimpleClientset().CoreV1().ConfigMaps(namespace)

		directorClient.On("GetConnectionToken", runtimeID, tenant).Return(oneTimeToken, nil)
		builder.On("CreateK8SConfigMapClient", kubeconfig, namespace).Return(configMapClient, nil)

		configProvider := NewRuntimeConfigurator(builder, directorClient)

		//when
		err := configProvider.ConfigureRuntime(cluster, kubeconfig)

		//then
		require.NoError(t, err)
		configMap, err := configMapClient.Get(configMapName, v1.GetOptions{})
		require.NoError(t, err)

		assert.Equal(t, connectorURL, configMap.Data["CONNECTOR_URL"])
		assert.Equal(t, runtimeID, configMap.Data["RUNTIME_ID"])
		assert.Equal(t, tenant, configMap.Data["TENANT"])
		assert.Equal(t, token, configMap.Data["TOKEN"])
	})

	t.Run("Should skip Runtime Agent configuration if component not provided", func(t *testing.T) {
		//given
		clusterWithoutAgent := model.Cluster{
			ID:     runtimeID,
			Tenant: tenant,
			KymaConfig: model.KymaConfig{
				Components: []model.KymaComponentConfig{
					{
						Namespace: namespace,
						Component: "core",
					},
				},
			},
		}

		configProvider := NewRuntimeConfigurator(nil, nil)

		//when
		err := configProvider.ConfigureRuntime(clusterWithoutAgent, kubeconfig)

		//then
		require.NoError(t, err)
	})

	t.Run("Should return error when failed to create client", func(t *testing.T) {
		//given

		builder := &mocks.ConfigMapClientBuilder{}
		directorClient := &mocks2.DirectorClient{}

		directorClient.On("GetConnectionToken", runtimeID, tenant).Return(oneTimeToken, nil)
		builder.On("CreateK8SConfigMapClient", kubeconfig, namespace).Return(nil, errors.New("Some bad bad error"))

		configProvider := NewRuntimeConfigurator(builder, directorClient)

		//when
		err := configProvider.ConfigureRuntime(cluster, kubeconfig)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when failed to fetch token", func(t *testing.T) {
		//given
		directorClient := &mocks2.DirectorClient{}

		directorClient.On("GetConnectionToken", runtimeID, tenant).Return(graphql.OneTimeTokenForRuntimeExt{}, errors.New("error"))

		configProvider := NewRuntimeConfigurator(nil, directorClient)

		//when
		err := configProvider.ConfigureRuntime(cluster, kubeconfig)

		//then
		require.Error(t, err)
	})
}
