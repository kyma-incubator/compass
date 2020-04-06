package stages

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/sirupsen/logrus"
	"time"
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

func (s *ConnectAgentStage) Run(cluster model.Cluster, logger logrus.FieldLogger) (operation.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operation.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	err := s.runtimeConfigurator.ConfigureRuntime(cluster, *cluster.Kubeconfig)
	if err != nil {
		return operation.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	return operation.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
