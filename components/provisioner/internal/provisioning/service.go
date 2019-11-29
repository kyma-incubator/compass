package provisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/converters"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
)

const (
	interval   = 2 * time.Second
	retryCount = 5
)

//go:generate mockery -name=Service
type Service interface {
	ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, <-chan struct{}, error)
	UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error)
	DeprovisionRuntime(id string) (string, <-chan struct{}, error)
	CleanupRuntimeData(id string) (string, error)
	ReconnectRuntimeAgent(id string) (string, error)
	RuntimeStatus(id string) (*gqlschema.RuntimeStatus, error)
	RuntimeOperationStatus(id string) (*gqlschema.OperationStatus, error)
}

type service struct {
	persistenceService  persistence.Service
	hydroform           hydroform.Service
	installationService installation.Service
	inputConverter      converters.InputConverter
	graphQLConverter    converters.GraphQLConverter
}

func NewProvisioningService(persistenceService persistence.Service, inputConverter converters.InputConverter,
	graphQLConverter converters.GraphQLConverter, hydroform hydroform.Service, installationService installation.Service) Service {
	return &service{
		persistenceService:  persistenceService,
		hydroform:           hydroform,
		installationService: installationService,
		inputConverter:      inputConverter,
		graphQLConverter:    graphQLConverter,
	}
}

func (r *service) ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, <-chan struct{}, error) {
	err := r.checkProvisioningRuntimeConditions(id)
	if err != nil {
		return "", nil, err
	}

	cluster, err := r.inputConverter.ProvisioningInputToCluster(id, config)
	if err != nil {
		return "", nil, err
	}

	operation, err := r.persistenceService.SetProvisioningStarted(id, cluster)
	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	go r.startProvisioning(operation.ID, cluster, finished)

	return operation.ID, finished, err
}

func (r *service) checkProvisioningRuntimeConditions(id string) error {
	lastOperation, err := r.persistenceService.GetLastOperation(id)

	if err == nil && !lastProvisioningFailed(lastOperation) {
		return errors.New(fmt.Sprintf("cannot provision runtime. Runtime %s already provisioned", id))
	}

	if err != nil && err.Code() != dberrors.CodeNotFound {
		return err
	}

	if lastProvisioningFailed(lastOperation) {
		if _, dbErr := r.CleanupRuntimeData(id); dbErr != nil {
			return dbErr
		}
	}

	return nil
}

func lastProvisioningFailed(operation model.Operation) bool {
	return operation.Type == model.Provision && operation.State == model.Failed
}

func (r *service) DeprovisionRuntime(id string) (string, <-chan struct{}, error) {
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

	go r.startDeprovisioning(operation.ID, cluster, finished)

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

func (r *service) CleanupRuntimeData(id string) (string, error) {
	return id, r.persistenceService.CleanupClusterData(id)
}

// TODO - refactor
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
		r.setOperationAsFailed(operationID, fmt.Sprintf("Provisioning failed from unknown reason, cluster status: %s", info.ClusterStatus))
		return
	}

	err = r.persistenceService.UpdateClusterData(cluster.ID, info.KubeConfig, info.State)
	if err != nil {
		log.Errorf("Failed to update runtime with status")
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Runtime %s provisioned successfully. Starting Kyma installation...", cluster.ID)

	err = r.installationService.InstallKyma(cluster.ID, info.KubeConfig, cluster.KymaConfig.Release)
	if err != nil {
		log.Errorf("Error installing Kyma on runtime %s: %s", cluster.ID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Kyma installed successfully on %s Runtime", cluster.ID)

	updateOperationStatus(func() error {
		return r.persistenceService.SetOperationAsSucceeded(operationID)
	})
}

func (r *service) startDeprovisioning(operationID string, cluster model.Cluster, finished chan<- struct{}) {
	defer close(finished)
	log.Infof("Deprovisioning runtime %s is starting", cluster.ID)
	err := r.hydroform.DeprovisionCluster(cluster)

	if err != nil {
		log.Errorf("Deprovisioning runtime %s failed: %s", cluster.ID, err.Error())
		updateOperationStatus(func() error {
			return r.persistenceService.SetOperationAsFailed(operationID, err.Error())
		})
	} else {
		log.Infof("Deprovisioning runtime %s finished successfully", cluster.ID)
		updateOperationStatus(func() error {
			return r.persistenceService.SetOperationAsSucceeded(operationID)
		})
	}
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
