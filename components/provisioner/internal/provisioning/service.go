package provisioning

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
)

const (
	interval   = 2 * time.Second
	retryCount = 5
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant string) (string, string, <-chan struct{}, error)
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id, tenant string) (string, <-chan struct{}, error)
	ReconnectRuntimeAgent(id string) (string, error)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, error)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, error)
}

type service struct {
	persistenceService  persistence.Service
	hydroform           hydroform.Service
	installationService installation.Service
	inputConverter      InputConverter
	graphQLConverter    GraphQLConverter
	directorService     director.DirectorClient
}

func NewProvisioningService(persistenceService persistence.Service, inputConverter InputConverter,
	graphQLConverter GraphQLConverter, hydroform hydroform.Service, installationService installation.Service, directorService director.DirectorClient) Service {
	return &service{
		persistenceService:  persistenceService,
		hydroform:           hydroform,
		installationService: installationService,
		inputConverter:      inputConverter,
		graphQLConverter:    graphQLConverter,
		directorService:     directorService,
	}
}

func (r *service) ProvisionRuntime(config gqlschema.ProvisionRuntimeInput, tenant string) (string, string, <-chan struct{}, error) {
	runtimeInput := config.RuntimeInput

	runtimeID, err := r.directorService.CreateRuntime(runtimeInput, tenant)
	if err != nil {
		return "", "", nil, err
	}

	cluster, err := r.inputConverter.ProvisioningInputToCluster(runtimeID, config)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return "", "", nil, err
	}

	operation, err := r.persistenceService.SetProvisioningStarted(runtimeID, cluster)
	if err != nil {
		r.unregisterFailedRuntime(runtimeID, tenant)
		return "", "", nil, err
	}

	finished := make(chan struct{})

	go r.startProvisioning(operation.ID, cluster, finished)

	return operation.ID, runtimeID, finished, nil
}

func (r *service) unregisterFailedRuntime(id, tenant string) {
	log.Infof("Unregistering failed Runtime %s...", id)
	err := r.directorService.DeleteRuntime(id, tenant)
	if err != nil {
		log.Warnf("Failed to unregister failed Runtime %s: %s", id, err.Error())
	}
}

func (r *service) DeprovisionRuntime(id, tenant string) (string, <-chan struct{}, error) {
	lastOperation, err := r.persistenceService.GetLastOperation(id)
	if err != nil {
		return "", nil, err
	}

	if lastOperation.State == model.InProgress {
		return "", nil, errors.Errorf("cannot start new operation for %s Runtime while previous one is in progress", id)
	}

	cluster, dberr := r.persistenceService.GetClusterData(id)
	if dberr != nil {
		return "", nil, dberr
	}

	operation, err := r.persistenceService.SetDeprovisioningStarted(id)
	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	go r.startDeprovisioning(operation.ID, tenant, cluster, finished)

	return operation.ID, finished, nil
}

func (r *service) UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, error) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, error) {
	runtimeStatus, err := r.persistenceService.GetRuntimeStatus(runtimeID)
	if err != nil {
		return nil, err
	}

	return r.graphQLConverter.RuntimeStatusToGraphQLStatus(runtimeStatus), nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, error) {
	operation, err := r.persistenceService.GetOperation(operationID)
	if err != nil {
		return nil, err
	}

	return r.graphQLConverter.OperationStatusToGQLOperationStatus(operation), nil
}

func (r *service) startProvisioning(operationID string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)

	log.Infof("Provisioning runtime %s is starting...", cluster.ID)
	info, err := r.hydroform.ProvisionCluster(cluster)
	if err != nil {
		log.Errorf("Error provisioning runtime %s: %s", cluster.ID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}
	if info.ClusterStatus != types.Provisioned {
		log.Errorf("Provisioning runtime %s failed, cluster status: %s", cluster.ID, info.ClusterStatus)
		r.setOperationAsFailed(operationID, fmt.Sprintf("Provisioning failed for unknown reason, cluster status: %s", info.ClusterStatus))
		return
	}

	err = r.persistenceService.UpdateClusterData(cluster.ID, info.KubeConfig, info.State)
	if err != nil {
		log.Errorf("Failed to update runtime with status")
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Runtime %s provisioned successfully. Starting Kyma installation...", cluster.ID)
	err = r.installationService.InstallKyma(cluster.ID, info.KubeConfig, cluster.KymaConfig.Release, cluster.KymaConfig.GlobalConfiguration, cluster.KymaConfig.Components)
	if err != nil {
		log.Errorf("Error installing Kyma on runtime %s: %s", cluster.ID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Kyma installed successfully on %s Runtime. Operation %s finished. Setting status to success.", cluster.ID, operationID)

	updateOperationStatus(func() error {
		return r.persistenceService.SetOperationAsSucceeded(operationID)
	})
}

func (r *service) startDeprovisioning(operationID, tenant string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)
	log.Infof("Deprovisioning runtime %s is starting", cluster.ID)
	err := r.hydroform.DeprovisionCluster(cluster)
	if err != nil {
		log.Errorf("Deprovisioning runtime %s failed: %s", cluster.ID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	err = r.directorService.DeleteRuntime(cluster.ID, tenant)
	if err != nil {
		log.Errorf("Deprovisioning finished. Failed to unregister Runtime %s: %s", cluster.ID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Deprovisioning runtime %s finished successfully. Operation %s finished. Setting status to success.", cluster.ID, operationID)
	updateOperationStatus(func() error {
		return r.persistenceService.SetOperationAsSucceeded(operationID)
	})
}

func (r *service) setOperationAsFailed(operationID, message string) {
	updateOperationStatus(func() error {
		return r.persistenceService.SetOperationAsFailed(operationID, message)
	})
}

func updateOperationStatus(updateFunction func() error) {
	err := util.Retry(interval, retryCount, updateFunction)
	if err != nil {
		log.Errorf("Failed to set operation status, %s", err.Error())
	}
}
