package stages

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	v1alpha12 "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/fake"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestWaitForAgentToConnect(t *testing.T) {

	cluster := model.Cluster{
		Kubeconfig: util.StringPtr(kubeconfig),
	}

	for _, testCase := range []struct {
		state v1alpha12.ConnectionState
	}{
		{
			state: v1alpha12.Synchronized,
		},
		{
			state: v1alpha12.SynchronizationFailed,
		},
		{
			state: v1alpha12.MetadataUpdateFailed,
		},
	} {
		t.Run(fmt.Sprintf("should proceed to next step when Compass connection in state: %s", testCase.state), func(t *testing.T) {
			// given
			clientProvider := NewMockClientProvider(&v1alpha12.CompassConnection{
				ObjectMeta: v1.ObjectMeta{Name: defaultCompassConnectionName},
				Status: v1alpha12.CompassConnectionStatus{
					State: testCase.state,
				},
			})

			waitForAgentToConnectStep := NewWaitForAgentToConnectStep(clientProvider.NewCompassConnectionClient, nextStageName, 10*time.Minute)

			// when
			result, err := waitForAgentToConnectStep.Run(cluster, model.Operation{}, logrus.New())

			// then
			require.NoError(t, err)
			require.Equal(t, nextStageName, result.Stage)
			require.Equal(t, time.Duration(0), result.Delay)
		})
	}

	t.Run("should proceed to next step when Agent connects", func(t *testing.T) {
		// given
		clientProvider := NewMockClientProvider(&v1alpha12.CompassConnection{
			ObjectMeta: v1.ObjectMeta{Name: defaultCompassConnectionName},
			Status: v1alpha12.CompassConnectionStatus{
				State: v1alpha12.MetadataUpdateFailed,
			},
		})

		waitForAgentToConnectStep := NewWaitForAgentToConnectStep(clientProvider.NewCompassConnectionClient, nextStageName, 10*time.Minute)

		// when
		result, err := waitForAgentToConnectStep.Run(cluster, model.Operation{}, logrus.New())

		// then
		require.NoError(t, err)
		require.Equal(t, nextStageName, result.Stage)
		require.Equal(t, time.Duration(0), result.Delay)
	})

	t.Run("should rerun step if connection not yet synchronize", func(t *testing.T) {
		// given
		clientProvider := NewMockClientProvider(&v1alpha12.CompassConnection{
			ObjectMeta: v1.ObjectMeta{Name: defaultCompassConnectionName},
			Status: v1alpha12.CompassConnectionStatus{
				State: v1alpha12.Connected,
			},
		})

		waitForAgentToConnectStep := NewWaitForAgentToConnectStep(clientProvider.NewCompassConnectionClient, nextStageName, 10*time.Minute)

		// when
		result, err := waitForAgentToConnectStep.Run(cluster, model.Operation{}, logrus.New())

		// then
		require.NoError(t, err)
		require.Equal(t, model.WaitForAgentToConnect, result.Stage)
		require.Equal(t, 2*time.Second, result.Delay)
	})

	t.Run("should rerun step if Compass connection not found", func(t *testing.T) {
		// given
		clientProvider := NewMockClientProvider(&v1alpha12.CompassConnection{})

		waitForAgentToConnectStep := NewWaitForAgentToConnectStep(clientProvider.NewCompassConnectionClient, nextStageName, 10*time.Minute)

		// when
		result, err := waitForAgentToConnectStep.Run(cluster, model.Operation{}, logrus.New())

		// then
		require.NoError(t, err)
		require.Equal(t, model.WaitForAgentToConnect, result.Stage)
		require.Equal(t, 5*time.Second, result.Delay)
	})

	t.Run("should return error if Compass Connection in Connection Failed state", func(t *testing.T) {
		// given
		clientProvider := NewMockClientProvider(&v1alpha12.CompassConnection{
			ObjectMeta: v1.ObjectMeta{Name: defaultCompassConnectionName},
			Status: v1alpha12.CompassConnectionStatus{
				State: v1alpha12.ConnectionFailed,
			},
		})

		waitForAgentToConnectStep := NewWaitForAgentToConnectStep(clientProvider.NewCompassConnectionClient, nextStageName, 10*time.Minute)

		// when
		_, err := waitForAgentToConnectStep.Run(cluster, model.Operation{}, logrus.New())

		// then
		require.Error(t, err)
	})
}

type mockClientProvider struct {
	compassConnection *v1alpha12.CompassConnection
}

func NewMockClientProvider(compassConnection *v1alpha12.CompassConnection) *mockClientProvider {
	return &mockClientProvider{
		compassConnection: compassConnection,
	}
}

func (m *mockClientProvider) NewCompassConnectionClient(k8sConfig *rest.Config) (v1alpha1.CompassConnectionInterface, error) {
	fakeClient := fake.NewSimpleClientset(m.compassConnection)

	return fakeClient.CompassV1alpha1().CompassConnections(), nil
}
