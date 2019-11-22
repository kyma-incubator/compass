package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

func TestOperationStatusToGQLOperationStatus(t *testing.T) {

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
		status := operationStatusToGQLOperationStatus(operation)

		//then
		assert.Equal(t, expectedOperationStatus, status)
	})
}

func TestRuntimeConfigFromGraphQLRuntimeConfig(t *testing.T) {

	createGQLRuntimeInputGCP := func(zone *string) gqlschema.ProvisionRuntimeInput {
		return gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GcpConfig: &gqlschema.GCPConfigInput{
					Name:              "Something",
					ProjectName:       "Project",
					NumberOfNodes:     3,
					BootDiskSizeGb:    256,
					MachineType:       "n1-standard-1",
					Region:            "region",
					Zone:              zone,
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
	}

	createExpectedRuntimeInputGCP := func(zone string) model.RuntimeConfig {
		return model.RuntimeConfig{
			ClusterConfig: model.GCPConfig{
				ID:                "id",
				Name:              "Something",
				ProjectName:       "Project",
				NumberOfNodes:     3,
				BootDiskSizeGB:    256,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Zone:              zone,
				KubernetesVersion: "version",
				ClusterID:         "runtimeID",
			},
			Kubeconfig: nil,
			KymaConfig: model.KymaConfig{
				ID:      "id",
				Version: "1.5",
				Modules: []model.KymaConfigModule{
					{ID: "id", Module: model.KymaModule("Backup"), KymaConfigID: "id"},
					{ID: "id", Module: model.KymaModule("BackupInit"), KymaConfigID: "id"},
				},
				ClusterID: "runtimeID",
			},
			CredentialsSecretName: "secretName",
		}
	}

	gardenerGCPQGLInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Name:              "Something",
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "GCP",
				Seed:              "gcp-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					GcpConfig: &gqlschema.GCPProviderConfigInput{
						Zone: "zone",
					},
				},
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

	expectedGardenerGCPRuntimeConfig := model.RuntimeConfig{
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "GCP",
			Seed:                   "gcp-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"zone\":\"zone\"}",
		},
		Kubeconfig: nil,
		KymaConfig: model.KymaConfig{
			ID:      "id",
			Version: "1.5",
			Modules: []model.KymaConfigModule{
				{ID: "id", Module: model.KymaModule("Backup"), KymaConfigID: "id"},
				{ID: "id", Module: model.KymaModule("BackupInit"), KymaConfigID: "id"},
			},
			ClusterID: "runtimeID",
		},
		CredentialsSecretName: "secretName",
	}

	gardenerAzureQGLInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Name:              "Something",
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "Azure",
				Seed:              "az-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					AzureConfig: &gqlschema.AzureProviderConfigInput{
						VnetCidr: "cidr",
					},
				},
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

	expectedGardenerAzureRuntimeConfig := model.RuntimeConfig{
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "Azure",
			Seed:                   "az-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"vnetCidr\":\"cidr\"}",
		},
		Kubeconfig: nil,
		KymaConfig: model.KymaConfig{
			ID:      "id",
			Version: "1.5",
			Modules: []model.KymaConfigModule{
				{ID: "id", Module: model.KymaModule("Backup"), KymaConfigID: "id"},
				{ID: "id", Module: model.KymaModule("BackupInit"), KymaConfigID: "id"},
			},
			ClusterID: "runtimeID",
		},
		CredentialsSecretName: "secretName",
	}

	gardenerAWSQGLInput := gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Name:              "Something",
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "AWS",
				Seed:              "aws-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					AwsConfig: &gqlschema.AWSProviderConfigInput{
						Zone:         "zone",
						InternalCidr: "cidr",
						VpcCidr:      "cidr",
						PublicCidr:   "cidr",
					},
				},
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

	expectedGardenerAWSRuntimeConfig := model.RuntimeConfig{
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "Something",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "AWS",
			Seed:                   "aws-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			ProviderSpecificConfig: "{\"zone\":\"zone\",\"vpcCidr\":\"cidr\",\"publicCidr\":\"cidr\",\"internalCidr\":\"cidr\"}",
		},
		Kubeconfig: nil,
		KymaConfig: model.KymaConfig{
			ID:      "id",
			Version: "1.5",
			Modules: []model.KymaConfigModule{
				{ID: "id", Module: model.KymaModule("Backup"), KymaConfigID: "id"},
				{ID: "id", Module: model.KymaModule("BackupInit"), KymaConfigID: "id"},
			},
			ClusterID: "runtimeID",
		},
		CredentialsSecretName: "secretName",
	}

	zone := "zone"

	configurations := []struct {
		input       gqlschema.ProvisionRuntimeInput
		expected    model.RuntimeConfig
		description string
	}{
		{
			input:       createGQLRuntimeInputGCP(&zone),
			expected:    createExpectedRuntimeInputGCP(zone),
			description: "Should create proper runtime config struct with GCP input",
		},
		{
			input:       createGQLRuntimeInputGCP(nil),
			expected:    createExpectedRuntimeInputGCP(""),
			description: "Should create proper runtime config struct with GCP input (empty zone)",
		},
		{
			input:       gardenerGCPQGLInput,
			expected:    expectedGardenerGCPRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for GCP provider",
		},
		{
			input:       gardenerAzureQGLInput,
			expected:    expectedGardenerAzureRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for Azure provider",
		},
		{
			input:       gardenerAWSQGLInput,
			expected:    expectedGardenerAWSRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for AWS provider",
		},
	}

	for _, config := range configurations {
		t.Run("Should create proper runtime config struct with GCP input", func(t *testing.T) {
			//given
			uuidGeneratorMock := &mocks.UUIDGenerator{}
			uuidGeneratorMock.On("New").Return("id").Times(4)

			//when
			runtimeConfig, err := runtimeConfigFromInput("runtimeID", config.input, uuidGeneratorMock)

			//then
			require.NoError(t, err)
			assert.Equal(t, config.expected, runtimeConfig)
			uuidGeneratorMock.AssertExpectations(t)
		})
	}
}

