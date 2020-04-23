package gardener

import (
	"testing"

	gardener_types "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_annotate(t *testing.T) {
	// given
	shoot := &gardener_types.Shoot{
		ObjectMeta: v1.ObjectMeta{Name: clusterName, Namespace: gardenerNamespace},
	}

	// when
	annotate(shoot, runtimeIdAnnotation, "abcd-efgh")

	// then
	assertAnnotation(t, shoot, runtimeIdAnnotation, "abcd-efgh")
}
