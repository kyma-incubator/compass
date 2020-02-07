package provisioner

import (
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/provisioner"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ProvisioningTimeout   = 60 * time.Minute
	DeprovisioningTimeout = 60 * time.Minute

	checkInterval = 10 * time.Second

	Azure = testkit.Azure
	GCP   = testkit.GCP
	AWS   = testkit.AWS
)

type TestSuite struct {
	TestId            string
	ProvisionerClient provisioner.Client

	gardenerProviders []string

	config        testkit.TestConfig
	secretsClient v1client.SecretInterface
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {
	rand.Seed(time.Now().UnixNano())

	// TODO: Sleep ensures that the Istio Sidecar is up before running the tests. We can consider adding some health endpoint in the service to avoid hardcoded sleep.
	time.Sleep(15 * time.Second)

	provisionerClient := provisioner.NewProvisionerClient(config.InternalProvisionerURL, config.Tenant, config.QueryLogging)

	testId := randStringBytes(8)

	return &TestSuite{
		TestId:            testId,
		ProvisionerClient: provisionerClient,

		gardenerProviders: config.Gardener.Providers,

		config: config,
	}, nil
}

func (ts *TestSuite) Setup() error {
	logrus.Infof("Setting up environment")

	return nil
}

func (ts *TestSuite) Cleanup() {
	provisionerLogs := "https://github.com/kubernetes/kubernetes/blob/c2e90cd1549dff87db7941544ce15f4c8ad0ba4c/pkg/kubectl/cmd/log.go#L188"
	logrus.Infof(provisionerLogs)
	logrus.Infof("Starting cleanup...")

}

func (ts *TestSuite) WaitUntilOperationIsFinished(timeout time.Duration, operationID string) (gqlschema.OperationStatus, error) {
	var operationStatus gqlschema.OperationStatus
	var err error

	err = testkit.WaitForFunction(checkInterval, timeout, func() bool {
		operationStatus, err = ts.ProvisionerClient.RuntimeOperationStatus(operationID)
		if err != nil {
			logrus.Warnf("Failed to get operation status: %s", err.Error())
			return false
		}

		if operationStatus.State == gqlschema.OperationStateInProgress {
			logrus.Infof("Operation %s in progress", operationID)
			return false
		}

		return true
	})

	return operationStatus, err
}

func (ts *TestSuite) KubernetesClientFromRawConfig(t *testing.T, rawConfig string) *kubernetes.Clientset {
	tempKubeconfigFile, err := ioutil.TempFile("", "kubeconfig")
	require.NoError(t, err)
	defer func() {
		err := os.RemoveAll(tempKubeconfigFile.Name())
		if err != nil {
			logrus.Warnf("Failed to delete temporary Kubeconfig file: %s", err.Error())
		}
	}()

	_, err = tempKubeconfigFile.WriteString(rawConfig)
	require.NoError(t, err)

	k8sConfig, err := clientcmd.BuildConfigFromFlags("", tempKubeconfigFile.Name())
	require.NoError(t, err)
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	require.NoError(t, err)

	return k8sClient
}

func (ts *TestSuite) removeCredentialsSecret(secretName string) error {
	return ts.secretsClient.Delete(secretName, &v1meta.DeleteOptions{})
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
