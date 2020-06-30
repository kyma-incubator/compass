package gardener

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/mitchellh/mapstructure"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v12 "k8s.io/api/core/v1"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	policyConfigMapName string, maintenanceWindowConfigPath string) *GardenerProvisioner {
	return &GardenerProvisioner{
		namespace:                   namespace,
		shootClient:                 shootClient,
		dbSessionFactory:            factory,
		policyConfigMapName:         policyConfigMapName,
		maintenanceWindowConfigPath: maintenanceWindowConfigPath,
	}
}

type GardenerProvisioner struct {
	namespace                   string
	shootClient                 gardener_apis.ShootInterface
	dbSessionFactory            dbsession.Factory
	directorService             director.DirectorClient
	policyConfigMapName         string
	maintenanceWindowConfigPath string
}

func (g *GardenerProvisioner) ProvisionCluster(cluster model.Cluster, operationId string) error {
	shootTemplate, err := cluster.ClusterConfig.ToShootTemplate(g.namespace, cluster.Tenant, util.UnwrapStr(cluster.SubAccountId))
	if err != nil {
		return fmt.Errorf("failed to convert cluster config to Shoot template")
	}

	region := getRegion(cluster)

	if g.shouldSetMaintenanceWindow() {
		err := g.setMaintenanceWindow(shootTemplate, region)

		if err != nil {
			return fmt.Errorf("error setting maintenance window for %s cluster: %s", cluster.ID, err.Error())
		}
	}

	annotate(shootTemplate, operationIdAnnotation, operationId)
	annotate(shootTemplate, runtimeIdAnnotation, cluster.ID)

	if g.policyConfigMapName != "" {
		g.applyAuditConfig(shootTemplate)
	}

	_, err = g.shootClient.Create(shootTemplate)
	if err != nil {
		return fmt.Errorf("error creating Shoot for %s cluster: %s", cluster.ID, err.Error())
	}

	return nil
}

func (g *GardenerProvisioner) UpgradeCluster(clusterID string, upgradeConfig model.GardenerConfig) error {

	shoot, err := g.shootClient.Get(upgradeConfig.Name, v1.GetOptions{})
	if err != nil {
		return /*model.Operation{}, */ fmt.Errorf("error getting Shoot for cluster ID %s and name %s : %s", clusterID, upgradeConfig.Name, err.Error())
	}

	allowPrivlagedContainers := true
	enableBasicAuthentication := false

	// update needed parameters

	shoot.Spec.Region = upgradeConfig.Region
	shoot.Spec.Kubernetes = gardener_types.Kubernetes{
		AllowPrivilegedContainers: &allowPrivlagedContainers,
		Version:                   upgradeConfig.KubernetesVersion,
		KubeAPIServer: &gardener_types.KubeAPIServerConfig{
			EnableBasicAuthentication: &enableBasicAuthentication,
		},
	}

	// ????? wtf %(o)%
	err = upgradeConfig.GardenerProviderConfig.ExtendShootConfig(upgradeConfig, shoot)

	if err != nil {
		return /*model.Operation{}, */ fmt.Errorf("error extending shoot config with Provider: %s", err.Error())
	}

	_, err = g.shootClient.Update(shoot)

	if err != nil {
		return /*model.Operation{},*/ fmt.Errorf("error executing update shoot configuration: %s", err.Error())
	}

	return nil
}

func (g *GardenerProvisioner) DeprovisionCluster(cluster model.Cluster, operationId string) (model.Operation, error) {

	gardenerCfg, ok := cluster.GardenerConfig()
	if !ok {
		return model.Operation{}, fmt.Errorf("cluster does not have Gardener configuration")
	}

	shoot, err := g.shootClient.Get(gardenerCfg.Name, v1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			message := fmt.Sprintf("Cluster %s already deleted. Proceeding to DeprovisionCluster stage.", cluster.ID)

			// Shoot was deleted. In order to make sure if all clean up actions were performed we need to proceed to WaitForClusterDeletion state
			return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.WaitForClusterDeletion, time.Now()), nil
		}
	}

	if shoot.DeletionTimestamp != nil {
		annotate(shoot, operationIdAnnotation, operationId)
		message := fmt.Sprintf("Cluster %s already %s scheduled for deletion.", gardenerCfg.Name, cluster.ID)
		return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.WaitForClusterDeletion, shoot.DeletionTimestamp.Time), nil
	}

	deletionTime := time.Now()

	annotate(shoot, operationIdAnnotation, operationId)

	annotateWithConfirmDeletion(shoot)
	if err != nil {
		return model.Operation{}, fmt.Errorf("error scheduling shoot %s for deletion: %s", shoot.Name, err.Error())
	}

	_, err = g.shootClient.Update(shoot)
	if err != nil {
		return model.Operation{}, fmt.Errorf("error updating Shoot: %s", err.Error())
	}

	message := fmt.Sprintf("Deprovisioning started")
	return newDeprovisionOperation(operationId, cluster.ID, message, model.InProgress, model.CleanupCluster, deletionTime), nil
}

func annotateWithConfirmDeletion(shoot *gardener_types.Shoot) {
	if shoot.Annotations == nil {
		shoot.Annotations = map[string]string{}
	}

	shoot.Annotations["confirmation.garden.sapcloud.io/deletion"] = "true"
}

func (g *GardenerProvisioner) shouldSetMaintenanceWindow() bool {
	return g.maintenanceWindowConfigPath != ""
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

func (g *GardenerProvisioner) applyAuditConfig(template *gardener_types.Shoot) {
	if template.Spec.Kubernetes.KubeAPIServer == nil {
		template.Spec.Kubernetes.KubeAPIServer = &gardener_types.KubeAPIServerConfig{}
	}

	template.Spec.Kubernetes.KubeAPIServer.AuditConfig = &gardener_types.AuditConfig{
		AuditPolicy: &gardener_types.AuditPolicy{
			ConfigMapRef: &v12.ObjectReference{Name: g.policyConfigMapName},
		},
	}
}

func (g *GardenerProvisioner) setMaintenanceWindow(template *gardener_types.Shoot, region string) error {
	window, err := g.getWindowByRegion(region)

	if err != nil {
		return err
	}

	if !window.isEmpty() {
		setMaintenanceWindow(window, template)
	} else {
		logrus.Warnf("Cannot set maintenance window. Config for region %s is empty", region)
	}
	return nil
}

func setMaintenanceWindow(window TimeWindow, template *gardener_types.Shoot) {
	template.Spec.Maintenance.TimeWindow = &gardener_types.MaintenanceTimeWindow{Begin: window.Begin, End: window.End}
}

func (g *GardenerProvisioner) getWindowByRegion(region string) (TimeWindow, error) {
	data, err := getDataFromFile(g.maintenanceWindowConfigPath, region)

	if err != nil {
		return TimeWindow{}, err
	}

	var window TimeWindow

	err = mapstructure.Decode(data, &window)

	if err != nil {
		return TimeWindow{}, err
	}

	return window, nil
}

type TimeWindow struct {
	Begin string
	End   string
}

func (tw TimeWindow) isEmpty() bool {
	return tw.Begin == "" || tw.End == ""
}

func getDataFromFile(filepath, region string) (interface{}, error) {
	file, err := os.Open(filepath)

	if err != nil {
		return "", err
	}

	defer file.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return "", err
	}
	return data[region], nil
}

func getRegion(cluster model.Cluster) string {
	config, ok := cluster.GardenerConfig()
	if ok {
		return config.Region
	}
	return ""
}
