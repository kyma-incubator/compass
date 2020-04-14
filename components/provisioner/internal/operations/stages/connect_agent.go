package stages

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/sirupsen/logrus"
)

type ConnectAgentStep struct {
	runtimeConfigurator runtime.Configurator
	nextStage           model.OperationStage
	timeLimit           time.Duration
}

func NewConnectAgentStep(configurator runtime.Configurator, nextStage model.OperationStage, timeLimit time.Duration) *ConnectAgentStep {
	return &ConnectAgentStep{
		runtimeConfigurator: configurator,
		nextStage:           nextStage,
		timeLimit:           timeLimit,
	}
}

func (s *ConnectAgentStep) Stage() model.OperationStage {
	return model.ConnectRuntimeAgent
}

func (s *ConnectAgentStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *ConnectAgentStep) Run(cluster model.Cluster, _ model.Operation, _ logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	err := s.runtimeConfigurator.ConfigureRuntime(cluster, *cluster.Kubeconfig)
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	return operations.StageResult{Stage: s.nextStage, Delay: 0}, nil
}
