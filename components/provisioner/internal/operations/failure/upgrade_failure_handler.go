package failure

import "github.com/kyma-incubator/compass/components/provisioner/internal/model"

type UpgradeFailureHandler struct {
}

func NewUpgradeFailureHandler() *UpgradeFailureHandler {
	return &UpgradeFailureHandler{}
}

func (u UpgradeFailureHandler) HandleFailure(operation model.Operation, cluster model.Cluster) error {

	// TODO: Set last runtime upgrade to failed state

	return nil
}
