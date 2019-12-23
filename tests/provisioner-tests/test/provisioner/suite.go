package provisioner

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1client "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/provisioner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	provisionerCredentialsSecretKey = "credentials"

	ProvisioningTimeout   = 25 * time.Minute
	DeprovisioningTimeout = 15 * time.Minute

	checkInterval = 2 * time.Second
)

var providers = []string{
	"azure",
	// "aws",
	// "gcp",
}

type TestSuite struct {
	TestId            string
	ProvisionerClient provisioner.Client

	CredentialsSecretName string

	providers []string

	config        testkit.TestConfig
	secretsClient v1client.SecretInterface
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {
	rand.Seed(time.Now().UnixNano())

	// TODO: Sleep ensures that the Istio Sidecar is up before running the tests. We can consider adding some health endpoint in the service to avoid hardcoded sleep.
	time.Sleep(15 * time.Second)

	k8sConfig, err := getK8sConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get K8s config")
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	provisionerClient := provisioner.NewProvisionerClient(config.InternalProvisionerURL, config.QueryLogging)

	testId := randStringBytes(8)

	return &TestSuite{
		TestId:            testId,
		ProvisionerClient: provisionerClient,

		CredentialsSecretName: fmt.Sprintf("tests-cred-%s", testId),

		providers: providers,

		config:        config,
		secretsClient: k8sClient.CoreV1().Secrets(config.CredentialsNamespace),
	}, nil
}

func (ts *TestSuite) Setup() error {
	logrus.Infof("Setting up environment")

	err := ts.saveCredentialsToSecret(ts.config.GCPCredentials)
	if err != nil {
		return errors.WithMessagef(err, "Failed to save credentials to %s secret", ts.CredentialsSecretName)
	}

	return nil
}

func (ts *TestSuite) Cleanup() {
	logrus.Infof("Starting cleanup...")

	logrus.Infof("Removing credentials secret %s ...", ts.CredentialsSecretName)
	err := ts.removeCredentialsSecret()
	if err != nil {
		logrus.Warnf("Failed to remove credentials secret: %s", err.Error())
	}
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

func (ts *TestSuite) saveCredentialsToSecret(credentials string) error {
	decodedCredentials, err := base64.StdEncoding.DecodeString(ts.config.GCPCredentials)
	if err != nil {
		return errors.Errorf("Failed to decode credentials from base64: %s", err.Error())
	}

	logrus.Infof("Creating credentials secret %s ...", ts.CredentialsSecretName)
	_, err = ts.secretsClient.Create(&v1.Secret{
		ObjectMeta: v1meta.ObjectMeta{Name: ts.CredentialsSecretName},
		Data: map[string][]byte{
			provisionerCredentialsSecretKey: decodedCredentials,
		},
	})
	if err != nil {
		return errors.Wrap(err, "Failed to save credentials to secret")
	}

	return nil
}

func (ts *TestSuite) removeCredentialsSecret() error {
	return ts.secretsClient.Delete(ts.CredentialsSecretName, &v1meta.DeleteOptions{})
}

func getK8sConfig() (*restclient.Config, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Info("Failed to read in cluster config, trying with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, err
		}
	}

	return k8sConfig, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyz123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
