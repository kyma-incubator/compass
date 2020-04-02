package upgrade

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

type UpgradeKymaStep struct {
	installationClient installation.Service
}

func NewUpgradeKymaStep(installationClient installation.Service) *UpgradeKymaStep {
	return &UpgradeKymaStep{
		installationClient: installationClient,
	}
}

func (s *UpgradeKymaStep) Name() string {
	return "UpgradeKyma"
}

// TODO: I think it would be better to return some step result
func (s *UpgradeKymaStep) Run(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (model.Operation, time.Duration, error) {

	if cluster.Kubeconfig == nil {
		// TODO: recoverable or not?
		return operation, 1 * time.Second, fmt.Errorf("error: kubeconfig is nil")
	}

	// TODO: cleanup kubeconfig stuff
	k8sConfig, err := ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		// TODO: recoverable or not?
		return operation, 1 * time.Second, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		// TODO: recoverable
		// TODO: check installation error
		if errors.Is(err, installationSDK.InstallationError{}) {
			return operation, 30 * time.Second, fmt.Errorf("error: installation error occured: %s", err.Error())
		}

		return operation, 1 * time.Second, fmt.Errorf("error: failed to check installation state: %s", err.Error())
	}
	//
	//if installationState.State == installationSDK.NoInstallationState {
	//	// TODO: it needs to run apply instead of create
	//	err := s.installationClient.TriggerInstallation(
	//		[]byte(*cluster.Kubeconfig),
	//		cluster.KymaConfig.Release,
	//		cluster.KymaConfig.GlobalConfiguration,
	//		cluster.KymaConfig.Components)
	//	if err != nil {
	//		// TODO: if it runs apply then recoverable error (else not)
	//		return operation, 2 * time.Second, fmt.Errorf("error: failed to start installation: %s", err.Error())
	//	}
	//
	//	// TODO: update operation that installation started
	//}
	//
	//if installationState.State == "Installed" {
	//	operation.Message = "Installation completed"
	//	logger.Infof("Installation completed: %s", installationState.Description)
	//	return operation, 0, nil
	//}
	//
	//operation.Message = fmt.Sprintf("Installation in progress: %s", installationState.Description)
	//
	//logger.Infof("Installation in progress: %s", installationState.Description)
	//return operation, 10 * time.Second, nil
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
