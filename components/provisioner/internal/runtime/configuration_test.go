package runtime

import (
	"errors"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mocks2 "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	kubeconfig = "some Kubeconfig"
)

func TestProvider_CreateConfigMapForRuntime(t *testing.T) {
	connectorURL := "https://kyma.cx/connector/graphql"
	runtimeID := "123-123-456"
	tenant := "tenant"
	token := "shdfv7123ygfbw832b"

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
		k8sClientProvider := newMockClientProvider(t)
		directorClient := &mocks2.DirectorClient{}

		directorClient.On("GetConnectionToken", runtimeID, tenant).Return(oneTimeToken, nil)

		configProvider := NewRuntimeConfigurator(k8sClientProvider, directorClient)

		//when
		err := configProvider.ConfigureRuntime(cluster, kubeconfig)

		//then
		require.NoError(t, err)
		configMap, err := k8sClientProvider.fakeClient.CoreV1().ConfigMaps(namespace).Get(AgentConfigurationSecretName, v1.GetOptions{})
		require.NoError(t, err)
		secret, err := k8sClientProvider.fakeClient.CoreV1().Secrets(namespace).Get(AgentConfigurationSecretName, v1.GetOptions{})
		require.NoError(t, err)

		assertData := func(data map[string]string) {
			assert.Equal(t, connectorURL, data["CONNECTOR_URL"])
			assert.Equal(t, runtimeID, data["RUNTIME_ID"])
			assert.Equal(t, tenant, data["TENANT"])
			assert.Equal(t, token, data["TOKEN"])
		}

		assertData(configMap.Data)
		assertData(secret.StringData)
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

		k8sClientProvider := newErrorClientProvider(t, fmt.Errorf("error"))
		directorClient := &mocks2.DirectorClient{}

		directorClient.On("GetConnectionToken", runtimeID, tenant).Return(oneTimeToken, nil)

		configProvider := NewRuntimeConfigurator(k8sClientProvider, directorClient)

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

type mockClientProvider struct {
	t          *testing.T
	fakeClient *fake.Clientset
	err        error
}

func newMockClientProvider(t *testing.T, objects ...runtime.Object) *mockClientProvider {
	return &mockClientProvider{
		t:          t,
		fakeClient: fake.NewSimpleClientset(objects...),
	}
}

func newErrorClientProvider(t *testing.T, err error) *mockClientProvider {
	return &mockClientProvider{
		t:   t,
		err: err,
	}
}

func (m *mockClientProvider) CreateK8SClient(kubeconfigRaw string) (kubernetes.Interface, error) {
	assert.Equal(m.t, kubeconfig, kubeconfigRaw)

	if m.err != nil {
		return nil, m.err
	}

	return m.fakeClient, nil
}
