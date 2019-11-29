package hydroform

import (
	"bytes"
	"testing"

	"github.com/hashicorp/terraform/states"

	"github.com/hashicorp/terraform/states/statefile"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client/mocks"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProviderConfiguration struct {
	cluster  *types.Cluster
	provider *types.Provider
}

func (c mockProviderConfiguration) ToHydroformConfiguration(credentialsFileName string) (*types.Cluster, *types.Provider) {
	return c.cluster, c.provider
}

const (
	credentialsSecret = "credentials-secret"
	secretsNamespace  = "namespace"
)

var (
	//terraformState = `{"TerraformState":{"TerraformVersion":null,"Serial":0,"Lineage":"","State":null}}`
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

	hydroformCluster := &types.Cluster{}
	hydroformProvider := &types.Provider{}

	clusterData := model.Cluster{
		ID:            "abcd",
		ClusterConfig: mockProviderConfiguration{cluster: hydroformCluster, provider: hydroformProvider},
	}

	stateFileBytes := bytes.NewBuffer([]byte{})
	err := statefile.Write(terraformStateFile, stateFileBytes)
	require.NoError(t, err)

	t.Run("Should provision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		secretsClient := fake.NewSimpleClientset(secret).CoreV1().Secrets(secretsNamespace)

		hydroformClient.On("Provision", hydroformCluster, hydroformProvider).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: terraformStateFile}}}, nil)
		hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
		hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return([]byte("kubeconfig"), nil)

		hydroformService := NewHydroformService(hydroformClient, secretsClient)

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

	hydroformCluster := &types.Cluster{}
	hydroformProvider := &types.Provider{}

	clusterData := model.Cluster{
		ID:            "abcd",
		ClusterConfig: mockProviderConfiguration{cluster: hydroformCluster, provider: hydroformProvider},
	}

	secretsClient := fake.NewSimpleClientset(secret).CoreV1().Secrets(secretsNamespace)

	for _, testCase := range []struct {
		description string
		mockFunc    func(hydroformClient *mocks.Client)
	}{
		{
			description: "fail to fetch kubeconfig",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &statefile.File{}}}}, nil)
				hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
				hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			description: "fail to get cluster status",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &statefile.File{}}}}, nil)
				hydroformClient.On("Status", mock.Anything, mock.Anything).Return(nil, errors.New("error"))
			},
		},
		{
			description: "fail to provision cluster",
			mockFunc: func(hydroformClient *mocks.Client) {
				hydroformClient.On("Provision", hydroformCluster, hydroformProvider).Return(nil, errors.New("error"))
			},
		},
	} {
		t.Run("should return error when "+testCase.description, func(t *testing.T) {
			//given
			hydroformClient := &mocks.Client{}

			testCase.mockFunc(hydroformClient)

			hydroformService := NewHydroformService(hydroformClient, secretsClient)

			//when
			_, err := hydroformService.ProvisionCluster(clusterData)

			//then
			require.Error(t, err)
			assert.Contains(t, err.Error(), "error")
			hydroformClient.AssertExpectations(t)
		})
	}

	t.Run("should return error when no credentials in secret", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		secretsClient := fake.NewSimpleClientset(&v1.Secret{ObjectMeta: v12.ObjectMeta{Name: credentialsSecret, Namespace: secretsNamespace}}).
			CoreV1().Secrets(secretsNamespace)

		hydroformService := NewHydroformService(hydroformClient, secretsClient)

		//when
		_, err := hydroformService.ProvisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

	t.Run("should return error when credentials secret not found", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		emptySecretsClient := fake.NewSimpleClientset().CoreV1().Secrets(secretsNamespace)

		hydroformService := NewHydroformService(hydroformClient, emptySecretsClient)

		//when
		_, err := hydroformService.ProvisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

}

func TestService_DeprovisionCluster(t *testing.T) {

	stateFileBytes := bytes.NewBuffer([]byte{})
	err := statefile.Write(terraformStateFile, stateFileBytes)
	require.NoError(t, err)

	clusterData := model.Cluster{
		ID:             "abcd",
		TerraformState: stateFileBytes.Bytes(),
		ClusterConfig:  mockProviderConfiguration{cluster: &types.Cluster{}, provider: &types.Provider{}},
	}

	t.Run("Should deprovision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		secretsClient := fake.NewSimpleClientset(secret).CoreV1().Secrets(secretsNamespace)

		hydroformService := NewHydroformService(hydroformClient, secretsClient)

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
			ClusterConfig:  mockProviderConfiguration{cluster: &types.Cluster{}, provider: &types.Provider{}},
		}

		hydroformClient := &mocks.Client{}
		secretsClient := fake.NewSimpleClientset(secret).CoreV1().Secrets(secretsNamespace)

		hydroformService := NewHydroformService(hydroformClient, secretsClient)

		//when
		err := hydroformService.DeprovisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

	t.Run("Should return error when credentials secret not found", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		secretsClient := fake.NewSimpleClientset().CoreV1().Secrets(secretsNamespace)

		hydroformService := NewHydroformService(hydroformClient, secretsClient)

		//when
		err := hydroformService.DeprovisionCluster(clusterData)

		//then
		require.Error(t, err)
		hydroformClient.AssertExpectations(t)
	})

}
