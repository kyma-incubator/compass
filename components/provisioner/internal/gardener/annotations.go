package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
)

const (
	auditLogsAnnotation = "custom.shoot.sapcloud.io/subaccountId"
)

type ProvisioningState string

func (s ProvisioningState) String() string {
	return string(s)
}

type KymaInstallationState string

func (s KymaInstallationState) String() string {
	return string(s)
}

const (
	operationIdAnnotation string = "compass.provisioner.kyma-project.io/operation-id"
	runtimeIdAnnotation   string = "compass.provisioner.kyma-project.io/runtime-id"
	licenceTypeAnnotation string = "compass.provisioner.kyma-project.io/licence-type"
)

func annotate(shoot *gardener_types.Shoot, annotation, value string) {
	if shoot.Annotations == nil {
		shoot.Annotations = map[string]string{}
	}

	shoot.Annotations[annotation] = value
}

func getRuntimeId(shoot gardener_types.Shoot) string {
	runtimeId, found := shoot.Annotations[runtimeIdAnnotation]
	if !found {
		return ""
	}

	return runtimeId
}
