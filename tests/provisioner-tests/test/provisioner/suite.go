package provisioner

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
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

	TestRuntimes []TestRuntime
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
	// TODO(@rafalpotempa): Fetching provisioner logs when tests fail
	logrus.Infof("Starting cleanup...")

	undeprovisionedRuntimes := ts.EnsureRuntimeDeprovisioning()
	if len(undeprovisionedRuntimes) > 0 {
		for _, runtime := range undeprovisionedRuntimes {
			logrus.Errorf("Error while performing cleanup: %s: %s", runtime.GetCurrentStatus(), runtime.StatusToString())
		}
		logrus.Infof("Cleanup failed.")
		return
	}
	logrus.Infof("Cleanup completed.")
}

func (ts *TestSuite) Recover() {
	if r := recover(); r != nil {
		logrus.Warn("Recovered after panic signal: ", r)
	}
}

func (ts *TestSuite) EnsureRuntimeDeprovisioning() []TestRuntime {
	failedRuntimes := []TestRuntime{}
	var wg sync.WaitGroup

	for _, runtime := range ts.TestRuntimes {
		go func(runtime TestRuntime) {
			wg.Add(1)
			defer wg.Done()

			if runtime.isRunning {
				runtime.LogStatus("Starting deprovisioning...")
				operationID, err := runtime.Deprovision()
				runtime.log = runtime.log.WithFields(logrus.Fields{
					"RuntimeID":   runtime.runtimeID,
					"OperationID": operationID,
				})
				if err != nil {
					runtime.LogStatus(fmt.Sprintf("Starting deprovisioning failed: %s", err))
				}
				runtime.LogStatus("Deprovisioning started.")

				operationStatus, err := ts.WaitUntilOperationIsFinished(30*time.Minute, operationID)
				if err != nil {
					runtime.LogStatus(fmt.Sprintf("Deprovisioning failed: %s. State: %s", err, operationStatus.State))
				}
				runtime.LogStatus("Deprovisioning completed.")
				runtime.isRunning = false
				return
			}
			runtime.LogStatus("Failed to deprovision Runtime.")
		}(runtime)
	}
	wg.Wait()

	for _, runtime := range ts.TestRuntimes {
		if runtime.isRunning {
			failedRuntimes = append(failedRuntimes, runtime)
		}
	}
	return failedRuntimes
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
			logrus.Infof("Operation '%s': %s in progress", operationStatus.Operation, operationID)
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
