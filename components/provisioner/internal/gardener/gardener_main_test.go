package gardener

import (
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var testEnv *envtest.Environment
var cfg *rest.Config

func TestMain(m *testing.M) {
	err := setupEnv()
	if err != nil {
		logrus.Errorf("Failed to setup test environment: %s", err.Error())
		os.Exit(1)
	}
	defer func() {
		err := testEnv.Stop()
		if err != nil {
			logrus.Errorf("error while deleting Compass Connection: %s", err.Error())
		}
	}()
	os.Exit(m.Run())
}

func setupEnv() error {
	testEnv = &envtest.Environment{
		CRDs: []*apiextensionsv1beta1.CustomResourceDefinition{
			{
				ObjectMeta: v1.ObjectMeta{Name: "shoots.core.gardener.cloud"},
				Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
					Group: "core.gardener.cloud",
					Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
						Plural:   "shoots",
						Singular: "shoot",
						Kind:     "Shoot",
					},
					Scope:        "",
					Validation:   nil,
					Subresources: nil,
					Versions: []apiextensionsv1beta1.CustomResourceDefinitionVersion{
						{
							Name:    "v1beta1",
							Storage: true,
						},
					},
					PreserveUnknownFields: util.BoolPtr(true),
				},
			},
		},
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		return errors.Wrap(err, "Failed to start test environment")
	}

	//err = v1alpha1.AddToScheme(scheme.Scheme)
	//if err != nil {
	//	return errors.Wrap(err, "Failed to add to schema")
	//}

	return nil
}

func Test(t *testing.T) {

}
