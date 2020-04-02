package steps

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	"github.com/sirupsen/logrus"
	"time"
)

type ConnectAgentStep struct {
	runtimeConfigurator runtime.Configurator
	nextStep            model.OperationStage
	timeLimit           time.Duration
}

func NewConnectAgentStep(configurator runtime.Configurator, nextStep model.OperationStage, timeLimit time.Duration) *ConnectAgentStep {
	return &ConnectAgentStep{
		runtimeConfigurator: configurator,
		nextStep:            nextStep,
		timeLimit:           timeLimit,
	}
}

func (s *ConnectAgentStep) Name() model.OperationStage {
	return model.ConnectRuntimeAgent
}

func (s *ConnectAgentStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *ConnectAgentStep) Run(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (installation.StepResult, error) {

	if cluster.Kubeconfig == nil {
		return installation.StepResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	err := s.runtimeConfigurator.ConfigureRuntime(cluster, *cluster.Kubeconfig)
	if err != nil {
		return installation.StepResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	return installation.StepResult{Step: s.nextStep, Delay: 0}, nil
}
