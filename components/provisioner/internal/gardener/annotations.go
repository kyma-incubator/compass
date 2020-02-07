package gardener

import (
	"time"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

const (
	timeLayout = "2006-01-02T15:04:05.000Z"
)

type ProvisioningState string

func (s ProvisioningState) String() string {
	return string(s)
}

const (
	UnknownProvisioningState ProvisioningState = ""
	Provisioned              ProvisioningState = "provisioned"
	Provisioning             ProvisioningState = "provisioning"
	Deprovisioning           ProvisioningState = "deprovisioning"
	ProvisioningFailed       ProvisioningState = "failed"
)

type KymaInstallationState string

func (s KymaInstallationState) String() string {
	return string(s)
}

const (
	UnknownKymaInstallationState KymaInstallationState = ""
	Installed                    KymaInstallationState = "installed"
	Installing                   KymaInstallationState = "installing"
	Uninstalling                 KymaInstallationState = "uninstalling"
	Upgrading                    KymaInstallationState = "upgrading"
	InstallationFailed           KymaInstallationState = "failed"
)

const (
	provisioningAnnotation string = "compass.provisioner.kyma-project.io/provisioning"
	installationAnnotation string = "compass.provisioner.kyma-project.io/kyma-installation"

	installationTimestampAnnotation string = "compass.provisioner.kyma-project.io/kyma-installation-timestamp"

	operationIdAnnotation string = "compass.provisioner.kyma-project.io/operation-id"
	runtimeIdAnnotation   string = "compass.provisioner.kyma-project.io/runtime-id"

	provisioningStepAnnotation string = "compass.provisioner.kyma-project.io/provisioning-step"
)

type ProvisioningStep string

func (s ProvisioningStep) String() string {
	return string(s)
}

const (
	UnknownProvisioningStep      ProvisioningStep = ""
	ProvisioningInProgressStep   ProvisioningStep = "provisioning-in-progress"
	InstallationInProgressStep   ProvisioningStep = "installation-in-progress"
	ProvisioningFinishedStep     ProvisioningStep = "provisioning-finished"
	ProvisioningFailedStep       ProvisioningStep = "failed"
	DeprovisioningInProgressStep ProvisioningStep = "deprovisioning"
)

func annotate(shoot *gardener_types.Shoot, annotation, value string) {
	if shoot.Annotations == nil {
		shoot.Annotations = map[string]string{}
	}

	shoot.Annotations[annotation] = value
}

func getOperationId(shoot gardener_types.Shoot) string {
	operationId, found := shoot.Annotations[operationIdAnnotation]
	if !found {
		return ""
	}

	return operationId
}

func getProvisioningState(shoot gardener_types.Shoot) ProvisioningState {
	provisioningState, found := shoot.Annotations[provisioningAnnotation]
	if !found {
		return UnknownProvisioningState
	}

	switch ProvisioningState(provisioningState) {
	case Provisioning, Provisioned, Deprovisioning, ProvisioningFailed:
		return ProvisioningState(provisioningState)
	default:
		return UnknownProvisioningState
	}
}

func getProvisioningStep(shoot gardener_types.Shoot) ProvisioningStep {
	provisioningStep, found := shoot.Annotations[provisioningStepAnnotation]
	if !found {
		return UnknownProvisioningStep
	}

	switch ProvisioningStep(provisioningStep) {
	case ProvisioningInProgressStep, InstallationInProgressStep, ProvisioningFailedStep, ProvisioningFinishedStep, DeprovisioningInProgressStep:
		return ProvisioningStep(provisioningStep)
	default:
		return UnknownProvisioningStep
	}
}

func getInstallationState(shoot gardener_types.Shoot) KymaInstallationState {
	installationState, found := shoot.Annotations[installationAnnotation]
	if !found {
		return UnknownKymaInstallationState
	}

	switch KymaInstallationState(installationState) {
	case Installed, Installing, Uninstalling, Upgrading, InstallationFailed:
		return KymaInstallationState(installationState)
	default:
		return UnknownKymaInstallationState
	}
}

func removeAnnotation(shoot *gardener_types.Shoot, annotation string) {
	if shoot.Annotations == nil {
		return
	}

	delete(shoot.Annotations, annotation)
}

func getInstallationTimestamp(shoot gardener_types.Shoot) (time.Time, error) {
	timeStamp, found := shoot.Annotations[installationTimestampAnnotation]
	if !found {
		return time.Time{}, nil
	}

	return time.Parse(timeLayout, timeStamp)
}
