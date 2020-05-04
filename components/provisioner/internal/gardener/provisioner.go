package gardener

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"os"
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	v12 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
)

func NewProvisioner(
	namespace string,
	shootClient gardener_apis.ShootInterface,
	factory dbsession.Factory,
	auditLogTenantConfigPath string,
	auditLogsCMName string) *GardenerProvisioner {
	return &GardenerProvisioner{
		namespace:                namespace,
		shootClient:              shootClient,
		dbSessionFactory:         factory,
		auditLogTenantConfigPath: auditLogTenantConfigPath,
		auditLogsConfigMapName:   auditLogsCMName,
	}
}

type GardenerProvisioner struct {
	namespace                string
	shootClient              gardener_apis.ShootInterface
	dbSessionFactory         dbsession.Factory
	directorService          director.DirectorClient
	auditLogsConfigMapName   string
	auditLogTenantConfigPath string
}

func (g *GardenerProvisioner) ProvisionCluster(cluster model.Cluster, operationId string) error {
	shootTemplate, err := cluster.ClusterConfig.ToShootTemplate(g.namespace, cluster.Tenant, util.UnwrapStr(cluster.SubAccountId))
	if err != nil {
		return fmt.Errorf("failed to convert cluster config to Shoot template")
	}

	region := getRegion(cluster)

	if g.shouldEnableAuditLogs() {
		if err := g.enableAuditLogs(shootTemplate, g.auditLogsConfigMapName, region); err != nil {
			return fmt.Errorf("error enabling audit logs for %s cluster: %s", cluster.ID, err.Error())
		}
	}

	annotate(shootTemplate, operationIdAnnotation, operationId)
	annotate(shootTemplate, provisioningAnnotation, Provisioning.String())
	annotate(shootTemplate, runtimeIdAnnotation, cluster.ID)

	_, err = g.shootClient.Create(shootTemplate)
	if err != nil {
		return fmt.Errorf("error creating Shoot for %s cluster: %s", cluster.ID, err.Error())
	}

	return nil
}

func (g *GardenerProvisioner) DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, error) {
	session := g.dbSessionFactory.NewWriteSession()

	gardenerCfg, ok := cluster.GardenerConfig()
	if !ok {
		return model.Operation{}, fmt.Errorf("cluster does not have Gardener configuration")
	}

	shoot, err := g.shootClient.Get(gardenerCfg.Name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			message := fmt.Sprintf("Cluster %s does not exist. Nothing to deprovision.", cluster.ID)

			dberr := session.MarkClusterAsDeleted(cluster.ID)
			if dberr != nil {
				return newDeprovisionOperation(operationId, cluster.ID, message, model.Failed, model.FinishedStage, time.Now()), dberr
			}
			return newDeprovisionOperation(operationId, cluster.ID, message, model.Succeeded, model.FinishedStage, time.Now()), nil
		}
	}

	if shoot.DeletionTimestamp != nil {
		message := fmt.Sprintf("Cluster %s already %s scheduled for deletion.", gardenerCfg.Name, cluster.ID)
		return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.DeprovisioningStage, shoot.DeletionTimestamp.Time), nil
	}

	deletionTime := time.Now()

	// TODO: consider adding some annotation and uninstall before deleting shoot
	annotate(shoot, provisioningAnnotation, Deprovisioning.String())
	annotate(shoot, operationIdAnnotation, operationId)
	AnnotateWithConfirmDeletion(shoot)
	err = UpdateAndDeleteShoot(g.shootClient, shoot)
	if err != nil {
		return model.Operation{}, fmt.Errorf("error scheduling shoot %s for deletion: %s", shoot.Name, err.Error())
	}

	message := fmt.Sprintf("Deprovisioning started")
	return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.DeprovisioningStage, deletionTime), nil
}

func (g *GardenerProvisioner) shouldEnableAuditLogs() bool {
	return g.auditLogsConfigMapName != "" && g.auditLogTenantConfigPath != ""
}

func newDeprovisionOperation(id, runtimeId, message string, state model.OperationState, stage model.OperationStage, startTime time.Time) model.Operation {
	return model.Operation{
		ID:             id,
		Type:           model.Deprovision,
		StartTimestamp: startTime,
		State:          state,
		Stage:          stage,
		Message:        message,
		ClusterID:      runtimeId,
	}
}

func (g *GardenerProvisioner) enableAuditLogs(shoot *gardener_types.Shoot, policyConfigMapName, region string) error {
	logrus.Info("Enabling audit logs")
	tenant, err := g.getAuditLogTenant(region)

	if err != nil {
		return err
	}

	if tenant != "" {
		setAuditConfig(shoot, policyConfigMapName, tenant)
	} else {
		logrus.Warnf("Cannot enable audit logs. Tenant for region %s is empty", region)
	}

	return nil
}

func (g *GardenerProvisioner) getAuditLogTenant(region string) (string, error) {
	file, err := os.Open(g.auditLogTenantConfigPath)

	if err != nil {
		return "", err
	}

	defer file.Close()

	var data map[string]string
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return "", err
	}
	return data[region], nil
}

func setAuditConfig(shoot *gardener_types.Shoot, policyConfigMapName, subAccountId string) {
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

func getRegion(cluster model.Cluster) string {
	config, ok := cluster.GardenerConfig()
	if ok {
		return config.Region
	}
	return ""
}
