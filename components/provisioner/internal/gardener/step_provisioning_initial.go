package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) ProvisioningInitial(
	log *logrus.Entry,
	shoot gardener_types.Shoot,
	operationId, runtimeId string) (ctrl.Result, error) {

	if shoot.Spec.DNS == nil || shoot.Spec.DNS.Domain == nil {
		log.Warn("DNS Domain is not set yet for runtime ID: %s", runtimeId)
		return ctrl.Result{}, nil
	}

	log.Infof("Updating in Director...")
	session := r.dbsFactory.NewReadSession()
	tenant, dberr := session.GetTenant(runtimeId)
	if dberr != nil {
		log.Errorf("Error getting Gardener cluster by name: %s", dberr.Error())
		return ctrl.Result{}, dberr
	}
	runtime, err := r.directorClient.GetRuntime(runtimeId, tenant)
	if err != nil {
		log.Errorf("Error getting Runtime by ID: %s", err.Error())
		return ctrl.Result{}, err
	}
	labels := gqlschema.Labels{
		"gardenerClusterName":   shoot.ObjectMeta.Name,
		"gardenerClusterDomain": *shoot.Spec.DNS.Domain,
	}
	statusCondition := gqlschema.RuntimeStatusConditionProvisioning
	runtimeInput := &gqlschema.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		Labels:          &labels,
		StatusCondition: &statusCondition,
	}
	if err := r.directorClient.UpdateRuntime(runtimeId, runtimeInput, tenant); err != nil {
		log.Errorf("Error updating Runtime in Director: %s", err.Error())
		return ctrl.Result{}, err
	}

	log.Infof("Updating Shoot...")
	err = r.updateShoot(shoot, func(shootToUpdate *gardener_types.Shoot) {
		annotate(shootToUpdate, provisioningAnnotation, Provisioning.String())
	})
	if err != nil {
		log.Errorf("Error updating Shoot with retries: %s", err.Error())
		return ctrl.Result{}, err
	}

	return r.ProvisioningInProgress(log, shoot, operationId, runtimeId)
}
