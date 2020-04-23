package failure

import "github.com/kyma-incubator/compass/components/provisioner/internal/model"

type NoopFailureHandler struct {
}

func NewNoopFailureHandler() *NoopFailureHandler {
	return &NoopFailureHandler{}
}

func (u NoopFailureHandler) HandleFailure(operation model.Operation, cluster model.Cluster) error {
	return nil
}
