package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

const (
	auditLogsAnnotation = "custom.shoot.sapcloud.io/subaccountId"
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
	Initial                  ProvisioningState = "initial"
	Deprovisioning           ProvisioningState = "deprovisioning"
	ProvisioningFailed       ProvisioningState = "failed"
)

type KymaInstallationState string

func (s KymaInstallationState) String() string {
	return string(s)
}

const (
	uninstallingAnnotation string = "compass.provisioner.kyma-project.io/uninstalling"

	operationIdAnnotation string = "compass.provisioner.kyma-project.io/operation-id"
	runtimeIdAnnotation   string = "compass.provisioner.kyma-project.io/runtime-id"
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

func getRuntimeId(shoot gardener_types.Shoot) string {
	runtimeId, found := shoot.Annotations[runtimeIdAnnotation]
	if !found {
		return ""
	}

	return runtimeId
}

func removeAnnotation(shoot *gardener_types.Shoot, annotation string) {
	if shoot.Annotations == nil {
		return
	}

	delete(shoot.Annotations, annotation)
}
