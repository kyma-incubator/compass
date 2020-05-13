package deprovisioning

import (
	"time"

	"github.com/pkg/errors"

	shootUtil "github.com/kyma-incubator/compass/components/provisioner/internal/gardener/shoot"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type TriggerKymaUninstallStep struct {
	installationClient installation.Service
	gardenerClient     GardenerClient
	secretsClient      v1core.SecretInterface
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewTriggerKymaUninstallStep(installationClient installation.Service, gardenerClient GardenerClient, nextStep model.OperationStage, timeLimit time.Duration) *TriggerKymaUninstallStep {
	return &TriggerKymaUninstallStep{
		installationClient: installationClient,
		gardenerClient:     gardenerClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *TriggerKymaUninstallStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *TriggerKymaUninstallStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *TriggerKymaUninstallStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	logger.Debug("Shoot is on deprovisioning in progress step")

	gardenerConfig, ok := cluster.GardenerConfig()
	if !ok {
		// Non recoverable error?
		return operations.StageResult{}, errors.New("failed to read GardenerConfig")
	}

	shoot, err := s.gardenerClient.Get(gardenerConfig.Name, v1.GetOptions{})
	if err != nil {
		return operations.StageResult{}, err
	}

	logger.Debugf("Starting Uninstall")
	k8sConfig, err := shootUtil.KubeconfigForShoot(s.secretsClient, shoot.Name)
	if err != nil {
		logger.Errorf("error fetching kubeconfig: %s", err.Error())
		return operations.StageResult{}, err
	}

	err = s.installationClient.TriggerUninstall(k8sConfig)
	if err != nil {
		logger.Errorf("error triggering uninstalling: %s", err.Error())
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 30 * time.Second}, nil
}
