package gardener

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	testEnv = &envtest.Environment{}

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
