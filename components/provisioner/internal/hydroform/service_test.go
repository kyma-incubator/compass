package hydroform

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"

	"github.com/hashicorp/terraform/states"

	"github.com/hashicorp/terraform/states/statefile"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client/mocks"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProviderConfiguration struct {
	t               *testing.T
	cluster         *types.Cluster
	provider        *types.Provider
	credentialsFile string
	err             error
}

func (c mockProviderConfiguration) ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider, error) {
	assert.Equal(c.t, c.credentialsFile, credentialsFileName)

	return c.cluster, c.provider, c.err
}

func (c mockProviderConfiguration) ToShootTemplate(namespace string, accountId string,subAccountId string) (*gardener_types.Shoot, error) {
	return nil, nil
}

const (
	credentialsSecret = "credentials-secret"
	secretsNamespace  = "namespace"

	testCredentials = "test credentials"
)

var (
	terraformStateFile = statefile.New(states.NewState(), "", 0)
	credentials        = []byte("credentials")
	secret             = &v1.Secret{
		ObjectMeta: v12.ObjectMeta{Name: credentialsSecret, Namespace: secretsNamespace},
		Data: map[string][]byte{
			"credentials": credentials,
		},
	}
)

func TestService_ProvisionCluster(t *testing.T) {

	file, err := createTestCredentialsFile()
	require.NoError(t, err)
	defer os.Remove(file)

	hydroformCluster := &types.Cluster{}
	hydroformProvider := &types.Provider{}

	clusterData := model.Cluster{
		ID: "abcd",
		ClusterConfig: mockProviderConfiguration{
			t:               t,
			credentialsFile: file,
			cluster:         hydroformCluster,
			provider:        hydroformProvider,
		},
	}

	stateFileBytes := bytes.NewBuffer([]byte{})
	err = statefile.Write(terraformStateFile, stateFileBytes)
	require.NoError(t, err)

	t.Run("Should provision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}

		hydroformClient.On("Provision", hydroformCluster, hydroformProvider, mock.AnythingOfType("types.Option")).
			Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: terraformStateFile}}}, nil)
		hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
		hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return([]byte("kubeconfig"), nil)

		hydroformService := NewHydroformService(hydroformClient, file)

		//when
		info, err := hydroformService.ProvisionCluster(clusterData)

		//then
		require.NoError(t, err)
		require.Equal(t, "kubeconfig", info.KubeConfig)
		require.Equal(t, types.Provisioned, info.ClusterStatus)
		require.Equal(t, stateFileBytes.Bytes(), info.State)
		hydroformClient.AssertExpectations(t)
	})

}

func TestService_ProvisionCluster_Errors(t *testing.T) {

	file, err := createTestCredentialsFile()
	require.NoError(t, err)
	defer os.Remove(file)

	hydroformCluster := &types.Cluster{}
	hydroformProvider := &types.Provider{}

	clusterData := model.Cluster{
		ID: "abcd",
		ClusterConfig: mockProviderConfiguration{
			t:               t,
			credentialsFile: file,
			cluster:         hydroformCluster,
			provider:        hydroformProvider,
		}}

	for _, testCase := range []struct {
		description string
		mockFunc    func(hydroformClient *mocks.Client)
	}{
		{
			description: "failed to fetch kubeconfig",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider, mock.AnythingOfType("types.Option")).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &statefile.File{}}}}, nil)
				hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
				hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			description: "failed to get cluster status",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider, mock.AnythingOfType("types.Option")).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &statefile.File{}}}}, nil)
				hydroformClient.On("Status", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			description: "failed to provision cluster",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider, mock.AnythingOfType("types.Option")).Return(nil, errors.New("error"))
			},
		},
	} {
		t.Run("should return error when "+testCase.description, func(t *testing.T) {
			//given
			hydroformClient := &mocks.Client{}

			testCase.mockFunc(hydroformClient)

			hydroformService := NewHydroformService(hydroformClient, file)

			//when
			_, err := hydroformService.ProvisionCluster(clusterData)

			//then
			require.Error(t, err)
			assert.Contains(t, err.Error(), "error")
			hydroformClient.AssertExpectations(t)
		})
	}

	t.Run("should return error when failed to convert provider config to Hydroform config", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		clusterData := model.Cluster{
			ID:            "abcd",
			ClusterConfig: mockProviderConfiguration{t: t, credentialsFile: file, err: errors.New("error")},
		}

		hydroformService := NewHydroformService(hydroformClient, file)

		//when
		_, err := hydroformService.ProvisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

}

func TestService_DeprovisionCluster(t *testing.T) {

	file, err := createTestCredentialsFile()
	require.NoError(t, err)
	defer os.Remove(file)

	stateFileBytes := bytes.NewBuffer([]byte{})
	err = statefile.Write(terraformStateFile, stateFileBytes)
	require.NoError(t, err)

	clusterData := model.Cluster{
		ID:             "abcd",
		TerraformState: stateFileBytes.Bytes(),
		ClusterConfig: mockProviderConfiguration{
			t:               t,
			credentialsFile: file,
			cluster:         &types.Cluster{},
			provider:        &types.Provider{},
		},
	}

	t.Run("Should deprovision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}

		hydroformService := NewHydroformService(hydroformClient, file)

		hydroformClient.On("Deprovision", mock.Anything, mock.Anything).Return(nil)

		//when
		err := hydroformService.DeprovisionCluster(clusterData)

		//then
		require.NoError(t, err)
		hydroformClient.AssertExpectations(t)
	})

	t.Run("Should return error when failed to decode terraform state", func(t *testing.T) {
		//given
		clusterData := model.Cluster{
			ID:             "abcd",
			TerraformState: []byte("invalid json"),
			ClusterConfig:  mockProviderConfiguration{t: t, credentialsFile: file, cluster: &types.Cluster{}, provider: &types.Provider{}},
		}

		hydroformClient := &mocks.Client{}

		hydroformService := NewHydroformService(hydroformClient, file)

		//when
		err := hydroformService.DeprovisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

	t.Run("should return error when failed to convert provider config to Hydroform config", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		clusterData := model.Cluster{
			ID:            "abcd",
			ClusterConfig: mockProviderConfiguration{t: t, credentialsFile: file, err: errors.New("error")},
		}

		hydroformService := NewHydroformService(hydroformClient, file)

		//when
		err := hydroformService.DeprovisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

}

func createTestCredentialsFile() (string, error) {
	tempFile, err := ioutil.TempFile("", "test_credentials")
	if err != nil {
		return "", errors.Wrap(err, "Failed to create credentials file")
	}

	_, err = tempFile.Write([]byte(testCredentials))
	if err != nil {
		return "", errors.WithMessagef(err, "Failed to save credentials to %s file", tempFile.Name())
	}

	return tempFile.Name(), nil
}
