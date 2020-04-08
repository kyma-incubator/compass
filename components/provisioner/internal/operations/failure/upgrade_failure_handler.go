package failure

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
)

type UpgradeFailureHandler struct {
	session dbsession.ReadWriteSession
}

func NewUpgradeFailureHandler(session dbsession.ReadWriteSession) *UpgradeFailureHandler {
	return &UpgradeFailureHandler{
		session: session,
	}
}

func (u UpgradeFailureHandler) HandleFailure(operation model.Operation, _ model.Cluster) error {
	return u.session.UpdateUpgradeState(operation.ID, model.UpgradeFailed)
}
