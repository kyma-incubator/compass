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

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		log.Error("Error converting to GardenerConfig")
		err := errors.New("failed to convert to GardenerConfig")
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	shoot, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		log.Errorf("Error getting Gardener Shoot: %s", err.Error())
		return operations.StageResult{}, err
	}

	lastOperation := shoot.Status.LastOperation

	if lastOperation.State == gardencorev1beta1.LastOperationStateSucceeded {
		return s.proceedToInstallation(logger, cluster, shoot, operation.ID)
	}

	if isShootFailed(shoot) {
		log.Warningf("Provisioning failed! Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)

		err := errors.New(fmt.Sprintf("cluster provisioning failed. Last Shoot state: %s, Shoot description: %s", lastOperation.State, lastOperation.Description))

		// return non-recoverable error
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	log.Debugf("Provisioning in progress. Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)

	return operations.StageResult{Stage: s.Name(), Delay: 30 * time.Second}, nil
}

func (s *WaitForClusterInitializationStep) proceedToInstallation(log log.FieldLogger, cluster model.Cluster, shoot *gardener_types.Shoot, operationId string) (operations.StageResult, error) {

	session := s.dbsFactory.NewReadWriteSession()

	log.Debugf("Getting Kubeconfig")
	kubeconfig, err := s.kubeconfigProvider.FetchRaw(shoot.Name)
	if err != nil {
		log.Errorf("Error fetching kubeconfig for Shoot: %s", err.Error())
		return operations.StageResult{}, err
	}

	dberr := session.UpdateCluster(cluster.ID, string(kubeconfig), nil)
	if dberr != nil {
		log.Errorf("Error saving kubeconfig in database: %s", dberr.Error())
		return operations.StageResult{}, dberr
	}

	//// Set Operation stage to Starting Installation so that is properly handled by the queue
	//dberr = session.TransitionOperation(operationId, "Starting installation", model.StartingInstallation, time.Now())
	//if dberr != nil {
	//	log.Errorf("Error transitioning operation stage: %s", dberr.Error())
	//	return operations.StageResult{}, dberr
	//}

	//log.Infof("Adding operation to installation queue")
	//r.installationQueue.Add(operationId)
	//
	//log.Infof("Updating Shoot...")
	//err = r.updateShoot(shoot, func(shootToUpdate *gardener_types.Shoot) {
	//	annotate(shootToUpdate, ProvisioningAnnotation, Provisioned.String())
	//})
	//if err != nil {
	//	log.Errorf("Error updating Shoot with retries: %s", err.Error())
	//	return err
	//}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}

func isShootFailed(shoot *gardener_types.Shoot) bool {
	return shoot.Status.LastOperation.State == gardencorev1beta1.LastOperationStateFailed
}