func TestRuntimeStatusToGraphQLStatus(t *testing.T) {
	t.Run("Should create proper runtime status struct for GCP config", func(t *testing.T) {
		name := "Something"
		project := "Project"
		numberOfNodes := 3
		bootDiskSize := 256
		machine := "machine"
		region := "region"
		zone := "zone"
		kubeversion := "kubeversion"
		version := "1.5"
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
			RuntimeConfiguration: model.RuntimeConfig{
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
					ID:      "id",
					Version: "1.5",
					Modules: []model.KymaConfigModule{
						{ID: "id", Module: model.KymaModule("Backup"), KymaConfigID: "id"},
						{ID: "id", Module: model.KymaModule("BackupInit"), KymaConfigID: "id"},
					},
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
		gqlStatus := runtimeStatusToGraphQLStatus(runtimeStatus)

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
		version := "1.5"
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

		runtimeStatus := model.RuntimeStatus{
			LastOperationStatus: model.Operation{
				ID:        "5f6e3ab6-d803-430a-8fac-29c9c9b4485a",
				Type:      model.Deprovision,
				State:     model.Failed,
				Message:   "Some message",
				ClusterID: "6af76034-272a-42be-ac39-30e075f515a3",
			},
			RuntimeConnectionStatus: model.RuntimeAgentConnectionStatusDisconnected,
			RuntimeConfiguration: model.RuntimeConfig{
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
					ProviderSpecificConfig: "{\"Zone\":\"zone\"}",
				},
				Kubeconfig: &kubeconfig,
				KymaConfig: model.KymaConfig{
					Version: version,
					Modules: []model.KymaConfigModule{
						{ID: "Id1", Module: model.KymaModule("Backup")},
						{ID: "Id1", Module: model.KymaModule("BackupInit")},
					},
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
		gqlStatus := runtimeStatusToGraphQLStatus(runtimeStatus)

		//then
		assert.Equal(t, expectedRuntimeStatus, gqlStatus)
	})
}
