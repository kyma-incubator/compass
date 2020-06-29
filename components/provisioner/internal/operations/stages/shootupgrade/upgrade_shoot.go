package shootupgrade

import (
	"errors"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

type GardenerClient interface {
	Get(name string, options v1.GetOptions) (*gardener_types.Shoot, error)
}

type WaitForShootClusterUpgradeStep struct {
	gardenerClient GardenerClient
	dbSession      dbsession.ReadWriteSession
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

func NewWaitForShootClusterUpgradeStep(gardenerClient GardenerClient, dbSession dbsession.ReadWriteSession, nextStep model.OperationStage, timeLimit time.Duration) *WaitForShootClusterUpgradeStep {
	return &WaitForShootClusterUpgradeStep{
		gardenerClient: gardenerClient,
		dbSession:      dbSession,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s WaitForShootClusterUpgradeStep) Name() model.OperationStage {
	return model.UpdatingUpgradeState
}

func (s *WaitForShootClusterUpgradeStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForShootClusterUpgradeStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	// here will go the implementation of the shoot uptdate

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		err := errors.New("failed to convert to GardenerConfig")
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	shoot, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		return operations.StageResult{}, err
	}


	lastOperation := shoot.Status.LastOperation



	// ???
	//dberr := s.dbSession.UpdateUpgradeState(operation.ID, model.UpgradeSucceeded)
	//if dberr != nil {
	//	return operations.StageResult{}, dberr
	//}

	logger.Info("Shoot upgrade state updated. Proceeding to next step...")
	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
