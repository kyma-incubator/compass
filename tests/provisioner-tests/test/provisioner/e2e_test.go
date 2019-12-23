package provisioner

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/assertions"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/compass/provisioner"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/require"
)

const (
	gcpMachineType = "n1-standard-4"
	gcpClusterZone = "europe-west4-b"
)

var GardenerInputs = map[string]gqlschema.GardenerConfigInput{
	"Azure": gqlschema.GardenerConfigInput{
		MachineType:  "Standard_D2_v3",
		DiskType:     "Standard_LRS",
		Region:       "westeurope",
		Seed:         "az-eu1",
		TargetSecret: "mszymanski-azure", // TODO

		ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
			AzureConfig: &gqlschema.AzureProviderConfigInput{
				VnetCidr: "10.250.0.0/19",
			},
		},
	},
}

func Test_E2E_Gardener(t *testing.T) {

	logrus.Infof("Starting Compass Provisioner tests concerning Gardener. Test ID: %s", testSuite.TestId)

	for _, provider := range testSuite.providers {

		runtimeId := uuid.New().String()

		// Provision runtime
		credentialsInput := gqlschema.CredentialsInput{SecretName: testSuite.CredentialsSecretName}

		provisioningInput := gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GardenerConfig: &gqlschema.GardenerConfigInput{
					Name:                   provider + "-test",
					ProjectName:            config.GardenerProjectName,
					KubernetesVersion:      "1.15.4",
					NodeCount:              3,
					DiskType:               GardenerInputs[provider].DiskType,
					VolumeSizeGb:           35,
					MachineType:            GardenerInputs[provider].MachineType,
					Region:                 GardenerInputs[provider].Region,
					Provider:               provider,
					Seed:                   GardenerInputs[provider].Seed,
					TargetSecret:           GardenerInputs[provider].TargetSecret,
					WorkerCidr:             "10.250.0.0/19",
					AutoScalerMin:          2,
					AutoScalerMax:          4,
					MaxSurge:               4,
					MaxUnavailable:         1,
					ProviderSpecificConfig: GardenerInputs[provider].ProviderSpecificConfig,
				},
			},
			Credentials: &credentialsInput,
			KymaConfig:  &gqlschema.KymaConfigInput{Version: "1.8", Modules: gqlschema.AllKymaModule},
		}

		logrus.Infof("Provisioning %s runtime on %s...", runtimeId, toName(provider))
		provisioningOperationId, err := testSuite.ProvisionerClient.ProvisionRuntime(runtimeId, provisioningInput)
		assertions.RequireNoError(t, err)
		logrus.Infof("Provisioning operation id: %s", provisioningOperationId)
		defer ensureClusterIsDeprovisioned(runtimeId)

		var provisioningOperationStatus gqlschema.OperationStatus
		err = testkit.RunParallelToMainFunction(ProvisioningTimeout+5*time.Second,
			func() error {
				logrus.Infof("Waiting for provisioning to finish...")
				var waitErr error
				provisioningOperationStatus, waitErr = testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, provisioningOperationId)
				return waitErr
			},
			func() error {
				logrus.Infof("Checking if operation will fail while other in progress...")
				operationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(provisioningOperationId)
				if err != nil {
					return errors.WithMessagef(err, "Failed to get %s operation status", provisioningOperationId)
				}

				if operationStatus.State != gqlschema.OperationStateInProgress {
					return errors.New("Operation %s not in progress")
				}

				_, err = testSuite.ProvisionerClient.ProvisionRuntime(runtimeId, provisioningInput)
				if err == nil {
					return errors.New("Operation scheduled successfully while other operation in progress")
				}

				return nil
			},
		)
		assertions.RequireNoError(t, err, "Provisioning operation status: ", provisioningOperationStatus.State)

		assertions.AssertOperationSucceed(t, gqlschema.OperationTypeProvision, runtimeId, provisioningOperationStatus)
		logrus.Infof("Runtime provisioned successfully on %s", provider)

		logrus.Infof("Fetching %s runtime status...", provider)
		runtimeStatus, err := testSuite.ProvisionerClient.GCPRuntimeStatus(runtimeId)
		assertions.RequireNoError(t, err)

		assertGCPRuntimeConfiguration(t, provisioningInput, runtimeStatus)

		logrus.Infof("Deprovisioning %s runtime on...", provider)
		deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtimeId)
		assertions.RequireNoError(t, err)
		logrus.Infof("Deprovisioning operation id: %s", deprovisioningOperationId)

		deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(DeprovisioningTimeout, deprovisioningOperationId)
		assertions.RequireNoError(t, err)
		assertions.AssertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtimeId, deprovisioningOperationStatus)
		logrus.Infof("Runtime deprovisioned successfully")
	}
}

