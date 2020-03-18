package gardener

import (
	"fmt"
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	v12 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
)

func NewProvisioner(namespace string, client gardener_apis.ShootInterface, auditLogsCMName, auditLogTenant string) *GardenerProvisioner {
	return &GardenerProvisioner{
		namespace:              namespace,
		client:                 client,
		auditLogsConfigMapName: auditLogsCMName,
		auditLogTenant:         auditLogTenant,
	}
}

type GardenerProvisioner struct {
	namespace              string
	client                 gardener_apis.ShootInterface
	auditLogsConfigMapName string
	auditLogTenant         string
	log                    *logrus.Entry
}

func (g *GardenerProvisioner) ProvisionCluster(cluster model.Cluster, operationId string) error {
	shootTemplate, err := cluster.ClusterConfig.ToShootTemplate(g.namespace, cluster.SubAccountId)
	if err != nil {
		return fmt.Errorf("Failed to convert cluster config to Shoot template")
	}

	if g.shouldEnableAuditLogs() {
		enableAuditLogs(shootTemplate, g.auditLogsConfigMapName, g.auditLogTenant)
	}

	annotate(shootTemplate, operationIdAnnotation, operationId)
	annotate(shootTemplate, provisioningAnnotation, Provisioning.String())
	annotate(shootTemplate, runtimeIdAnnotation, cluster.ID)
	annotate(shootTemplate, provisioningStepAnnotation, ProvisioningInProgressStep.String())

	_, err = g.client.Create(shootTemplate)
	if err != nil {
		return fmt.Errorf("error creating Shoot for %s cluster: %s", cluster.ID, err.Error())
	}

	return nil
}

// TODO: If already deleted - try to unregister Runtime in Director?
func (g *GardenerProvisioner) DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, error) {
	gardenerCfg, ok := cluster.GardenerConfig()
	if !ok {
		return model.Operation{}, fmt.Errorf("cluster does not have Gardener configuration")
	}

	shoot, err := g.client.Get(gardenerCfg.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			message := fmt.Sprintf("Cluster %s does not exist. Nothing to deprovision.", cluster.ID)
			return newDeprovisionOperation(operationId, cluster.ID, message, model.Succeeded, time.Now()), nil
		}
	}

	if shoot.DeletionTimestamp != nil {
		message := fmt.Sprintf("Cluster %s already %s scheduled for deletion.", gardenerCfg.Name, cluster.ID)
		return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, shoot.DeletionTimestamp.Time), nil
	}

	deletionTime := time.Now()

	// TODO: consider adding some annotation and uninstall before deleting shoot
	annotate(shoot, provisioningAnnotation, Deprovisioning.String())
	annotate(shoot, provisioningStepAnnotation, DeprovisioningInProgressStep.String())
	annotate(shoot, operationIdAnnotation, operationId)
	AnnotateWithConfirmDeletion(shoot)
	err = UpdateAndDeleteShoot(g.client, shoot)
	if err != nil {
		return model.Operation{}, fmt.Errorf("error scheduling shoot %s for deletion: %s", shoot.Name, err.Error())
	}

	message := fmt.Sprintf("Deprovisioning started")
	return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, deletionTime), nil
}

func (g *GardenerProvisioner) shouldEnableAuditLogs() bool {
	return g.auditLogsConfigMapName != "" && g.auditLogTenant != ""
}

func newDeprovisionOperation(id, runtimeId, message string, state model.OperationState, startTime time.Time) model.Operation {
	return model.Operation{
		ID:             id,
		Type:           model.Deprovision,
		StartTimestamp: startTime,
		State:          state,
		Message:        message,
		ClusterID:      runtimeId,
	}
}

func enableAuditLogs(shoot *gardener_types.Shoot, policyConfigMapName, subAccountId string) {
	if shoot.Spec.Kubernetes.KubeAPIServer == nil {
		shoot.Spec.Kubernetes.KubeAPIServer = &gardener_types.KubeAPIServerConfig{}
	}

	shoot.Spec.Kubernetes.KubeAPIServer.AuditConfig = &gardener_types.AuditConfig{
		AuditPolicy: &gardener_types.AuditPolicy{
			ConfigMapRef: &v12.ObjectReference{Name: policyConfigMapName},
		},
	}

	annotate(shoot, auditLogsAnnotation, subAccountId)
}
