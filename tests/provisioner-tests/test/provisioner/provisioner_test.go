package provisioner

import (
	"fmt"
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
	gcpRegion      = "europe-west4-b"
	gcpZone        = "europe-west4-b"
)

// TODO - decide whether to use logrus or t.Log

func Test_E2e(t *testing.T) {
	t.Logf("Starting tests. Test id: %s", testSuite.TestId)

	// Register runtime
	t.Logf("Registering runtime...")
	runtimeInput := graphql.RuntimeInput{
		Name: "test-runtime-" + testSuite.TestId,
	}

	runtime, err := testSuite.DirectorClient.RegisterRuntime(runtimeInput)
	requireNoError(t, err)
	t.Logf("Runtime registered successfully id: %s", runtime.ID)
	defer func() {
		// TODO - deleting runtime fails not sure why
		t.Logf("Removing %s runtime...", runtime.ID)
		_, err := testSuite.DirectorClient.DeleteRuntime(runtime.ID)
		assertNoError(t, err)
	}()

	// Provision runtime
	credentialsInput := gqlschema.CredentialsInput{SecretName: testSuite.CredentialsSecretName}

	provisioningInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{
				Name:              "tests-runtime-" + testSuite.TestId, // TODO - should be complient with cleaners
				ProjectName:       config.GCPProjectName,
				KubernetesVersion: "1.14",
				NumberOfNodes:     3,
				BootDiskSize:      "30",
				MachineType:       gcpMachineType,
				Region:            gcpRegion,
			},
		},
		Credentials: &credentialsInput,
		KymaConfig:  &gqlschema.KymaConfigInput{Version: "1.6", Modules: nil},
	}

	t.Logf("Provsisioning runtime on GCP...")
	provisioningOperationId, err := testSuite.ProvisionerClient.ProvisionRuntime(runtime.ID, provisioningInput)
	requireNoError(t, err)
	t.Logf("Provisioning operation id: %s", provisioningOperationId)
	defer func() {
		t.Logf("Ensuring the cluster is deprovisioned...")
		deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtime.ID, credentialsInput)
		if err != nil {
			t.Logf("Error while ensuring the cluster is deprovisioned (cluster might have been deprovisioned already): %s", err.Error())
			return
		}

		t.Logf("Deprovisioning operation id: %s", deprovisioningOperationId)
		deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, deprovisioningOperationId)
		if err != nil {
			t.Logf("Error while waiting for deprovisioning operation to finish: %s", err.Error())
			return
		}

		assertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtime.ID, deprovisioningOperationStatus)
		// TODO - force delete runtime data?
	}()

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
	requireNoError(t, err, "Provisioning operation status: ", provisioningOperationStatus.State)

	assertOperationSucceed(t, gqlschema.OperationTypeProvision, runtime.ID, provisioningOperationStatus)
	t.Logf("Runtime provisioned successfully")

	t.Logf("Fetching runtime status...")
	runtimeStatus, err := testSuite.ProvisionerClient.RuntimeStatus(runtime.ID)
	requireNoError(t, err)
	assertGCPRuntimeConfiguration(t, provisioningInput, runtimeStatus)

	t.Logf("Preparing K8s client...")
	k8sClient := testSuite.KubernetesClientFromRawConfig(t, *runtimeStatus.RuntimeConfiguration.Kubeconfig)

	t.Logf("Accessing API Server on provisioned cluster...")
	version, err := k8sClient.ServerVersion()
	requireNoError(t, err)

	// TODO - make sure it will work
	assert.Equal(t, provisioningInput.ClusterConfig.GcpConfig.KubernetesVersion, version.Major)

	// TODO - Run Compass Runtime Agent Tests - it may require passing Credentials for MP

	t.Logf("Deprovisioning runtime...")
	deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtime.ID, credentialsInput)
	requireNoError(t, err)
	t.Logf("Deprovisioning operation id: %s", deprovisioningOperationId)

	deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, deprovisioningOperationId)
	requireNoError(t, err)
	assertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtime.ID, deprovisioningOperationStatus)
	t.Logf("Runtime deprovisioned successfully")
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
	t.Logf("Assering operation %s is in %s state.", operation.ID, expectedState)
	t.Logf("Operation message: %s", operation.Message)
	require.Equal(t, expectedState, operation.State)
	assert.Equal(t, expectedRuntimeId, operation.RuntimeID)
	assert.Equal(t, expectedType, operation.Operation)
}

// Standard require.NoError print only the top wrapper of error
func requireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if assert.NoError(t, err) {
		return
	}
	fullError := fmt.Sprintf("Received unexpected error: %s", err.Error())
	t.Fatal(fullError, msgAndArgs)
}

// Standard require.NoError print only the top wrapper of error
func assertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	if assert.NoError(t, err) {
		return
	}
	fullError := fmt.Sprintf("Received unexpected error: %s", err.Error())
	t.Error(fullError, msgAndArgs)
}
