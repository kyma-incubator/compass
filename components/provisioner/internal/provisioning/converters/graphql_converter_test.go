package converters

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestOperationStatusToGQLOperationStatus(t *testing.T) {

	graphQLConverter := NewGraphQLConverter()

	t.Run("Should create proper operation status struct", func(t *testing.T) {
		//given
		operation := model.Operation{
			ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
			Type:      model.Upgrade,
			State:     model.InProgress,
			Message:   "Some message",
			ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
		}

		operationID := "5f6e3ab6-d803-430a-8fac-29c9c9b4485a"
		message := "Some message"
		runtimeID := "6af76034-272a-42be-ac39-30e075f515a3"

		expectedOperationStatus := &gqlschema.OperationStatus{
			ID:        &operationID,
			Operation: gqlschema.OperationTypeUpgrade,
			State:     gqlschema.OperationStateInProgress,
			Message:   &message,
			RuntimeID: &runtimeID,
		}

		//when
		status := graphQLConverter.OperationStatusToGQLOperationStatus(operation)

		//then
		assert.Equal(t, expectedOperationStatus, status)
	})
}

func TestRuntimeStatusToGraphQLStatus(t *testing.T) {

	graphQLConverter := NewGraphQLConverter()

	t.Run("Should create proper runtime status struct for GCP config", func(t *testing.T) {
		name := "Something"
		project := "Project"
		numberOfNodes := 3
		bootDiskSize := 256
		machine := "machine"
		region := "region"
		zone := "zone"
		kubeversion := "kubeversion"
		version := kymaVersion
		backup := gqlschema.KymaModuleBackup
		backupInit := gqlschema.KymaModuleBackupInit
		kubeconfig := "kubeconfig"
		secretName := "secretName"

		runtimeStatus := model.RuntimeStatus{
			LastOperationStatus: model.Operation{
				ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
				Type:      model.Provision,
				State:     model.Succeeded,
				Message:   "Some message",
				ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: model.RuntimeAgentConnectionStatusConnected,
			RuntimeConfiguration: model.Cluster{
				ClusterConfig: model.GCPConfig{
					ID:                "id",
					Name:              "Something",
					ProjectName:       "Project",
					NumberOfNodes:     3,
					BootDiskSizeGB:    256,
					MachineType:       "machine",
					Region:            "region",
					Zone:              "zone",
					KubernetesVersion: "kubeversion",
					ClusterID:         "runtimeID",
				},
				Kubeconfig: &kubeconfig,
				KymaConfig: model.KymaConfig{
					ID:        "id",
					Release:   fixKymaRelease(),
					Modules:   fixKymaModules(),
					ClusterID: "runtimeID",
				},
				CredentialsSecretName: secretName,
			},
		}

		operationID := "5f6e3ab6-d803-430a-8fac-29c9c9b4485a"
		message := "Some message"
		runtimeID := "6af76034-272a-42be-ac39-30e075f515a3"

		expectedRuntimeStatus := &gqlschema.RuntimeStatus{
			LastOperationStatus: &gqlschema.OperationStatus{
				ID:        &operationID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateSucceeded,
				Message:   &message,
				RuntimeID: &runtimeID,
			},
			RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{
				Status: gqlschema.RuntimeAgentConnectionStatusConnected,
			},
			RuntimeConfiguration: &gqlschema.RuntimeConfig{
				ClusterConfig: gqlschema.GCPConfig{
					Name:              &name,
					ProjectName:       &project,
					MachineType:       &machine,
					Region:            &region,
					Zone:              &zone,
					NumberOfNodes:     &numberOfNodes,
					BootDiskSizeGb:    &bootDiskSize,
					KubernetesVersion: &kubeversion,
				},
				KymaConfig: &gqlschema.KymaConfig{
					Version: &version,
					Modules: []*gqlschema.KymaModule{&backup, &backupInit},
				},
				Kubeconfig:            &kubeconfig,
				CredentialsSecretName: &secretName,
			},
		}

		//when
		gqlStatus := graphQLConverter.RuntimeStatusToGraphQLStatus(runtimeStatus)

		//then
		assert.Equal(t, expectedRuntimeStatus, gqlStatus)
	})

	t.Run("Should create proper runtime status struct for gardener config", func(t *testing.T) {
		//given
		name := "Something"
		project := "Project"
		nodes := 3
		disk := "standard"
		machine := "machine"
		region := "region"
		zone := "zone"
		volume := 256
		kubeversion := "kubeversion"
		version := kymaVersion
		backup := gqlschema.KymaModuleBackup
		backupInit := gqlschema.KymaModuleBackupInit
		kubeconfig := "kubeconfig"
		provider := "GCP"
		seed := "gcp-eu1"
		secret := "secret"
		cidr := "cidr"
		autoScMax := 2
		autoScMin := 2
		surge := 1
		unavailable := 1
		secretName := "secretName"

		gardenerProviderConfig, err := model.NewGardenerProviderConfigFromJSON(`{"Zone":"zone"}`)
		require.NoError(t, err)

		runtimeStatus := model.RuntimeStatus{
			LastOperationStatus: model.Operation{
				ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
				Type:      model.Deprovision,
				State:     model.Failed,
				Message:   "Some message",
				ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: model.RuntimeAgentConnectionStatusDisconnected,
			RuntimeConfiguration: model.Cluster{
				ClusterConfig: model.GardenerConfig{
					Name:                   name,
					ProjectName:            project,
					NodeCount:              nodes,
					DiskType:               disk,
					MachineType:            machine,
					Region:                 region,
					VolumeSizeGB:           volume,
					KubernetesVersion:      kubeversion,
					Provider:               provider,
					Seed:                   seed,
					TargetSecret:           secret,
					WorkerCidr:             cidr,
					AutoScalerMax:          autoScMax,
					AutoScalerMin:          autoScMin,
					MaxSurge:               surge,
					MaxUnavailable:         unavailable,
					GardenerProviderConfig: gardenerProviderConfig,
				},
				Kubeconfig: &kubeconfig,
				KymaConfig: model.KymaConfig{
					Release: fixKymaRelease(),
					Modules: fixKymaModules(),
				},
				CredentialsSecretName: secretName,
			},
		}

		operationID := "5f6e3ab6-d803-430a-8fac-29c9c9b4485a"
		message := "Some message"
		runtimeID := "6af76034-272a-42be-ac39-30e075f515a3"

		expectedRuntimeStatus := &gqlschema.RuntimeStatus{
			LastOperationStatus: &gqlschema.OperationStatus{
				ID:        &operationID,
				Operation: gqlschema.OperationTypeDeprovision,
				State:     gqlschema.OperationStateFailed,
				Message:   &message,
				RuntimeID: &runtimeID,
			},
			RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{
				Status: gqlschema.RuntimeAgentConnectionStatusDisconnected,
			},
			RuntimeConfiguration: &gqlschema.RuntimeConfig{
				ClusterConfig: gqlschema.GardenerConfig{
					Name:              &name,
					ProjectName:       &project,
					NodeCount:         &nodes,
					DiskType:          &disk,
					MachineType:       &machine,
					Region:            &region,
					VolumeSizeGb:      &volume,
					KubernetesVersion: &kubeversion,
					Provider:          &provider,
					Seed:              &seed,
					TargetSecret:      &secret,
					WorkerCidr:        &cidr,
					AutoScalerMax:     &autoScMax,
					AutoScalerMin:     &autoScMin,
					MaxSurge:          &surge,
					MaxUnavailable:    &unavailable,
					ProviderSpecificConfig: gqlschema.GCPProviderConfig{
						Zone: &zone,
					},
				},
				KymaConfig: &gqlschema.KymaConfig{
					Version: &version,
					Modules: []*gqlschema.KymaModule{&backup, &backupInit},
				},
				Kubeconfig:            &kubeconfig,
				CredentialsSecretName: &secretName,
			},
		}

		//when
		gqlStatus := graphQLConverter.RuntimeStatusToGraphQLStatus(runtimeStatus)

		//then
		assert.Equal(t, expectedRuntimeStatus, gqlStatus)
	})
}

func fixKymaModules() []model.KymaConfigModule {
	return []model.KymaConfigModule{
		{ID: "id", KymaConfigID: "id", Module: model.KymaModule("Backup")},
		{ID: "id", KymaConfigID: "id", Module: model.KymaModule("BackupInit")},
	}
}

func fixKymaRelease() model.Release {
	return model.Release{
		Id:            "d829b1b5-2e82-426d-91b0-f94978c0c140",
		Version:       kymaVersion,
		TillerYAML:    "tiller yaml",
		InstallerYAML: "installer yaml",
	}
}
