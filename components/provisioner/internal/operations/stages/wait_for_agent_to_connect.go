package stages

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	compass_conn_clientset "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	defaultCompassConnectionName = "compass-connection"
)

type CompassConnectionClientConstructor func(k8sConfig *rest.Config) (compass_conn_clientset.CompassConnectionInterface, error)

func NewCompassConnectionClient(k8sConfig *rest.Config) (compass_conn_clientset.CompassConnectionInterface, error) {
	compassConnClientset, err := compass_conn_clientset.NewForConfig(k8sConfig)
	if err != nil {
		return nil, fmt.Errorf("error: failed to create Compass Connection client: %s", err.Error())
	}

	return compassConnClientset.CompassConnections(), nil
}

type WaitForAgentToConnectStep struct {
	newCompassConnectionClient CompassConnectionClientConstructor
	directorClient             director.DirectorClient
	nextStep                   model.OperationStage
	timeLimit                  time.Duration
}

func NewWaitForAgentToConnectStep(
	ccClientProvider CompassConnectionClientConstructor,
	nextStep model.OperationStage,
	timeLimit time.Duration,
	directorClient director.DirectorClient) *WaitForAgentToConnectStep {

	return &WaitForAgentToConnectStep{
		newCompassConnectionClient: ccClientProvider,
		directorClient:             directorClient,
		nextStep:                   nextStep,
		timeLimit:                  timeLimit,
	}
}

func (s *WaitForAgentToConnectStep) Name() model.OperationStage {
	return model.WaitForAgentToConnect
}

func (s *WaitForAgentToConnectStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForAgentToConnectStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	compassConnClient, err := s.newCompassConnectionClient(k8sConfig)
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to create Compass Connection client: %s", err.Error())
	}

	compassConnCR, err := compassConnClient.Get(defaultCompassConnectionName, v1meta.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.Infof("Compass Connection not yet found on cluster")
			return operations.StageResult{Stage: s.Name(), Delay: 5 * time.Second}, nil
		}

		return operations.StageResult{}, fmt.Errorf("error getting Compass Connection CR on the Runtime: %s", err.Error())
	}

	if compassConnCR.Status.State == v1alpha1.ConnectionFailed {
		return operations.StageResult{}, fmt.Errorf("error: Compass Connection is in Failed state")
	}

	if compassConnCR.Status.State != v1alpha1.Synchronized {
		if compassConnCR.Status.State == v1alpha1.SynchronizationFailed {
			logger.Warnf("Runtime Agent Connected but resource synchronization failed state: %s", compassConnCR.Status.State)
			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		}
		if compassConnCR.Status.State == v1alpha1.MetadataUpdateFailed {
			logger.Warnf("Runtime Agent Connected but metadata update failed: %s", compassConnCR.Status.State)
			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		}

		logger.Infof("Compass Connection not yet in Synchronized state, current state: %s", compassConnCR.Status.State)
		return operations.StageResult{Stage: s.Name(), Delay: 2 * time.Second}, nil
	}

	if err := s.directorClient.SetRuntimeStatusCondition(cluster.ID, gqlschema.RuntimeStatusConditionConnected, cluster.Tenant); err != nil {
		logger.Errorf("Failed to set runtime %s status condition: %s", gqlschema.RuntimeStatusConditionConnected.String(), err.Error())
		return operations.StageResult{Stage: s.Name(), Delay: 2 * time.Second}, nil
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
