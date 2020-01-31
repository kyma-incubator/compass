package gardener

import (
	"errors"
	"fmt"
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/garden/v1beta1"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ProvisioningOperator) ProceedToInstallation(log *logrus.Entry, shoot gardener_types.Shoot) error {
	// Provisioning is finished. Start installation.
	log.Infof("Provisioning finished. Starting installation...")

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

	// TODO: this operation block for few seconds, we can consider making it non blocking later (with changes to Installation SDK)
	log.Infof("Triggering installation")
	err = r.installationService.TriggerInstallation(kubeconfig, cluster.KymaConfig.Release, cluster.KymaConfig.GlobalConfiguration, cluster.KymaConfig.Components)
	if err != nil {
		log.Errorf("Error triggering Kyma installation: %s", err.Error())
		return err
	}

	log.Infof("Updating Shoot...")
	err = r.updateShoot(shoot, func(shootToUpdate *gardener_types.Shoot) {
		annotate(shootToUpdate, provisioningAnnotation, Provisioned.String())
		annotate(shootToUpdate, installationAnnotation, Installing.String())
		annotate(shootToUpdate, installationTimestampAnnotation, time.Now().Format(timeLayout))
		annotate(shootToUpdate, provisioningStepAnnotation, InstallationInProgressStep.String())
	})
	if err != nil {
		log.Errorf("Error updating Shoot with retries: %s", err.Error())
		return err
	}

	log.Infof("Installation triggered on cluster.")
	return nil
}

func (r *ProvisioningOperator) InstallationInProgress(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	log.Infof("Shoot is on installation in progress step")

	rawKubeconfig, err := FetchKubeconfigForShoot(r.secretsClient, shoot.Name)
	if err != nil {
		log.Errorf("failed to get client kubeconfig when checking Installation state: %s", err.Error())
		return ctrl.Result{}, err
	}

	kubeconfig, err := ParseToK8sConfig(rawKubeconfig)
	if err != nil {
		log.Errorf("failed to parse kubeconfig when checking Installation state: %s", err.Error())
		return ctrl.Result{}, err
	}

	installationState, err := installationSDK.CheckInstallationState(kubeconfig)
	if err != nil {
		var installationErr installationSDK.InstallationError
		if !errors.As(err, &installationErr) {
			log.Errorf("failed to get installation state for shoot: %s", err.Error())
			return ctrl.Result{}, err
		}

		log.Infof("Installation Error: %s, Details: %s", installationErr.Error(), installationErr.Details())
		if len(installationErr.ErrorEntries) > r.installationErrorsThreshold {
			// Set as failed
			// TODO: Should we fail here or just wait for timeout?
		}

		return r.proceedToFailedStepIfTimeoutReached(log, shoot, operationId)
	}

	log.Infof("Installation State: %s, Description: %s", installationState.State, installationState.Description)

	if installationState.State == "Installed" {
		err := r.configureRuntime(log, shoot, operationId, string(rawKubeconfig))
		if err != nil {
			log.Errorf("Failed to configure Runtime: %s", err.Error())
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	return r.proceedToFailedStepIfTimeoutReached(log, shoot, operationId)
}

func (r *ProvisioningOperator) configureRuntime(log *logrus.Entry, shoot gardener_types.Shoot, operationId, kubeconfig string) error {
	cluster, dberr := r.dbsFactory.NewReadSession().GetGardenerClusterByName(shoot.Name)
	if dberr != nil {
		return fmt.Errorf("error configuring Runtime: %s", dberr.Error())
	}

	err := r.runtimeConfigurator.ConfigureRuntime(cluster, kubeconfig)
	if err != nil {
		return fmt.Errorf("error configuring Runtime: %s", err.Error())
	}
	log.Infof("Runtime %s configured", cluster.ID)

	return r.ProceedToFinishedStep(log, shoot, operationId)
}

func (r *ProvisioningOperator) proceedToFailedStepIfTimeoutReached(log *logrus.Entry, shoot gardener_types.Shoot, operationId string) (ctrl.Result, error) {
	installationTimestamp, err := getInstallationTimestamp(shoot)
	if err != nil {
		log.Errorf("error: failed to parse installation timestamp: %s", err.Error())
		return ctrl.Result{}, err
	}

	now := time.Now()

	if now.Sub(installationTimestamp) > r.kymaInstallationTimeout {
		log.Infof("Installation timeout reached. Setting operation as failed.")
		err := r.ProceedToFailedStep(log, shoot, operationId, "Timeout waiting for Kyma to install.")
		if err != nil {
			log.Errorf("error proceeding to Provisioning Failed Step: %s", err.Error())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil // TODO: consider to requeue after - time left for installation timeout?
}
