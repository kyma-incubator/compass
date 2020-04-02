package gardener

import (
	"fmt"
	gardencorev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
	"time"
)

func (r *ProvisioningOperator) ProvisioningInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	lastOperation := shoot.Status.LastOperation

	if lastOperation.State == gardencorev1alpha1.LastOperationStateSucceeded {
		err := r.ProceedToInstallation(log, shoot, operationId)
		if err != nil {
			log.Errorf("Error proceeding to installation: %s", err.Error())
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if isShootFailed(shoot) {
		log.Infof("Provisioning failed! Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)
		err := r.ProceedToFailedStep(log, shoot, operationId, "Provisioning failed.")
		if err != nil {
			log.Errorf("error proceeding to provisioning failed step: %s", err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	log.Debugf("Provisioning in progress. Last state: %s, Description: %s", lastOperation.State, lastOperation.Description)
	return ctrl.Result{}, nil
}

func (r *ProvisioningOperator) ProceedToInstallation(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) error {
	// Provisioning is finished. Start installation.
	log.Infof("Shoot provisioning finished.")

	session := r.dbsFactory.NewReadWriteSession()

	log.Infof("Getting cluster from DB")
	cluster, dberr := session.GetGardenerClusterByName(shoot.Name)
	if dberr != nil {
		return fmt.Errorf("error getting Gardener cluster by name: %s", dberr.Error())
	}

	log.Infof("Getting Kubeconfig")
	kubeconfig, err := FetchKubeconfigForShoot(r.secretsClient, shoot.Name)
	if err != nil {
		log.Errorf("Error fetching kubeconfig for Shoot: %s", err.Error())
		return err
	}

	dberr = session.UpdateCluster(cluster.ID, string(kubeconfig), nil)
	if dberr != nil {
		log.Errorf("Error saving kubeconfig in database: %s", dberr.Error())
		return dberr
	}

	// TODO: consider passing first step as param
	dberr = session.TransitionOperation(operationId, "Starting installation", model.StartingInstallation, time.Now())
	if dberr != nil {
		log.Errorf("Error transitioning operation stage: %s", dberr.Error())
		return dberr
	}

	log.Infof("Adding operation to installation queue")
	r.installationQueue.Add(operationId)

	log.Infof("Updating Shoot...")
	err = r.updateShoot(shoot, func(shootToUpdate *gardener_types.Shoot) {
		annotate(shootToUpdate, provisioningAnnotation, Provisioned.String())
		annotate(shootToUpdate, installationTimestampAnnotation, time.Now().Format(timeLayout)) // TODO: modify this to provisioning finished annotation?
		annotate(shootToUpdate, provisioningStepAnnotation, ProvisioningFinishedStep.String())
	})
	if err != nil {
		log.Errorf("Error updating Shoot with retries: %s", err.Error())
		return err
	}

	return nil
}
