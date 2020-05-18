package provisioning

import (
	"errors"
	"fmt"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WaitForClusterInitializationStep struct {
	gardenerClient     GardenerClient
	dbsFactory         dbsession.Factory
	kubeconfigProvider KubeconfigProvider
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

//go:generate mockery -name=KubeconfigProvider
type KubeconfigProvider interface {
	FetchRaw(shootName string) ([]byte, error)
}

func NewWaitForClusterInitializationStep(gardenerClient GardenerClient, dbsFactory dbsession.Factory, kubeconfigProvider KubeconfigProvider, nextStep model.OperationStage, timeLimit time.Duration) *WaitForClusterInitializationStep {
	return &WaitForClusterInitializationStep{
		gardenerClient:     gardenerClient,
		dbsFactory:         dbsFactory,
		kubeconfigProvider: kubeconfigProvider,

		nextStep:  nextStep,
		timeLimit: timeLimit,
	}
}

func (s *WaitForClusterInitializationStep) Name() model.OperationStage {
	return model.WaitingForClusterInitialization
}

func (s *WaitForClusterInitializationStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterInitializationStep) Run(cluster model.Cluster, operation model.Operation, logger log.FieldLogger) (operations.StageResult, error) {

	logger.Info("Started waiting for cluster initialization")
	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		log.Error("Error converting to GardenerConfig")
		err := errors.New("failed to convert to GardenerConfig")
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	logger.Info("Getting Shoot")
	shoot, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		log.Errorf("Error getting Gardener Shoot: %s", err.Error())
		return operations.StageResult{}, err
	}

	logger.Info("Checking Shoot state")
	lastOperation := shoot.Status.LastOperation

	if lastOperation != nil {
		if lastOperation.State == gardencorev1beta1.LastOperationStateSucceeded {
			return s.proceedToInstallation(logger, cluster, shoot, operation.ID)
		}

		if lastOperation.State == gardencorev1beta1.LastOperationStateFailed {
			log.Warningf("Provisioning failed! Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)

			err := errors.New(fmt.Sprintf("cluster provisioning failed. Last Shoot state: %s, Shoot description: %s", lastOperation.State, lastOperation.Description))

			return operations.StageResult{}, operations.NewNonRecoverableError(err)
		}
	}

	logger.Info("Cluster not initialized yet.")

	return operations.StageResult{Stage: s.Name(), Delay: 20 * time.Second}, nil
}

func (s *WaitForClusterInitializationStep) proceedToInstallation(logger log.FieldLogger, cluster model.Cluster, shoot *gardener_types.Shoot, operationId string) (operations.StageResult, error) {

	logger.Info("Proceeding to installation")
	session := s.dbsFactory.NewReadWriteSession()

	log.Info("Getting Kubeconfig")
	kubeconfig, err := s.kubeconfigProvider.FetchRaw(shoot.Name)
	if err != nil {
		log.Errorf("Error fetching kubeconfig for Shoot: %s", err.Error())
		return operations.StageResult{}, err
	}

	log.Info("Updating cluster")
	dberr := session.UpdateCluster(cluster.ID, string(kubeconfig), nil)
	if dberr != nil {
		log.Errorf("Error saving kubeconfig in database: %s", dberr.Error())
		return operations.StageResult{}, dberr
	}

	log.Info("Ready to start installation")
	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
