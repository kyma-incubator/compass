package hydroform

import (
	"github.com/hashicorp/terraform/terraform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/kubernetes/typed/core/v1"
	"testing"
)

const (
	secretName = "gardener"
	namespace  = "compass-system"
)

var terraformState = `{"TerraformState":{"version":0,"serial":0,"lineage":"","modules":null}}`
var config = model.RuntimeConfig{ClusterConfig: model.GCPConfig{
	ID:                "id",
	Name:              "Something",
	ProjectName:       "Project",
	NumberOfNodes:     3,
	BootDiskSize:      "256",
	MachineType:       "n1-standard-1",
	Region:            "region",
	KubernetesVersion: "version",
	ClusterID:         "runtimeID",
}}

func TestService_ProvisionCluster(t *testing.T) {
	t.Run("Should provision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		coreV1 := fake.NewSimpleClientset()
		secrets := coreV1.CoreV1().Secrets(namespace)

		createFakeCredentialsSecret(t, secrets)
		defer deleteSecret(t, secrets)

		hydroformService := NewHydroformService(secrets, hydroformClient)

		hydroformClient.On("Provision", mock.Anything, mock.Anything).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &terraform.State{}}}}, nil)
		hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
		hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return([]byte("kubeconfig"), nil)

		//when
		info, err := hydroformService.ProvisionCluster(config, secretName)

		//then
		require.NoError(t, err)
		require.Equal(t, "kubeconfig", info.KubeConfig)
		require.Equal(t, types.Provisioned, info.ClusterStatus)
		require.Equal(t, terraformState, info.State)
	})
}

func TestService_DeprovisionCluster(t *testing.T) {
	//given
	hydroformClient := &mocks.Client{}
	coreV1 := fake.NewSimpleClientset()
	secrets := coreV1.CoreV1().Secrets(namespace)

	createFakeCredentialsSecret(t, secrets)
	defer deleteSecret(t, secrets)

	hydroformService := NewHydroformService(secrets, hydroformClient)

	hydroformClient.On("Deprovision", mock.Anything, mock.Anything).Return(nil)

	//when
	err := hydroformService.DeprovisionCluster(config, secretName, terraformState)

	//then
	require.NoError(t, err)
}

func createFakeCredentialsSecret(t *testing.T, secrets core.SecretInterface) {
	secret := &v1.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Data: map[string][]byte{
			credentialsKey: []byte("YmFzZTY0IGNyZWRlbnRpYWxz"),
		},
	}

	_, err := secrets.Create(secret)

	require.NoError(t, err)
}

func deleteSecret(t *testing.T, secrets core.SecretInterface) {
	err := secrets.Delete(secretName, &meta.DeleteOptions{})
	require.NoError(t, err)
}