func Test_E2e(t *testing.T) {
	// TODO: Support for GCP was dropped and for now the GCP tests are skipped
	t.SkipNow()

	logrus.Infof("Starting provisioner tests concerning GCP. Test id: %s", testSuite.TestId)

	runtimeId := uuid.New().String()

	// Provision runtime
	credentialsInput := gqlschema.CredentialsInput{SecretName: testSuite.CredentialsSecretName}

	provisioningInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{
				Name:              "gke-provisioner-test-" + testSuite.TestId,
				ProjectName:       config.GCPProjectName,
				KubernetesVersion: "1.14",
				NumberOfNodes:     3,
				BootDiskSizeGb:    35, // minimal value
				MachineType:       gcpMachineType,
				Region:            gcpClusterZone,
			},
		},
		Credentials: &credentialsInput,
		KymaConfig:  &gqlschema.KymaConfigInput{Version: "1.6", Modules: gqlschema.AllKymaModule},
	}

	logrus.Infof("Provisioning %s runtime on GCP...", runtimeId)
	provisioningOperationId, err := testSuite.ProvisionerClient.ProvisionRuntime(runtimeId, provisioningInput)
	assertions.RequireNoError(t, err)
	logrus.Infof("Provisioning operation id: %s", provisioningOperationId)
	defer ensureClusterIsDeprovisioned(runtimeId)

	var provisioningOperationStatus gqlschema.OperationStatus
	err = testkit.RunParallelToMainFunction(ProvisioningTimeout+5*time.Second,
		func() error {
			logrus.Infof("Waiting for provisioning to finish...")
			var waitErr error
			provisioningOperationStatus, waitErr = testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, provisioningOperationId)
			return waitErr
		},
		func() error {
			logrus.Infof("Checking if operation will fail while other in progress...")
			operationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(provisioningOperationId)
			if err != nil {
				return errors.WithMessagef(err, "Failed to get %s operation status", provisioningOperationId)
			}

			if operationStatus.State != gqlschema.OperationStateInProgress {
				return errors.New("Operation %s not in progress")
			}

			_, err = testSuite.ProvisionerClient.ProvisionRuntime(runtimeId, provisioningInput)
			if err == nil {
				return errors.New("Operation scheduled successfully while other operation in progress")
			}

			return nil
		},
	)
	assertions.RequireNoError(t, err, "Provisioning operation status: ", provisioningOperationStatus.State)

	assertions.AssertOperationSucceed(t, gqlschema.OperationTypeProvision, runtimeId, provisioningOperationStatus)
	logrus.Infof("Runtime provisioned successfully")

	logrus.Infof("Fetching runtime status...")
	runtimeStatus, err := testSuite.ProvisionerClient.GCPRuntimeStatus(runtimeId)
	assertions.RequireNoError(t, err)

	assertGCPRuntimeConfiguration(t, provisioningInput, runtimeStatus)

	// TODO - Perform check when the Hydroform issue is resolved (https://github.com/kyma-incubator/hydroform/issues/26)
	//logrus.Infof("Preparing K8s client...")
	//k8sClient := testSuite.KubernetesClientFromRawConfig(t, *runtimeStatus.RuntimeConfiguration.Kubeconfig)
	//
	//logrus.Infof("Accessing API Server on provisioned cluster...")
	//_, err = k8sClient.ServerVersion()
	//requireNoError(t, err)

	// TODO - Run Compass Runtime Agent Tests - it may require passing Credentials for MP

	logrus.Infof("Deprovisioning runtime...")
	deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtimeId)
	assertions.RequireNoError(t, err)
	logrus.Infof("Deprovisioning operation id: %s", deprovisioningOperationId)

	deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(DeprovisioningTimeout, deprovisioningOperationId)
	assertions.RequireNoError(t, err)
	assertions.AssertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtimeId, deprovisioningOperationStatus)
	logrus.Infof("Runtime deprovisioned successfully")
}

func ensureClusterIsDeprovisioned(runtimeId string) {
	logrus.Infof("Ensuring the cluster is deprovisioned...")
	deprovisioningOperationId, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtimeId)
	if err != nil {
		logrus.Warnf("Ensuring the cluster is deprovisioned failed, cluster might have already been deprovisioned: %s", err.Error())
		return
	}

	logrus.Infof("Deprovisioning operation id: %s", deprovisioningOperationId)
	deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(DeprovisioningTimeout, deprovisioningOperationId)
	if err != nil {
		logrus.Errorf("Error while waiting for deprovisioning operation to finish: %s", err.Error())
		return
	}

	if deprovisioningOperationStatus.State != gqlschema.OperationStateSucceeded {
		logrus.Errorf("Ensuring the cluster is deprovisioned failed with operation status %s with message %s", deprovisioningOperationStatus.State, unwrapString(deprovisioningOperationStatus.Message))
	}
}

func assertGCPRuntimeConfiguration(t *testing.T, input gqlschema.ProvisionRuntimeInput, status provisioner.GCPRuntimeStatus) {
	require.NotNil(t, status.RuntimeConfiguration)
	require.NotNil(t, status.RuntimeConfiguration.ClusterConfig)
	require.NotNil(t, status.RuntimeConfiguration.Kubeconfig)
	require.NotNil(t, status.RuntimeConfiguration.KymaConfig)
	require.NotNil(t, status.LastOperationStatus)
	//require.NotNil(t, status.RuntimeConnectionStatus) // TODO - uncomment when implemented

	gcpClusterConfig := status.RuntimeConfiguration.ClusterConfig

	assertions.AssertNotNillAndEqualString(t, input.ClusterConfig.GcpConfig.Name, gcpClusterConfig.Name)
	assertions.AssertNotNillAndEqualString(t, input.ClusterConfig.GcpConfig.Region, gcpClusterConfig.Region)
	assertions.AssertNotNillAndEqualString(t, input.ClusterConfig.GcpConfig.KubernetesVersion, gcpClusterConfig.KubernetesVersion)
	assertions.AssertNotNillAndEqualInt(t, input.ClusterConfig.GcpConfig.BootDiskSizeGb, gcpClusterConfig.BootDiskSizeGb)
	assertions.AssertNotNillAndEqualString(t, input.ClusterConfig.GcpConfig.MachineType, gcpClusterConfig.MachineType)
	assertions.AssertNotNillAndEqualInt(t, input.ClusterConfig.GcpConfig.NumberOfNodes, gcpClusterConfig.NumberOfNodes)
	assert.Equal(t, unwrapString(input.ClusterConfig.GcpConfig.Zone), unwrapString(gcpClusterConfig.Zone))
}

func unwrapString(str *string) string {
	if str != nil {
		return *str
	}

	return ""
}

func toName(provider string) string {
	return strings.ToUpper(string(provider[0])) + provider[1:]
}
