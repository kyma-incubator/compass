package hydroform

import (
	configMock "github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/configuration/mocks"
	"testing"

	"github.com/hashicorp/terraform/terraform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var terraformState = `{"TerraformState":{"version":0,"serial":0,"lineage":"","modules":null}}`
var config = model.RuntimeConfig{ClusterConfig: model.GCPConfig{
	ID:                "id",
	Name:              "Something",
	ProjectName:       "Project",
	NumberOfNodes:     3,
	BootDiskSizeGB:    256,
	MachineType:       "n1-standard-1",
	Region:            "region",
	KubernetesVersion: "version",
	ClusterID:         "runtimeID",
}}

func TestService_ProvisionCluster(t *testing.T) {
	t.Run("Should provision cluster", func(t *testing.T) {
		//given
		hydroformClient := &mocks.Client{}
		builder := &configMock.ConfigBuilder{}

		hydroformService := NewHydroformService(hydroformClient)

		builder.On("Create").Return(&types.Cluster{}, &types.Provider{}, nil)
		builder.On("CleanUp").Return()
		hydroformClient.On("Provision", mock.Anything, mock.Anything).Return(&types.Cluster{ClusterInfo: &types.ClusterInfo{InternalState: &types.InternalState{TerraformState: &terraform.State{}}}}, nil)
		hydroformClient.On("Status", mock.Anything, mock.Anything).Return(&types.ClusterStatus{Phase: types.Provisioned}, nil)
		hydroformClient.On("Credentials", mock.Anything, mock.Anything).Return([]byte("kubeconfig"), nil)

		//when
		info, err := hydroformService.ProvisionCluster(builder)

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
	builder := &configMock.ConfigBuilder{}

	hydroformService := NewHydroformService(hydroformClient)

	builder.On("Create").Return(&types.Cluster{}, &types.Provider{}, nil)
	builder.On("CleanUp").Return()
	hydroformClient.On("Deprovision", mock.Anything, mock.Anything).Return(nil)

	//when
	err := hydroformService.DeprovisionCluster(builder, terraformState)

	//then
	require.NoError(t, err)
}
