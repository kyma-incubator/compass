package deprovisioning

import (
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeleteClusterStep struct {
	gardenerClient GardenerClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

//go:generate mockery -name=GardenerClient
type GardenerClient interface {
	Get(name string, options v1.GetOptions) (*gardener_types.Shoot, error)
	Delete(name string, options *v1.DeleteOptions) error
}

func NewDeleteClusterStep(gardenerClient GardenerClient, nextStep model.OperationStage, timeLimit time.Duration) *DeleteClusterStep {
	return &DeleteClusterStep{
		gardenerClient: gardenerClient,
		nextStep:       nextStep,
		timeLimit:      timeLimit,
	}
}

func (s *DeleteClusterStep) Name() model.OperationStage {
	return model.DeleteCluster
}

func (s *DeleteClusterStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *DeleteClusterStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	err := s.deleteShoot(cluster.ClusterConfig.Name)
	if err != nil {
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}

func (s *DeleteClusterStep) deleteShoot(gardenerClusterName string) error {
	err := s.gardenerClient.Delete(gardenerClusterName, &v1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return nil
}
