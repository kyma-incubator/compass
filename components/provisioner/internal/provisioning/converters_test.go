package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestOperationStatusToGQLOperationStatus(t *testing.T) {
	t.Run("Should create proper operation status struct", func(t *testing.T) {
		//given
		operation := model.Operation{
			ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
			Type:      model.Provision,
			State:     model.Succeeded,
			Message:   "Some message",
			ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
		}

		expectedOperationStatus := &gqlschema.OperationStatus{
			Operation: gqlschema.OperationTypeProvision,
			State:     gqlschema.OperationStateSucceeded,
			Message:   "Some message",
			RuntimeID: "6af76034-272a-42be-ac39-30e075f515a3",
		}

		//when
		status := operationStatusToGQLOperationStatus(operation)

		//then
		assert.Equal(t, expectedOperationStatus, status)
	})
}

func TestRuntimeConfigFromInput(t *testing.T) {
	t.Run("Should create proper runtime config struct with gcp input", func(t *testing.T) {
		//given
		zone := "zone"
		input := &gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GcpConfig: &gqlschema.GCPConfigInput{
					Name:              "Something",
					ProjectName:       "Project",
					NumberOfNodes:     3,
					BootDiskSize:      "256",
					MachineType:       "machine",
					Region:            "region",
					Zone:              &zone,
					KubernetesVersion: "version",
				},
			},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: "secretName",
			},
			KymaConfig: &gqlschema.KymaConfigInput{
				Version: "1.5",
				Modules: []gqlschema.KymaModule{gqlschema.KymaModuleBackup, gqlschema.KymaModuleBackupInit},
			},
		}

		expectedRuntimeConfig := model.RuntimeConfig{
			ClusterConfig: model.GCPConfig{
				Name:              "Something",
				ProjectName:       "Project",
				NumberOfNodes:     3,
				BootDiskSize:      "256",
				MachineType:       "machine",
				Region:            "region",
				Zone:              "zone",
				KubernetesVersion: "version",
			},
			Kubeconfig: "",
			KymaConfig: model.KymaConfig{
				Version: "1.5",
				Modules: []model.KymaModule{"Backup", "BackupInit"},
			},
		}

		//when
		runtimeConfig := runtimeConfigFromInput(input)

		//then
		assert.Equal(t, expectedRuntimeConfig, runtimeConfig)
	})
}

func TestRuntimeStatusToGraphQLStatus(t *testing.T) {
	t.Run("Should create proper runtime status struct for gardener config", func(t *testing.T) {
		//given
		name := "Something"
		project := "Project"
		nodes := 3
		disk := "256"
		machine := "machine"
		region := "region"
		zone := "zone"
		volume := "volume"
		kubeversion := "kubeversion"
		version := "1.5"
		backup := gqlschema.KymaModuleBackup
		backupInit := gqlschema.KymaModuleBackupInit
		kubeconfig := "kubeconfig"
		provider := "GCP"
		secret := "secret"
		cidr := "cidr"
		autoScMax := 2
		autoScMin := 2
		surge := 1
		unavailable := 1

		runtimeStatus := model.RuntimeStatus{
			LastOperationStatus: model.Operation{
				ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
				Type:      model.Provision,
				State:     model.Succeeded,
				Message:   "Some message",
				ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: model.RuntimeAgentConnectionStatusConnected,
			RuntimeConfiguration: model.RuntimeConfig{
				ClusterConfig: model.GardenerConfig{
					Name:              name,
					ProjectName:       project,
					NodeCount:         nodes,
					DiskType:          disk,
					MachineType:       machine,
					Region:            region,
					Zone:              zone,
					VolumeSize:        volume,
					KubernetesVersion: kubeversion,
					TargetProvider:    provider,
					TargetSecret:      secret,
					Cidr:              cidr,
					AutoScalerMax:     autoScMax,
					AutoScalerMin:     autoScMin,
					MaxSurge:          surge,
					MaxUnavailable:    unavailable,
				},
				Kubeconfig: kubeconfig,
				KymaConfig: model.KymaConfig{
					Version: version,
					Modules: []model.KymaModule{"Backup", "BackupInit"},
				},
			},
		}

		expectedRuntimeStatus := &gqlschema.RuntimeStatus{
			LastOperationStatus: &gqlschema.OperationStatus{
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateSucceeded,
				Message:   "Some message",
				RuntimeID: "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{
				Status: gqlschema.RuntimeAgentConnectionStatusConnected,
			},
			RuntimeConfiguration: &gqlschema.RuntimeConfig{
				ClusterConfig: gqlschema.GardenerConfig{
					Name:              &name,
					ProjectName:       &project,
					NodeCount:         &nodes,
					DiskType:          &disk,
					MachineType:       &machine,
					Region:            &region,
					Zone:              &zone,
					VolumeSize:        &volume,
					KubernetesVersion: &kubeversion,
					TargetProvider:    &provider,
					TargetSecret:      &secret,
					Cidr:              &cidr,
					AutoScalerMax:     &autoScMax,
					AutoScalerMin:     &autoScMin,
					MaxSurge:          &surge,
					MaxUnavailable:    &unavailable,
				},
				KymaConfig: &gqlschema.KymaConfig{
					Version: &version,
					Modules: []*gqlschema.KymaModule{&backup, &backupInit},
				},
				Kubeconfig: &kubeconfig,
			},
		}

		//when
		gqlStatus := runtimeStatusToGraphQLStatus(runtimeStatus)

		//then
		assert.Equal(t, expectedRuntimeStatus, gqlStatus)
	})
}
