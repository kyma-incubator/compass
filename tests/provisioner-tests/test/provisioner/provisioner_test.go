package provisioner

import (
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/require"
)

const (
	gcpMachineType = "n1-standard-4"
	gcpRegion      = "europe-west4"
	gcpZone        = "europe-west4-b"
)

// TODO - decide whether to use logrus or t.Log

func Test_E2e(t *testing.T) {
	t.Logf("Starting tests. Test id: %s", testSuite.TestId)

	// Register runtime
	t.Logf("Registering runtime... Test id: %s", testSuite.TestId)
	runtimeInput := graphql.RuntimeInput{
		Name: "test-runtime-" + testSuite.TestId,
	}

	runtime, err := testSuite.DirectorClient.RegisterRuntime(runtimeInput)
	require.NoError(t, err)

	// Provision runtime
	zone := gcpZone

	provisioningInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{
				Name:              "tests-runtime-" + testSuite.TestId, // TODO - should be complient with cleaners
				KubernetesVersion: "1.14",
				NumberOfNodes:     3,
				BootDiskSize:      "30GB",
				MachineType:       gcpMachineType,
				Region:            gcpRegion,
				Zone:              &zone,
			},
		},
		Credentials: &gqlschema.CredentialsInput{SecretName: testSuite.CredentialsSecretName},
		KymaConfig:  nil,
	}

	t.Logf("Provsisioning runtime on GCP...")
	provisioningOperationId, err := testSuite.ProvisionerClient.ProvisionRuntime(runtime.ID, provisioningInput)
	require.NoError(t, err)
	t.Logf("Provisioning operation id: %s", provisioningOperationId)

	var provisioningOperationStatus gqlschema.OperationStatus
	err = testkit.RunParallelToMainFunction(ProvisioningTimeout+5*time.Second,
		func() error {
			t.Log("Waiting for provisioning to finish...")
			var waitErr error
			provisioningOperationStatus, waitErr = testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, provisioningOperationId)
			return waitErr
		},
		func() error {
			t.Log("Checking if operation will fail while other in progress...")
			operationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(provisioningOperationId)
			if err != nil {
				return errors.WithMessagef(err, "Failed to get %s operation status", provisioningOperationId)
			}

			if operationStatus.State != gqlschema.OperationStateInProgress {
				return errors.New("Operation %s not in progress")
			}

			_, err = testSuite.ProvisionerClient.ProvisionRuntime(runtime.ID, provisioningInput)
			if err == nil {
				return errors.New("Operation scheduled successfully while other operation in progress")
			}

			return nil
		},
	)
	require.NoError(t, err)

	assertOperationSucceed(t, gqlschema.OperationTypeProvision, runtime.ID, provisioningOperationStatus)

	// Get Kubeconfig
	t.Logf("Fetching runtime status...")
	runtimeStatus, err := testSuite.ProvisionerClient.RuntimeStatus(runtime.ID)
	require.NoError(t, err)
	assertGCPRuntimeConfiguration(t, provisioningInput, runtimeStatus)

	t.Logf("Preparing K8s client...")
	k8sClient := testSuite.KubernetesClientFromRawConfig(t, *runtimeStatus.RuntimeConfiguration.Kubeconfig)

	t.Logf("Accessing API Server on provisioned cluster...")
	version, err := k8sClient.ServerVersion()
	require.NoError(t, err)

	// TODO - make sure it will work
	assert.Equal(t, provisioningInput.ClusterConfig.GcpConfig.KubernetesVersion, version.Major)

	// TODO- HERE - Run Compass Runtime Agent Tests (it may require passing AccessToken or Credentials for Director?) (Maybe pass credentials and tests will only generate Access Token?)

	t.Logf("Deprovisioning runtime...")
	deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtime.ID)
	require.NoError(t, err)
	t.Logf("Deprovisioning operation id: %s", deprovisioningOperationId)

	deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, deprovisioningOperationId)
	require.NoError(t, err)
	assertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtime.ID, deprovisioningOperationStatus)
}

func assertGCPRuntimeConfiguration(t *testing.T, input gqlschema.ProvisionRuntimeInput, status gqlschema.RuntimeStatus) {
	require.NotNil(t, status.RuntimeConfiguration)
	require.NotNil(t, status.RuntimeConfiguration.ClusterConfig)
	require.NotNil(t, status.RuntimeConfiguration.Kubeconfig)
	//require.NotNil(t, status.RuntimeConfiguration.KymaConfig) // TODO - uncomment when implemented

	require.NotNil(t, status.LastOperationStatus)

	//require.NotNil(t, status.RuntimeConnectionStatus) // TODO - uncomment when implemented

	gcpClusterConfig, ok := status.RuntimeConfiguration.ClusterConfig.(*gqlschema.GCPConfig)
	require.True(t, ok)

	assert.Equal(t, input.ClusterConfig.GcpConfig.Name, gcpClusterConfig.Name)
	assert.Equal(t, input.ClusterConfig.GcpConfig.Region, gcpClusterConfig.Region)
	assert.Equal(t, input.ClusterConfig.GcpConfig.KubernetesVersion, gcpClusterConfig.KubernetesVersion)
	assert.Equal(t, input.ClusterConfig.GcpConfig.BootDiskSize, gcpClusterConfig.BootDiskSize)
	assert.Equal(t, input.ClusterConfig.GcpConfig.MachineType, gcpClusterConfig.MachineType)
	assert.Equal(t, input.ClusterConfig.GcpConfig.NumberOfNodes, gcpClusterConfig.NumberOfNodes)
	assert.Equal(t, input.ClusterConfig.GcpConfig.Zone, gcpClusterConfig.Zone)
}

func assertOperationFailed(t *testing.T, expectedType gqlschema.OperationType, expectedRuntimeId string, operation gqlschema.OperationStatus) {
	assertOperation(t, gqlschema.OperationStateFailed, expectedType, expectedRuntimeId, operation)
}

func assertOperationSucceed(t *testing.T, expectedType gqlschema.OperationType, expectedRuntimeId string, operation gqlschema.OperationStatus) {
	assertOperation(t, gqlschema.OperationStateSucceeded, expectedType, expectedRuntimeId, operation)
}

func assertOperation(t *testing.T, expectedState gqlschema.OperationState, expectedType gqlschema.OperationType, expectedRuntimeId string, operation gqlschema.OperationStatus) {
	t.Logf("Assering operation %s is in %s state.", "", expectedState) // TODO - pass operation ID here (modify the API)
	t.Logf("Operation message: %s", operation.Message)
	require.Equal(t, expectedState, operation.State)
	assert.Equal(t, expectedRuntimeId, operation.RuntimeID)
	assert.Equal(t, expectedType, operation.Operation)
}
