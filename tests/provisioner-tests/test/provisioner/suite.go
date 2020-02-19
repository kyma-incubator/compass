package provisioner

import (
	"errors"
	"fmt"
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

	TestRuntimes []TestRuntime
}

type TestRuntime struct {
	testSuite          *TestSuite
	provisioningInput  gqlschema.ProvisionRuntimeInput
	runtimeID          string
	isRunning          bool
	currentOperationID string
	status             []string
}

func (ts *TestSuite) NewRuntime(provisioningInput gqlschema.ProvisionRuntimeInput) *TestRuntime {
	testRuntime := TestRuntime{
		testSuite:         ts,
		provisioningInput: provisioningInput,
		isRunning:         false,
		status:            []string{"Initialized."},
	}
	ts.TestRuntimes = append(ts.TestRuntimes, testRuntime)
	return &testRuntime
}

func (r *TestRuntime) AddStatus(status string) {
	r.status = append(r.status, status)
}

func (r *TestRuntime) Provision() (operationStatusID, runtimeID string, err error) {
	r.AddStatus("Starting provisioning...")
	r.currentOperationID, r.runtimeID, err = r.testSuite.ProvisionerClient.ProvisionRuntime(r.provisioningInput)
	if err != nil {
		r.AddStatus(fmt.Sprintf("Error while provisioning Runtime: %s", err))
		return "", "", errors.New(r.GetCurrentStatus())
	}
	r.AddStatus("Provisioning started.")
	r.isRunning = true
	return r.currentOperationID, r.runtimeID, nil
}

func (r *TestRuntime) GetOperationStatus() (gqlschema.OperationStatus, error) {
	operationStatus, err := r.testSuite.ProvisionerClient.RuntimeOperationStatus(r.currentOperationID)
	if err != nil {
		r.AddStatus(fmt.Sprintf("Error while fetching Runtime Status: %s", err))
		return gqlschema.OperationStatus{}, errors.New(r.GetCurrentStatus())
	}
	r.AddStatus(fmt.Sprintf("%s: %s: %s", operationStatus.Operation, operationStatus.State, *operationStatus.Message))
	return operationStatus, nil
}

func (r *TestRuntime) Deprovision() (operationStatusID string, err error) {
	r.AddStatus("Starting deprovisioning...")
	r.currentOperationID, err = r.testSuite.ProvisionerClient.DeprovisionRuntime(r.runtimeID)
	if err != nil {
		r.AddStatus(fmt.Sprintf("Error while deprovisioning runtime: %s", err))
		return "", errors.New(r.GetCurrentStatus())
	}
	r.AddStatus("Deprovisioning started.")
	return r.currentOperationID, nil
}

func (r *TestRuntime) GetCurrentStatus() string {
	return r.status[len(r.status)-1]
}

func (r *TestRuntime) StatusToString() string {
	strStatus := ""
	for _, state := range r.status {
		strStatus += fmt.Sprintf("\t%s\n", state)
	}
	return strStatus
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
	if undeprovisionedRuntimes != nil {
		for _, runtime := range undeprovisionedRuntimes {
			logrus.Errorf("Error while performing cleanup: %s: %s", runtime.GetCurrentStatus(), runtime.StatusToString())
		}
		logrus.Infof("Cleanup failed.")
		return
	}
	logrus.Infof("Cleanup completed.")
}

func (ts *TestSuite) EnsureRuntimeDeprovisioning() []TestRuntime {
	failedRuntimes := []TestRuntime{}
	for _, runtime := range ts.TestRuntimes {
		if runtime.isRunning {
			go func(runtime TestRuntime) {
				for i := 0; i < 5; i++ {
					if runtime.isRunning {
						runtime.AddStatus("Starting deprovisioning...")
						operationID, err := runtime.Deprovision()
						if err != nil {
							runtime.AddStatus(fmt.Sprintf("Starting deprovisioning failed: %s", err))
							logrus.Infof("Failed to start deprovisioning Runtime '%s': %s: %s", runtime.runtimeID, operationID, err)
							continue
						}
						runtime.AddStatus("Deprovisioning started.")
						logrus.Infof("Deprovisioning Runtime '%s' started: %s", runtime.runtimeID, operationID)

						operationStatus, err := ts.WaitUntilOperationIsFinished(30*time.Minute, operationID)
						if err != nil {
							runtime.AddStatus(fmt.Sprintf("Deprovisioning failed: %s. State: %s", err, operationStatus.State))
							logrus.Infof("Deprovisioning Runtime '%s' failed: %s. State: %s", runtime.runtimeID, err, operationStatus.State)
							continue
						}
						runtime.AddStatus("Deprovisioning completed.")
						logrus.Infof("Deprovisioning Runtime '%s' completed.", runtime.runtimeID)
						runtime.isRunning = false
						break
					}
				}
				runtime.AddStatus("Failed to deprovision Runtime.")
				failedRuntimes = append(failedRuntimes, runtime)
			}(runtime)
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
