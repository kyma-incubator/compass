package provisioner

import (
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
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/director"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/provisioner"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/oauth"
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
	provisionerCredentialsSecretKey = "credentials.json"

	ProvisioningTimeout   = 25 * time.Minute
	DeprovisioningTimeout = 15 * time.Minute

	checkInterval = 2 * time.Second
)

type TestSuite struct {
	TestId            string
	ProvisionerClient provisioner.Client
	DirectorClient    *director.Client

	CredentialsSecretName string

	config        testkit.TestConfig
	secretsClient v1client.SecretInterface
}

func NewTestSuite(config testkit.TestConfig) (*TestSuite, error) {

	// TODO - need some endpoint to check if sidecar is up

	k8sConfig, err := getK8sConfig()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get K8s config")
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	oauthCredentials, err := oauth.RegisterClient(config.HydraAdminURL)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to register OAuth client")
	}

	oauthTokenClient := oauth.NewOauthTokensClient(config.HydraPublicURL, oauthCredentials)

	provisionerClient := provisioner.NewProvisionerClient(config.InternalProvisionerURL, config.QueryLogging)
	directorClient := director.NewDirectorClient(config.DirectorURL, config.Tenant, oauthTokenClient, config.QueryLogging)

	testId := randStringBytes(8)

	return &TestSuite{
		TestId:            testId,
		ProvisionerClient: provisionerClient,
		DirectorClient:    directorClient,

		config: config,

		secretsClient:         k8sClient.CoreV1().Secrets(config.CredentialsNamespace),
		CredentialsSecretName: fmt.Sprintf("tests-cred-%s", testId),
	}, nil
}

func (ts *TestSuite) Setup() error {
	err := ts.saveCredentialsToSecret(ts.config.GCPCredentials)
	if err != nil {
		return errors.WithMessagef(err, "Failed to save credentials to %s secret", ts.CredentialsSecretName)
	}

	return nil
}

func (ts *TestSuite) Cleanup() {
	err := ts.removeCredentialsSecret()
	if err != nil {
		logrus.Warnf("Failed to remove credentials secret: %s", err.Error())
	}
}

func (ts *TestSuite) ProvisionRuntime(t *testing.T, runtimeId string, input gqlschema.ProvisionRuntimeInput) gqlschema.OperationStatus {
	operationId, err := ts.ProvisionerClient.ProvisionRuntime(runtimeId, input)
	require.NoError(t, err)

	var provisioningOperationStatus gqlschema.OperationStatus
	err = testkit.RunParallelToMainFunction(ProvisioningTimeout+5*time.Second,
		func() error {
			t.Log("Waiting for provisioning to finish...")
			var waitErr error
			provisioningOperationStatus, waitErr = ts.WaitUntilOperationIsFinished(ProvisioningTimeout, operationId)
			return waitErr
		},
		func() error {
			t.Log("Should fail to schedule operation while other in progress.")
			operationStatus, err := ts.ProvisionerClient.RuntimeOperationStatus(operationId)
			if err != nil {
				return errors.WithMessagef(err, "Failed to get %s operation status", operationId)
			}

			if operationStatus.State != gqlschema.OperationStateInProgress {
				return errors.New("Operation %s not in progress")
			}

			_, err = ts.ProvisionerClient.ProvisionRuntime(runtimeId, input)
			if err == nil {
				return errors.New("Operation scheduled successfully while other operation in progress")
			}

			return nil
		},
	)
	require.NoError(t, err)

	return provisioningOperationStatus
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
	tempKubeconfigFile, err := ioutil.TempFile("tmp", "kubeconfig")
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
	_, err := ts.secretsClient.Create(&v1.Secret{
		ObjectMeta: v1meta.ObjectMeta{Name: ts.CredentialsSecretName},
		StringData: map[string]string{
			provisionerCredentialsSecretKey: credentials,
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
