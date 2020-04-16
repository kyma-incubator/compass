package gardener

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/core/clientset/versioned/fake"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gardenerNamespace = "default"
	runtimeId         = "runtimeId"
	tenant            = "tenant"
	subAccountId      = "sub-account"
	operationId       = "operationId"
	clusterName       = "test-cluster"

	auditLogsPolicyCMName = "audit-logs-policy"
	auditLogsTenant       = "audit-tenant"
)

func TestGardenerProvisioner_ProvisionCluster(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	gcpGardenerConfig, err := model.NewGCPGardenerConfig(&gqlschema.GCPProviderConfigInput{})
	require.NoError(t, err)

	cluster := newClusterConfig("test-cluster", "", gcpGardenerConfig)

	t.Run("should start provisioning", func(t *testing.T) {
		// given
		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, "", "")

		// when
		err := provisionerClient.ProvisionCluster(cluster, operationId)
		require.NoError(t, err)

		// then
		shoot, err := shootClient.Get(clusterName, v1.GetOptions{})
		require.NoError(t, err)
		assertAnnotation(t, shoot, operationIdAnnotation, operationId)
		assertAnnotation(t, shoot, runtimeIdAnnotation, runtimeId)
		assertAnnotation(t, shoot, provisioningStepAnnotation, ProvisioningInProgressStep.String())
		assert.Equal(t, "", shoot.Labels[model.SubAccountLabel])
	})

	for _, testCase := range []struct {
		description      string
		clusterName      string
		subAccountId     string
		configMapName    string
		auditLogsTenant  string
		auditLogsEnabled bool
	}{
		{
			description:      "audit logs enabled",
			clusterName:      "test-1",
			subAccountId:     subAccountId,
			configMapName:    auditLogsPolicyCMName,
			auditLogsTenant:  auditLogsTenant,
			auditLogsEnabled: true,
		},
		{
			description:      "audit logs disabled when no tenant",
			clusterName:      "test-2",
			subAccountId:     "acc",
			configMapName:    auditLogsPolicyCMName,
			auditLogsTenant:  "",
			auditLogsEnabled: false,
		},
		{
			description:      "audit logs disabled when no CM name",
			clusterName:      "test-3",
			subAccountId:     "",
			configMapName:    "",
			auditLogsTenant:  auditLogsTenant,
			auditLogsEnabled: false,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			// given
			shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

			provisionerClient := NewProvisioner(gardenerNamespace, shootClient, testCase.configMapName, testCase.auditLogsTenant)

			// when
			err := provisionerClient.ProvisionCluster(newClusterConfig(testCase.clusterName, testCase.subAccountId, gcpGardenerConfig), operationId)
			require.NoError(t, err)

			// then
			shoot, err := shootClient.Get(testCase.clusterName, v1.GetOptions{})
			require.NoError(t, err)
			assertAnnotation(t, shoot, operationIdAnnotation, operationId)
			assertAnnotation(t, shoot, runtimeIdAnnotation, runtimeId)
			assertAnnotation(t, shoot, provisioningStepAnnotation, ProvisioningInProgressStep.String())

			assert.Equal(t, testCase.subAccountId, shoot.Labels[model.SubAccountLabel])

			require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer)
			require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication)
			assert.False(t, *shoot.Spec.Kubernetes.KubeAPIServer.EnableBasicAuthentication)

			if testCase.auditLogsEnabled {
				assertAnnotation(t, shoot, auditLogsAnnotation, auditLogsTenant)

				require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig)
				require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy)
				require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef)
				assert.Equal(t, auditLogsPolicyCMName, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef.Name)
			} else {
				assertNoAnnotation(t, shoot, auditLogsAnnotation)
			}
		})
	}

}

func newClusterConfig(name, subAccountId string, providerConfig model.GardenerProviderConfig) model.Cluster {
	return model.Cluster{
		ID:           runtimeId,
		Tenant:       tenant,
		SubAccountId: subAccountId,
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			ClusterID:              runtimeId,
			Name:                   name,
			ProjectName:            "project-name",
			KubernetesVersion:      "1.16",
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
			GardenerProviderConfig: providerConfig,
		},
	}
}

func TestGardenerProvisioner_DeprovisionCluster(t *testing.T) {

	gcpGardenerConfig, err := model.NewGCPGardenerConfig(&gqlschema.GCPProviderConfigInput{})
	require.NoError(t, err)

	cluster := model.Cluster{
		ID: runtimeId,
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			ClusterID:              runtimeId,
			Name:                   clusterName,
			ProjectName:            "project-name",
			GardenerProviderConfig: gcpGardenerConfig,
		},
	}

	t.Run("should start deprovisioning", func(t *testing.T) {
		// given
		clientset := fake.NewSimpleClientset(
			&gardener_types.Shoot{
				ObjectMeta: v1.ObjectMeta{Name: clusterName, Namespace: gardenerNamespace, Finalizers: []string{"test"}},
			})

		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, "", "")

		// when
		operation, err := provisionerClient.DeprovisionCluster(cluster, operationId)
		require.NoError(t, err)

		// then
		assert.Equal(t, model.InProgress, operation.State)
		assert.Equal(t, operationId, operation.ID)
		assert.Equal(t, runtimeId, operation.ClusterID)
		assert.Equal(t, model.Deprovision, operation.Type)

		_, err = shootClient.Get(clusterName, v1.GetOptions{})
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	})

	t.Run("should set operation success if shoot does not exist", func(t *testing.T) {
		// given
		clientset := fake.NewSimpleClientset()

		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, "", "")

		// when
		operation, err := provisionerClient.DeprovisionCluster(cluster, operationId)
		require.NoError(t, err)

		// then
		assert.Equal(t, model.Succeeded, operation.State)
		assert.Equal(t, operationId, operation.ID)
		assert.Equal(t, runtimeId, operation.ClusterID)
		assert.Equal(t, model.Deprovision, operation.Type)

		_, err = shootClient.Get(clusterName, v1.GetOptions{})
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
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

func assertNoAnnotation(t *testing.T, shoot *gardener_types.Shoot, name string) {
	annotations := shoot.Annotations
	if annotations == nil {
		return
	}

	_, found := annotations[name]
	if found {
		t.Errorf("annotation %s found when not expected", name)
	}
}
