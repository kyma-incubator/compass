package provisioning

import (
	"errors"
	"fmt"
	"time"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/configuration"

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
	persistenceService   persistence.Service
	hydroform            hydroform.Service
	configBuilderFactory configuration.BuilderFactory
	installationService installation.Service
	uuidGenerator        persistence.UUIDGenerator
}

func NewProvisioningService(persistenceService persistence.Service, uuidGenerator persistence.UUIDGenerator,
	hydroform hydroform.Service, configBuilderFactory configuration.BuilderFactory, installationService installation.Service) Service {
	return &service{
		persistenceService:   persistenceService,
		hydroform:            hydroform,
		configBuilderFactory: configBuilderFactory,
		installationService: installationService,
		uuidGenerator:        uuidGenerator,
	}
}

func (r *service) ProvisionRuntime(id string, config gqlschema.ProvisionRuntimeInput) (string, <-chan struct{}, error) {
	err := r.checkProvisioningRuntimeConditions(id)
	if err != nil {
		return "", nil, err
	}

	runtimeConfig, err := runtimeConfigFromInput(id, config, r.uuidGenerator)

	if err != nil {
		return "", nil, err
	}

	operation, err := r.persistenceService.SetProvisioningStarted(id, runtimeConfig)

	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	builder := r.configBuilderFactory.NewProvisioningBuilder(config)

	go r.startProvisioning(operation.ID, id, builder, finished)

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
	runtimeStatus, err := r.persistenceService.GetStatus(id)

	if err != nil {
		return "", nil, err
	}

	if runtimeStatus.LastOperationStatus.State == model.InProgress {
		return "", nil, errors.New("cannot start new operation while previous one is in progress")
	}

	operation, err := r.persistenceService.SetDeprovisioningStarted(id)

	if err != nil {
		return "", nil, err
	}

	finished := make(chan struct{})

	cluster, dberr := r.persistenceService.GetClusterData(id)

	if dberr != nil {
		return "", nil, dberr
	}

	builder := r.configBuilderFactory.NewDeprovisioningBuilder(runtimeStatus.RuntimeConfiguration)

	go r.startDeprovisioning(operation.ID, id, builder, cluster.TerraformState, finished)

	return operation.ID, finished, nil
}

func (r *service) UpgradeRuntime(id string, config *gqlschema.UpgradeRuntimeInput) (string, error) {
	return "", nil
}

func (r *service) ReconnectRuntimeAgent(id string) (string, error) {
	return "", nil
}

func (r *service) RuntimeStatus(runtimeID string) (*gqlschema.RuntimeStatus, error) {
	runtimeStatus, err := r.persistenceService.GetStatus(runtimeID)

	if err != nil {
		return nil, err
	}

	status := runtimeStatusToGraphQLStatus(runtimeStatus)

	return status, nil
}

func (r *service) RuntimeOperationStatus(operationID string) (*gqlschema.OperationStatus, error) {
	operation, err := r.persistenceService.Get(operationID)

	if err != nil {
		return nil, err
	}

	status := operationStatusToGQLOperationStatus(operation)

	return status, nil
}

func (r *service) CleanupRuntimeData(id string) (string, error) {
	return id, r.persistenceService.CleanupClusterData(id)
}

func (r *service) startProvisioning(operationID, runtimeID string, builder configuration.Builder, finished chan<- struct{}) {
	defer close(finished)

	log.Infof("Provisioning runtime %s is starting...", runtimeID)
	info, err := r.hydroform.ProvisionCluster(builder)
	if err != nil {
		log.Errorf("Error provisioning runtime %s: %s", runtimeID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}
	if info.ClusterStatus != types.Provisioned {
		log.Errorf("Provisioning runtime %s failed, cluster status: %s", runtimeID, info.ClusterStatus)
		r.setOperationAsFailed(operationID, fmt.Sprintf("Provisioning failed from unknown reason, cluster status: %s", info.ClusterStatus))
		return
	}

	err = r.persistenceService.Update(runtimeID, info.KubeConfig, info.State)
	if err != nil {
		log.Errorf("Failed to update runtime with status")
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Runtime %s provisioned successfully. Starting Kyma installation...", runtimeID)

	kubeconfig, err := clientcmd.NewClientConfigFromBytes([]byte(info.KubeConfig))
	if err != nil {
		log.Errorf("Error provisioning runtime %s: %s", runtimeID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Errorf("Error provisioning runtime %s: %s", runtimeID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	err = r.installationService.InstallKyma(clientConfig, config.KymaConfig.Version)
	if err != nil {
		log.Errorf("Error installing Kyma on runtime %s: %s", runtimeID, err.Error())
		r.setOperationAsFailed(operationID, err.Error())
		return
	}

	log.Infof("Kyma installed successfully on %s Runtime", runtimeID)

	updateOperationStatus(func() error {
		return r.persistenceService.SetAsSucceeded(operationID)
	})
}

func (r *service) startDeprovisioning(operationID, runtimeID string, builder configuration.Builder, terraformState string, finished chan<- struct{}) {
	defer close(finished)
	log.Infof("Deprovisioning runtime %s is starting", runtimeID)
	err := r.hydroform.DeprovisionCluster(builder, terraformState)

	if err != nil {
		log.Errorf("Deprovisioning runtime %s failed: %s", runtimeID, err.Error())
		updateOperationStatus(func() error {
			return r.persistenceService.SetAsFailed(operationID, err.Error())
		})
	} else {
		log.Infof("Deprovisioning runtime %s finished successfully", runtimeID)
		updateOperationStatus(func() error {
			return r.persistenceService.SetAsSucceeded(operationID)
		})
	}
}

func (r *service) setOperationAsFailed(operationID, message string) {
	updateOperationStatus(func() error {
		return r.persistenceService.SetAsFailed(operationID, message)
	})
}

func updateOperationStatus(updateFunction func() error) {
	err := util.Retry(interval, retryCount, updateFunction)
	if err != nil {
		log.Errorf("Failed to set operation status, %s", err.Error())
	}
}
