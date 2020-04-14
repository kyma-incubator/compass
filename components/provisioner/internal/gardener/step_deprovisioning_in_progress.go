package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) DeprovisioningInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	log.Infof("Shoot is on deprovisioning in progress step")

	return ctrl.Result{}, nil
}
