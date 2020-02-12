package provisioner

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/assertions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	gcpMachineType = "n1-standard-4"
	gcpClusterZone = "europe-west4-b"
)

// TODO: Consider fetching logs from Provisioner on error (or from created Runtime)

func Test_E2E_Gardener(t *testing.T) {
	log := logrus.WithField("TestId", testSuite.TestId)

	log.Infof("Starting Compass Provisioner tests on Gardener")

	for _, provider := range testSuite.gardenerProviders {
		t.Run(provider, func(t *testing.T) {
			t.Parallel()

			log.Info(provider)

			// Provisioning runtime
			// Get Kyma modules from Installation CR
			provisioningInput, err := testkit.CreateGardenerProvisioningInput(&testSuite.config, provider)
			runtimeName := provisioningInput.RuntimeInput.Name

			log.Infof("Provisioning '%s' runtime on %s...", runtimeName, provider)
			provisioningOperationID, runtimeID, err := testSuite.ProvisionerClient.ProvisionRuntime(provisioningInput)
			assertions.RequireNoError(t, err)
			defer ensureClusterIsDeprovisioned(runtimeID)

			log.Infof("Provisioning operation id: %s, runtime id: %s", provisioningOperationID, runtimeID)

			var provisioningOperationStatus gqlschema.OperationStatus

			//Check if another provisioning of the same cluster can start while previous one is in progress
			err = testkit.RunParallelToMainFunction(ProvisioningTimeout+5*time.Second,
				func() error {
					log.Infof("Waiting for provisioning to finish...")
					var waitErr error
					provisioningOperationStatus, waitErr = testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, provisioningOperationID)
					return waitErr
				},
				func() error {
					log.Infof("Checking if operation will fail while other in progress...")
					operationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(provisioningOperationID)
					if err != nil {
						return errors.WithMessagef(err, "Failed to get %s operation status", provisioningOperationID)
					}

					if operationStatus.State != gqlschema.OperationStateInProgress {
						return errors.New("Operation %s not in progress")
					}

					_, _, err = testSuite.ProvisionerClient.ProvisionRuntime(provisioningInput)
					if err == nil {
						return errors.New("Operation scheduled successfully while other operation in progress")
					}

					return nil
				},
			)
			assertions.RequireNoError(t, err, "Provisioning operation status: ", provisioningOperationStatus.State)

			assertions.AssertOperationSucceed(t, gqlschema.OperationTypeProvision, runtimeID, provisioningOperationStatus)
			log.Infof("Runtime provisioned successfully on %s", provider)

			log.Infof("Fetching %s runtime status...", provider)
			runtimeStatus, err := testSuite.ProvisionerClient.RuntimeStatus(runtimeID)
			assertions.RequireNoError(t, err)

			assertGardenerRuntimeConfiguration(t, provisioningInput, runtimeStatus)

			log.Infof("Preparing K8s client...")
			k8sClient := testSuite.KubernetesClientFromRawConfig(t, *runtimeStatus.RuntimeConfiguration.Kubeconfig)

			logrus.Infof("Accessing API Server on provisioned cluster...")
			_, err = k8sClient.ServerVersion()
			assertions.RequireNoError(t, err)

			// TODO: Run E2e Runtime tests

			log.Infof("Deprovisioning %s runtime %s...", provider, runtimeName)
			deprovisioningOperationID, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtimeID)
			assertions.RequireNoError(t, err)
			log.Infof("Deprovisioning operation id: %s", deprovisioningOperationID)

			deprovisioningOperationStatus, err := testSuite.WaitUntilOperationIsFinished(DeprovisioningTimeout, deprovisioningOperationID)
			assertions.RequireNoError(t, err)
			assertions.AssertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtimeID, deprovisioningOperationStatus)
			log.Infof("Runtime deprovisioned successfully")
		})
	}
}

func Test_E2e(t *testing.T) {
	// Support for GCP was dropped and for now the GCP tests are skipped
	t.SkipNow()

	logrus.Infof("Starting provisioner tests concerning GCP. Test id: %s", testSuite.TestId)

	// Provision runtime
	runtimeName := fmt.Sprintf("%s%s", "runtime", uuid.New().String()[:4])

	provisioningInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name: runtimeName,
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{
				Name:              "gke-provisioner-test-" + testSuite.TestId,
				ProjectName:       config.GCP.ProjectName,
				KubernetesVersion: "1.14",
				NumberOfNodes:     3,
				BootDiskSizeGb:    35, // minimal value
				MachineType:       gcpMachineType,
				Region:            gcpClusterZone,
			},
		},
		KymaConfig: &gqlschema.KymaConfigInput{Version: testSuite.config.Kyma.Version, Components: []*gqlschema.ComponentConfigurationInput{
			{Component: "core", Namespace: "kyma-system"}, // TODO: modules need to be adjusted
		}},
	}

	logrus.Infof("Provisioning %s runtime on GCP...", runtimeName)
	provisioningOperationId, runtimeId, err := testSuite.ProvisionerClient.ProvisionRuntime(provisioningInput)
	assertions.RequireNoError(t, err)
	logrus.Infof("Provisioning operation id: %s, runtime id: %s", provisioningOperationId, runtimeId)
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

			_, _, err = testSuite.ProvisionerClient.ProvisionRuntime(provisioningInput)
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
	runtimeStatus, err := testSuite.ProvisionerClient.RuntimeStatus(runtimeId)
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

