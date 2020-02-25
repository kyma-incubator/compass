package provisioner

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit/assertions"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO: Consider fetching logs from Provisioner on error (or from created Runtime)

func Test_E2E_Gardener(t *testing.T) {
	globalLog := logrus.WithField("TestId", testSuite.TestId)

	globalLog.Infof("Starting Compass Provisioner tests on Gardener")
	wg := &sync.WaitGroup{}

	for _, provider := range testSuite.gardenerProviders {
		wg.Add(1)
		go func(provider string) {
			defer wg.Done()
			defer testSuite.Recover()

			t.Run(provider, func(t *testing.T) {
				log := NewLogger(t, fmt.Sprintf("Provider=%s", provider))

				// Provisioning runtime
				// Create provisioning input
				provisioningInput, err := testkit.CreateGardenerProvisioningInput(&testSuite.config, provider)
				assertions.RequireNoError(t, err)

				runtimeName := fmt.Sprintf("provisioner-test-%s-%s", provider, uuid.New().String()[:4])
				provisioningInput.RuntimeInput.Name = runtimeName
				//runtime := NewRuntime(provisioningInput)

				// Provision runtime
				log.Log("Starting provisioning...")
				provisioningOperationID, runtimeID, err := testSuite.ProvisionerClient.ProvisionRuntime(provisioningInput)
				assertions.RequireNoError(t, err, "Error while starting Runtime provisioning")
				defer ensureClusterIsDeprovisioned(runtimeID)

				log.AddField(fmt.Sprintf("RuntimeId=%s", runtimeID))
				log.AddField(fmt.Sprintf("ProvisioningOperationId=%s", provisioningOperationID))

				// Get provisioning Operation Status
				log.Log("Getting operation status...")
				provisioningOperationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(provisioningOperationID)
				assertions.RequireNoError(t, err, "Error while getting operation id")
				assertions.AssertOperationInProgress(t, gqlschema.OperationTypeProvision, runtimeID, provisioningOperationStatus)

				// Wait for provisioning to finish
				log.Log("Waiting for provisioning to finish...")
				provisioningOperationStatus, err = testSuite.WaitUntilOperationIsFinished(ProvisioningTimeout, provisioningOperationID)
				assertions.RequireNoError(t, err)
				assertions.AssertOperationSucceed(t, gqlschema.OperationTypeProvision, runtimeID, provisioningOperationStatus)
				log.Log("Provisioning finished.")

				// Fetch Runtime Status
				log.Log("Getting Runtime status...")
				runtimeStatus, err := testSuite.ProvisionerClient.RuntimeStatus(runtimeID)
				assertions.RequireNoError(t, err)

				// Asserting Gardener Configuration
				log.Log("Verifying configuration...")
				assertGardenerRuntimeConfiguration(t, provisioningInput, runtimeStatus)

				log.Log("Preparing K8s client...")
				k8sClient := testSuite.KubernetesClientFromRawConfig(t, *runtimeStatus.RuntimeConfiguration.Kubeconfig)

				log.Log("Accessing API Server on provisioned cluster...")
				_, err = k8sClient.ServerVersion()
				assertions.RequireNoError(t, err)

				// TODO: Consider running E2E Runtime tests

				// Deprovisioning runtime
				log.Log("Starting Runtime deprovisioning...")
				deprovisioningOperationID, err := testSuite.ProvisionerClient.DeprovisionRuntime(runtimeID)
				assertions.RequireNoError(t, err)

				log.AddField(fmt.Sprintf("DeprovisioningOperationId=%s", deprovisioningOperationID))

				// Get provisioning Operation Status
				deprovisioningOperationStatus, err := testSuite.ProvisionerClient.RuntimeOperationStatus(deprovisioningOperationID)
				assertions.RequireNoError(t, err)
				assertions.AssertOperationInProgress(t, gqlschema.OperationTypeDeprovision, runtimeID, deprovisioningOperationStatus)

				log.Log("Waitinh for deprovisioning to finish...")
				deprovisioningOperationStatus, err = testSuite.WaitUntilOperationIsFinished(DeprovisioningTimeout, deprovisioningOperationID)
				assertions.RequireNoError(t, err)
				assertions.AssertOperationSucceed(t, gqlschema.OperationTypeDeprovision, runtimeID, deprovisioningOperationStatus)
				log.Log("Deprovisioning finished.")
			})
		}(provider)
	}
	wg.Wait()
}

type Logger struct {
	t            *testing.T
	fields       []string
	joinedFields string
}

func NewLogger(t *testing.T, fields ...string) *Logger {
	return &Logger{
		t:            t,
		fields:       fields,
		joinedFields: strings.Join(fields, " "),
	}
}

func (l Logger) Log(msg string) {
	l.t.Logf("%s   %s", msg, l.joinedFields)
}

func (l *Logger) AddField(field string) {
	l.fields = append(l.fields, field)
	l.joinedFields = strings.Join(l.fields, " ")
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

func assertGardenerRuntimeConfiguration(t *testing.T, input gqlschema.ProvisionRuntimeInput, status gqlschema.RuntimeStatus) {
	assert.NotEmpty(t, status)
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
