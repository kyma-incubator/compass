package steps

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

type InstallKymaStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewInstallKymaStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *InstallKymaStep {
	return &InstallKymaStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *InstallKymaStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *InstallKymaStep) TimeLimit() time.Duration {
	return s.timeLimit
}

// TODO: probably you can remove operation
func (s *InstallKymaStep) Run(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (installation.StepResult, error) {

	if cluster.Kubeconfig == nil {
		return installation.StepResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	// TODO: cleanup kubeconfig stuff
	k8sConfig, err := ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		// TODO: recoverable or not?
		return installation.StepResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	// TODO: check if installation is started
	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("Installation already in progress, proceeding to next step...")
			return installation.StepResult{Step: s.Name(), Delay: 0}, nil
		}

		return installation.StepResult{}, fmt.Errorf("error: failed to check installation state: %s", err.Error())
	}

	if installationState.State != installationSDK.NoInstallationState {
		logger.Warnf("Installation already in progress, proceeding to next step...")
		return installation.StepResult{Step: s.nextStep, Delay: 0}, nil
	}

	// TODO: it needs to run apply instead of create
	err = s.installationClient.TriggerInstallation(
		[]byte(*cluster.Kubeconfig),
		cluster.KymaConfig.Release,
		cluster.KymaConfig.GlobalConfiguration,
		cluster.KymaConfig.Components)
	if err != nil {
		// TODO: if it runs apply then recoverable error (else not)
		return installation.StepResult{}, fmt.Errorf("error: failed to start installation: %s", err.Error())
	}

	logger.Warnf("Installation started, proceeding to next step...")
	return installation.StepResult{Step: s.nextStep, Delay: 30 * time.Second}, nil
}

func ParseToK8sConfig(kubeconfigRaw []byte) (*restclient.Config, error) {
	kubeconfig, err := clientcmd.NewClientConfigFromBytes(kubeconfigRaw)
	if err != nil {
		return nil, fmt.Errorf("error constructing kubeconfig from raw config: %s", err.Error())
	}

	clientConfig, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client kubeconfig from parsed config: %s", err.Error())
	}

	return clientConfig, nil
}
