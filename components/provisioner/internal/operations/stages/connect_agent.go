package stages

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/sirupsen/logrus"
)

type ConnectAgentStage struct {
	runtimeConfigurator runtime.Configurator
	nextStep            model.OperationStage
	timeLimit           time.Duration
}

func NewConnectAgentStage(configurator runtime.Configurator, nextStep model.OperationStage, timeLimit time.Duration) *ConnectAgentStage {
	return &ConnectAgentStage{
		runtimeConfigurator: configurator,
		nextStep:            nextStep,
		timeLimit:           timeLimit,
	}
}

func (s *ConnectAgentStage) Name() model.OperationStage {
	return model.ConnectRuntimeAgent
}

func (s *ConnectAgentStage) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *ConnectAgentStage) Run(cluster model.Cluster, _ model.Operation, _ logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	err := s.runtimeConfigurator.ConfigureRuntime(cluster, *cluster.Kubeconfig)
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
