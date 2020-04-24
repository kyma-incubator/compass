package gardener

import (
	"fmt"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) ProvisioningInitial(
	log *logrus.Entry,
	shoot gardener_types.Shoot,
	operationId, runtimeId string) (ctrl.Result, error) {

	if shoot.Spec.DNS == nil || shoot.Spec.DNS.Domain == nil {
		log.Warnf("DNS Domain is not set yet for runtime ID: %s", runtimeId)
		return ctrl.Result{}, nil
	}

	log.Infof("Updating in Director...")
	tenant, dberr := r.dbsFactory.NewReadSession().GetTenant(runtimeId)
	if dberr != nil {
		log.Errorf("Error getting Gardener cluster by name: %s", dberr.Error())
		return ctrl.Result{}, dberr
	}
	// TODO: Consider updating Labels and StatusCondition separately without getting the Runtime
	//       It'll be possible after this issue implementation:
	//       - https://github.com/kyma-incubator/compass/issues/1186
	runtimeInput, err := r.prepareProvisioningUpdateRuntimeInput(runtimeId, tenant, shoot)
	if err != nil {
		log.Errorf("Error preparing Runtime Input: %s", err.Error())
		return ctrl.Result{}, err
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

func (r *ProvisioningOperator) prepareProvisioningUpdateRuntimeInput(
	runtimeId, tenant string, shoot gardener_types.Shoot) (*gqlschema.RuntimeInput, error) {

	runtime, err := r.directorClient.GetRuntime(runtimeId, tenant)
	if err != nil {
		return &gqlschema.RuntimeInput{}, errors.Wrap(err, fmt.Sprintf("failed to get Runtime by ID: %s", runtimeId))
	}
	labels := gqlschema.Labels{
		"gardenerClusterName":   shoot.ObjectMeta.Name,
		"gardenerClusterDomain": *shoot.Spec.DNS.Domain,
	}
	for key, value := range runtime.Labels {
		labels[key] = value
	}
	statusCondition := gqlschema.RuntimeStatusConditionProvisioning
	runtimeInput := &gqlschema.RuntimeInput{
		Name:            runtime.Name,
		Description:     runtime.Description,
		Labels:          &labels,
		StatusCondition: &statusCondition,
	}
	return runtimeInput, nil
}