func assertGCPRuntimeConfiguration(t *testing.T, input gqlschema.ProvisionRuntimeInput, status gqlschema.RuntimeStatus) {
	assertRuntimeConfiguration(t, status)

	clusterConfig, ok := status.RuntimeConfiguration.ClusterConfig.(*gqlschema.GCPConfig)
	if !ok {
		t.Error("Cluster Config does not match GCPConfig type")
		t.FailNow()
	}

	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GcpConfig.Name, clusterConfig.Name)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GcpConfig.Region, clusterConfig.Region)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GcpConfig.KubernetesVersion, clusterConfig.KubernetesVersion)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GcpConfig.BootDiskSizeGb, clusterConfig.BootDiskSizeGb)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GcpConfig.MachineType, clusterConfig.MachineType)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GcpConfig.NumberOfNodes, clusterConfig.NumberOfNodes)
	assert.Equal(t, unwrapString(input.ClusterConfig.GcpConfig.Zone), unwrapString(clusterConfig.Zone))
}

func assertGardenerRuntimeConfiguration(t *testing.T, input gqlschema.ProvisionRuntimeInput, status gqlschema.RuntimeStatus) {
	assertRuntimeConfiguration(t, status)

	clusterConfig, ok := status.RuntimeConfiguration.ClusterConfig.(*gqlschema.GardenerConfig)
	if !ok {
		t.Error("Cluster Config does not match GardenerConfig type")
		t.FailNow()
	}

	assert.NotEmpty(t, clusterConfig.Name)
	assert.NotEmpty(t, clusterConfig.Seed)

	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.Region, clusterConfig.Region)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.KubernetesVersion, clusterConfig.KubernetesVersion)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.VolumeSizeGb, clusterConfig.VolumeSizeGb)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.MachineType, clusterConfig.MachineType)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.NodeCount, clusterConfig.NodeCount)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.DiskType, clusterConfig.DiskType)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.Provider, clusterConfig.Provider)
	assertions.AssertNotNilAndEqualString(t, input.ClusterConfig.GardenerConfig.WorkerCidr, clusterConfig.WorkerCidr)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.MaxUnavailable, clusterConfig.MaxUnavailable)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.AutoScalerMin, clusterConfig.AutoScalerMin)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.AutoScalerMax, clusterConfig.AutoScalerMax)
	assertions.AssertNotNilAndEqualInt(t, input.ClusterConfig.GardenerConfig.MaxSurge, clusterConfig.MaxSurge)

	verifyProviderConfig(t, *input.ClusterConfig.GardenerConfig.ProviderSpecificConfig, status.RuntimeConfiguration.ClusterConfig)
}

func assertRuntimeConfiguration(t *testing.T, status gqlschema.RuntimeStatus) {
	require.NotNil(t, status.RuntimeConfiguration)
	require.NotNil(t, status.RuntimeConfiguration.ClusterConfig)
	require.NotNil(t, status.RuntimeConfiguration.Kubeconfig)
	require.NotNil(t, status.RuntimeConfiguration.KymaConfig)
	require.NotNil(t, status.LastOperationStatus)
	//require.NotNil(t, status.RuntimeConnectionStatus) // TODO - uncomment when implemented
}

func verifyProviderConfig(t *testing.T, input gqlschema.ProviderSpecificInput, config interface{}) {
	if input.AzureConfig != nil {
		azureConfig, ok := config.(gqlschema.AzureProviderConfig)
		if !ok {
			t.Failed()
		}
		assertions.AssertNotNilAndEqualString(t, input.AzureConfig.VnetCidr, azureConfig.VnetCidr)
	}

	if input.AwsConfig != nil {
		awsConfig, ok := config.(gqlschema.AWSProviderConfig)
		if !ok {
			t.Failed()
		}
		assertions.AssertNotNilAndEqualString(t, input.AwsConfig.VpcCidr, awsConfig.VpcCidr)
		assertions.AssertNotNilAndEqualString(t, input.AwsConfig.Zone, awsConfig.Zone)
		assertions.AssertNotNilAndEqualString(t, input.AwsConfig.InternalCidr, awsConfig.InternalCidr)
		assertions.AssertNotNilAndEqualString(t, input.AwsConfig.PublicCidr, awsConfig.PublicCidr)
	}

	if input.GcpConfig != nil {
		gcpConfig, ok := config.(gqlschema.GCPProviderConfigInput)
		if !ok {
			t.Failed()
		}
		assertions.AssertNotNilAndEqualString(t, input.GcpConfig.Zone, &gcpConfig.Zone)
	}
}

func unwrapString(str *string) string {
	if str != nil {
		return *str
	}

	return ""
}
