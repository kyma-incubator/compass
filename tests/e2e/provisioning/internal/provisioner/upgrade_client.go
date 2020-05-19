package provisioner

import (
	"time"

	"github.com/avast/retry-go"
	schema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/director"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

func NewRuntimeUpgradeClient(
	directorClient *director.Client,
	provisionerClient *Client,
	componentsProvider *ComponentsListProvider,
	tenantID,
	instanceID string,
	logger logrus.FieldLogger) *RuntimeUpgradeClient {
	return &RuntimeUpgradeClient{
		directorClient:     directorClient,
		provisionerClient:  provisionerClient,
		componentsProvider: componentsProvider,
		tenantID:           tenantID,
		instanceID:         instanceID,
		log:                logger,
	}
}

type RuntimeUpgradeClient struct {
	directorClient     *director.Client
	provisionerClient  *Client
	componentsProvider *ComponentsListProvider
	tenantID           string
	instanceID         string
	log                logrus.FieldLogger
}

func (c RuntimeUpgradeClient) UpgradeRuntimeToVersion(kymaVersion string) (string, error) {
	runtimeID, err := c.directorClient.GetRuntimeID(c.tenantID, c.instanceID)
	if err != nil {
		return "", errors.Wrap(err, "error while upgrading Runtime")
	}

	var components []v1alpha1.KymaComponent
	err = retry.Do(func() error {
		var err error
		components, err = c.componentsProvider.AllComponents(kymaVersion)
		if err != nil {
			return err
		}
		return nil
	})

	componentsInput := mapComponentsToGQLInput(components)

	upgradeInput := schema.UpgradeRuntimeInput{
		KymaConfig: &schema.KymaConfigInput{
			Version:    kymaVersion,
			Components: componentsInput,
		},
	}

	operationStatus, err := c.provisionerClient.UpgradeRuntime(runtimeID, upgradeInput)
	if err != nil {
		return "", errors.Wrap(err, "failed to upgrade Runtime to version")
	}

	if operationStatus.ID == nil {
		return "", errors.New("error upgrade operation ID is nil")
	}

	return *operationStatus.ID, nil
}

func (c RuntimeUpgradeClient) AwaitOperationFinished(operationId string, timeout time.Duration) error {
	c.log.Infof("Waiting for operation at most %s", timeout.String())

	err := wait.Poll(2*time.Minute, timeout, func() (bool, error) {
		operationStatus, err := c.provisionerClient.RuntimeOperationStatus(operationId)
		if err != nil {
			c.log.Warn(errors.Wrap(err, "while getting Runtime operation status").Error())
			return false, nil
		}

		c.log.Infof("Last operation status: %s", operationStatus.State)
		switch operationStatus.State {
		case schema.OperationStateSucceeded:
			c.log.Infof("Operation succeeded!")
			return true, nil
		case schema.OperationStateInProgress:
			return false, nil
		case schema.OperationStateFailed:
			c.log.Info("Operation failed!")
			return true, errors.Errorf("upgrade failed with message: %s", unwrapStr(operationStatus.Message))
		default:
			return false, nil
		}
	})
	if err != nil {
		return errors.Wrap(err, "while waiting for succeeded last operation")
	}
	return nil
}

func mapComponentsToGQLInput(kymaComponents []v1alpha1.KymaComponent) []*schema.ComponentConfigurationInput {
	componentsGQLInput := make([]*schema.ComponentConfigurationInput, len(kymaComponents))
	for i, component := range kymaComponents {
		var sourceURL *string
		if component.Source != nil {
			sourceURL = &component.Source.URL
		}

		componentsGQLInput[i] = &schema.ComponentConfigurationInput{
			Component: component.Name,
			Namespace: component.Namespace,
			SourceURL: sourceURL,
		}
	}
	return componentsGQLInput
}

func unwrapStr(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
