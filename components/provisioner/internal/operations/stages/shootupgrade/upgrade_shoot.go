package shootupgrade

import (
	"errors"
	"fmt"
	"time"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GardenerClient interface {
	Get(name string, options v1.GetOptions) (*gardener_types.Shoot, error)
}

type WaitForShootClusterUpgradeStep struct {
	gardenerClient GardenerClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

func NewWaitForShootClusterUpgradeStep(gardenerClient GardenerClient, nextStep model.OperationStage, timeLimit time.Duration) *WaitForShootClusterUpgradeStep {
	return &WaitForShootClusterUpgradeStep{
		gardenerClient: gardenerClient,
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

	if lastOperation != nil {
		if lastOperation.State == gardencorev1beta1.LastOperationStateSucceeded {
			logger.Info("Shoot upgrade state updated.")
			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		}

		if lastOperation.State == gardencorev1beta1.LastOperationStateFailed {
			logger.Warningf("Gardener Shoot cluster upgrade failed! Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)

			err := errors.New(fmt.Sprintf("Gardener Shoot cluster upgrade failed. Last Shoot state: %s, Shoot description: %s", lastOperation.State, lastOperation.Description))

			return operations.StageResult{}, operations.NewNonRecoverableError(err)
		}
	}

	return operations.StageResult{Stage: s.Name(), Delay: 20 * time.Second}, nil
}
