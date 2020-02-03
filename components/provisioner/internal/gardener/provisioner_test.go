package gardener

import (
	"fmt"
	"testing"

	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	"github.com/gardener/gardener/pkg/client/garden/clientset/versioned/fake"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gardenerNamespace = "default"
	runtimeId         = "runtimeId"
	operationId       = "operationId"
	clusterName       = "test-cluster"
)

func TestGardenerProvisioner_ProvisionCluster(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	//shootClient := clientset.GardenV1beta1().Shoots(gardenerNamespace)

	gcpGardenerConfig, err := model.NewGCPGardenerConfig(&gqlschema.GCPProviderConfigInput{})
	require.NoError(t, err)

	cluster := model.Cluster{
		ID: runtimeId,
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			ClusterID:              runtimeId,
			Name:                   clusterName,
			ProjectName:            "project-name",
			KubernetesVersion:      "1.16",
			NodeCount:              4,
			VolumeSizeGB:           50,
			DiskType:               "standard",
			MachineType:            "n1-standard-4",
			Provider:               "gcp",
			TargetSecret:           "secret",
			Region:                 "eu",
			WorkerCidr:             "10.10.10.10",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               25,
			MaxUnavailable:         1,
			GardenerProviderConfig: gcpGardenerConfig,
		},
	}

	t.Run("should start provisioning", func(t *testing.T) {
		// given
		shootClient := clientset.GardenV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient)

		// when
		err := provisionerClient.ProvisionCluster(cluster, operationId)
		require.NoError(t, err)

		// then
		shoot, err := shootClient.Get(clusterName, v1.GetOptions{})
		require.NoError(t, err)
		assertAnnotation(t, shoot, operationIdAnnotation, operationId)
		assertAnnotation(t, shoot, runtimeIdAnnotation, runtimeId)
		assertAnnotation(t, shoot, provisioningStepAnnotation, ProvisioningInProgressStep.String())
	})

}

func assertAnnotation(t *testing.T, shoot *gardener_types.Shoot, name, value string) {
	annotations := shoot.Annotations
	if annotations == nil {
		t.Errorf("annotations are nil, expected annotation: %s, value: %s", name, value)
		return
	}

	val, found := annotations[name]
	if !found {
		t.Errorf("annotation not found, expected annotation: %s, value: %s", name, value)
		return
	}

	assert.Equal(t, value, val, fmt.Sprintf("invalid value for %s annotation", name))
}
