package gardener

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/gardener/gardener/pkg/client/core/clientset/versioned/fake"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	sessionMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"

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
	operationId       = "operationId"
	clusterName       = "test-cluster"
	region            = "westeurope"

	auditLogsPolicyCMName = "audit-logs-policy"
)

func TestGardenerProvisioner_ProvisionCluster(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	gcpGardenerConfig, err := model.NewGCPGardenerConfig(&gqlschema.GCPProviderConfigInput{})
	require.NoError(t, err)

	cluster := newClusterConfig("test-cluster", "", gcpGardenerConfig, region)

	t.Run("should start provisioning", func(t *testing.T) {
		// given
		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, nil, auditLogsPolicyCMName)

		// when
		err := provisionerClient.ProvisionCluster(cluster, operationId)
		require.NoError(t, err)

		// then
		shoot, err := shootClient.Get(clusterName, v1.GetOptions{})
		require.NoError(t, err)
		assertAnnotation(t, shoot, operationIdAnnotation, operationId)
		assertAnnotation(t, shoot, runtimeIdAnnotation, runtimeId)
		assert.Equal(t, "", shoot.Labels[model.SubAccountLabel])

		require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig)
		require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy)
		require.NotNil(t, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef)
		assert.Equal(t, auditLogsPolicyCMName, shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig.AuditPolicy.ConfigMapRef.Name)
	})
}

func newClusterConfig(name, subAccountId string, providerConfig model.GardenerProviderConfig, region string) model.Cluster {
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
			Region:                 region,
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

		sessionFactoryMock := &sessionMocks.Factory{}
		session := &sessionMocks.WriteSession{}

		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, sessionFactoryMock, auditLogsPolicyCMName)

		// when
		sessionFactoryMock.On("NewWriteSession").Return(session)

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

	t.Run("should set operation success and mark shoot as deleted if shoot does not exist", func(t *testing.T) {
		// given
		clientset := fake.NewSimpleClientset()

		sessionFactoryMock := &sessionMocks.Factory{}
		session := &sessionMocks.WriteSession{}

		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, sessionFactoryMock, auditLogsPolicyCMName)

		// when
		sessionFactoryMock.On("NewWriteSession").Return(session)
		session.On("MarkClusterAsDeleted", cluster.ID).Return(nil)

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

	t.Run("should set operation failed if shoot does not exist and making it as deleted fails", func(t *testing.T) {
		// given
		clientset := fake.NewSimpleClientset()

		sessionFactoryMock := &sessionMocks.Factory{}
		session := &sessionMocks.WriteSession{}

		shootClient := clientset.CoreV1beta1().Shoots(gardenerNamespace)

		provisionerClient := NewProvisioner(gardenerNamespace, shootClient, sessionFactoryMock, auditLogsPolicyCMName)

		// when
		sessionFactoryMock.On("NewWriteSession").Return(session)
		session.On("MarkClusterAsDeleted", cluster.ID).Return(dberrors.Internal("some db error"))

		operation, err := provisionerClient.DeprovisionCluster(cluster, operationId)
		require.Error(t, err)

		// then
		assert.Equal(t, model.Failed, operation.State)
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
