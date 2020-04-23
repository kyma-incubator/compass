package hydroform

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/sirupsen/logrus"
)

const (
	interval   = 2 * time.Second
	retryCount = 5
)

func NewHydroformProvisioner(
	hydroformSvc Service,
	installationSvc installation.Service,
	factory dbsession.Factory,
	directorClient director.DirectorClient,
	runtimeConfigurator runtime.Configurator) *HydroformProvisioner {
	return &HydroformProvisioner{
		hydroformSvc:        hydroformSvc,
		installationSvc:     installationSvc,
		dbSessionFactory:    factory,
		directorService:     directorClient,
		runtimeConfigurator: runtimeConfigurator,
		logger:              logrus.WithField("Component", "HydroformProvisioner"),
	}
}

type HydroformProvisioner struct {
	hydroformSvc        Service
	installationSvc     installation.Service
	dbSessionFactory    dbsession.Factory
	directorService     director.DirectorClient
	runtimeConfigurator runtime.Configurator
	logger              *logrus.Entry
}

func (h *HydroformProvisioner) ProvisionCluster(clusterConfig model.Cluster, operationId string) error {
	_, err := h.Provision(clusterConfig, operationId)
	return err
}

func (h *HydroformProvisioner) DeprovisionCluster(clusterConfig model.Cluster, operationId string) (model.Operation, error) {
	op, _, err := h.Deprovision(clusterConfig, operationId)
	return op, err
}

func (h *HydroformProvisioner) Provision(clusterConfig model.Cluster, operationId string) (chan struct{}, error) {
	finished := make(chan struct{})

	go h.startProvisioning(operationId, clusterConfig, finished)

	return finished, nil
}

func (h *HydroformProvisioner) Deprovision(clusterConfig model.Cluster, operationId string) (model.Operation, chan struct{}, error) {
	finished := make(chan struct{})

	deprovisionTime := time.Now()

	go h.startDeprovisioning(operationId, clusterConfig, finished)

	return model.Operation{
		ID:             operationId,
		Type:           model.Deprovision,
		StartTimestamp: deprovisionTime,
		State:          model.InProgress,
		Message:        "Deprovisioning started.",
		ClusterID:      clusterConfig.ID,
	}, finished, nil
}

func (h *HydroformProvisioner) startProvisioning(operationID string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)
	log := h.newLogger(cluster.ID, operationID)

	log.Infof("Provisioning runtime %s is starting...", cluster.ID)
	info, err := h.hydroformSvc.ProvisionCluster(cluster)
	if err != nil {
		log.Errorf("Error provisioning runtime %s: %s", cluster.ID, err.Error())
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}
	if info.ClusterStatus != types.Provisioned {
		log.Errorf("Provisioning runtime %s failed, cluster status: %s", cluster.ID, info.ClusterStatus)
		h.setOperationAsFailed(log, operationID, fmt.Sprintf("Provisioning failed for unknown reason, cluster status: %s", info.ClusterStatus))
		return
	}

	dbSession := h.dbSessionFactory.NewWriteSession()

	err = dbSession.UpdateCluster(cluster.ID, info.KubeConfig, info.State)
	if err != nil {
		log.Errorf("Failed to update runtime with status")
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}

	log.Infof("Runtime %s provisioned successfully. Starting Kyma installation...", cluster.ID)
	err = h.installationSvc.InstallKyma(cluster.ID, info.KubeConfig, cluster.KymaConfig.Release, cluster.KymaConfig.GlobalConfiguration, cluster.KymaConfig.Components)
	if err != nil {
		log.Errorf("Error installing Kyma on runtime %s: %s", cluster.ID, err.Error())
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}

	log.Infof("Kyma installed successfully on %s Runtime. Applying configuration to Runtime", cluster.ID)

	err = h.runtimeConfigurator.ConfigureRuntime(cluster, info.KubeConfig)
	if err != nil {
		log.Errorf("Error applying configuration to runtime %s: %s", cluster.ID, err.Error())
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}

	log.Infof("Operation %s finished. Setting status to success.", operationID)

	updateOperationStatus(log, func() error {
		return h.setOperationAsSucceeded(operationID)
	})
}

func (h *HydroformProvisioner) startDeprovisioning(operationID string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)
	log := h.newLogger(cluster.ID, operationID)

	log.Infof("Deprovisioning runtime %s is starting", cluster.ID)
	err := h.hydroformSvc.DeprovisionCluster(cluster)
	if err != nil {
		log.Errorf("Deprovisioning runtime %s failed: %s", cluster.ID, err.Error())
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}

	session := h.dbSessionFactory.NewWriteSession()
	dberr := session.MarkClusterAsDeleted(cluster.ID)
	if dberr != nil {
		log.Errorf("Deprovisioning finished. Failed to mark cluster as deleted: %s", dberr.Error())
		h.setOperationAsFailed(log, operationID, dberr.Error())
		return
	}

	err = h.directorService.DeleteRuntime(cluster.ID, cluster.Tenant)
	if err != nil {
		log.Errorf("Deprovisioning finished. Failed to unregister Runtime %s: %s", cluster.ID, err.Error())
		h.setOperationAsFailed(log, operationID, err.Error())
		return
	}

	log.Infof("Deprovisioning runtime %s finished successfully. Operation %s finished. Setting status to success.", cluster.ID, operationID)
	updateOperationStatus(log, func() error {
		return h.setOperationAsSucceeded(operationID)
	})
}

func (h *HydroformProvisioner) newLogger(runtimeId, operationId string) *logrus.Entry {
	return h.logger.WithFields(logrus.Fields{
		"RuntimeId":   runtimeId,
		"OperationId": operationId,
	})

}

func (h *HydroformProvisioner) setOperationAsFailed(log *logrus.Entry, operationID, message string) {
	updateOperationStatus(log, func() error {
		session := h.dbSessionFactory.NewWriteSession()
		return session.UpdateOperationState(operationID, message, model.Failed, time.Now())
	})
}

func (h *HydroformProvisioner) setOperationAsSucceeded(operationID string) error {
	session := h.dbSessionFactory.NewWriteSession()
	return session.UpdateOperationState(operationID, "Operation succeeded.", model.Succeeded, time.Now())
}

func updateOperationStatus(log *logrus.Entry, updateFunction func() error) {
	err := util.Retry(interval, retryCount, updateFunction)
	if err != nil {
		log.Errorf("Failed to set operation status, %s", err.Error())
	}
}
