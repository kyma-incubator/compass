package model

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestOperationStatusToGQLOperationStatus(t *testing.T) {
	t.Run("Should create proper operation status struct", func(t *testing.T) {
		//given
		operation := Operation{
			OperationID: "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
			Operation:   Provision,
			State:       Succeeded,
			Message:     "Some message",
			RuntimeID:   "6af76034-272a-42be-ac39-30e075f515a3",
		}

		expectedOperationStatus := &gqlschema.OperationStatus{
			Operation: gqlschema.OperationTypeProvision,
			State:     gqlschema.OperationStateSucceeded,
			Message:   "Some message",
			RuntimeID: "6af76034-272a-42be-ac39-30e075f515a3",
		}

		//when
		status := OperationStatusToGQLOperationStatus(operation)

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

		expectedRuntimeConfig := RuntimeConfig{
			ClusterConfig: GCPConfig{
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
			KymaConfig: KymaConfig{
				Version: "1.5",
				Modules: []KymaModule{"Backup", "BackupInit"},
			},
		}

		//when
		runtimeConfig := RuntimeConfigFromInput(input)

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

		runtimeStatus := RuntimeStatus{
			LastOperationStatus: Operation{
				OperationID: "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
				Operation:   Provision,
				State:       Succeeded,
				Message:     "Some message",
				RuntimeID:   "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: RuntimeAgentConnectionStatusConnected,
			RuntimeConfiguration: RuntimeConfig{
				ClusterConfig: GardenerConfig{
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
				KymaConfig: KymaConfig{
					Version: version,
					Modules: []KymaModule{"Backup", "BackupInit"},
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
		gqlStatus := RuntimeStatusToGraphQLStatus(runtimeStatus)

		//then
		assert.Equal(t, expectedRuntimeStatus, gqlStatus)
	})
}
