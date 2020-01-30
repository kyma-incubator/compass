package gardener

import (
	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_annotate(t *testing.T) {
	// given
	shoot := &gardener_types.Shoot{
		ObjectMeta: v1.ObjectMeta{Name: clusterName, Namespace: gardenerNamespace},
	}

	// when
	annotate(shoot, provisioningStepAnnotation, ProvisioningInProgressStep.String())

	// then
	assertAnnotation(t, shoot, provisioningStepAnnotation, "provisioning-in-progress")
}
