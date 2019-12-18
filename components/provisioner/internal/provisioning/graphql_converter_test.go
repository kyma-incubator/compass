package provisioning

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

const (
	kymaSystemNamespace      = "kyma-system"
	kymaIntegrationNamespace = "kyma-integration"
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
				Kubeconfig:            &kubeconfig,
				KymaConfig:            fixKymaConfig(),
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
				KymaConfig:            fixKymaGraphQLConfig(),
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
				Kubeconfig:            &kubeconfig,
				KymaConfig:            fixKymaConfig(),
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
				KymaConfig:            fixKymaGraphQLConfig(),
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

func fixKymaGraphQLConfig() *gqlschema.KymaConfig {
	return &gqlschema.KymaConfig{
		Version: util.StringPtr(kymaVersion),
		Components: []*gqlschema.ComponentConfiguration{
			{
				Component:     clusterEssentialsComponent,
				Namespace:     kymaSystemNamespace,
				Configuration: make([]*gqlschema.ConfigEntry, 0, 0),
			},
			{
				Component: coreComponent,
				Namespace: kymaSystemNamespace,
				Configuration: []*gqlschema.ConfigEntry{
					fixGQLConfigEntry("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntry("test.config.key2", "value2", util.BoolPtr(false)),
				},
			},
			{
				Component: applicationConnectorComponent,
				Namespace: kymaIntegrationNamespace,
				Configuration: []*gqlschema.ConfigEntry{
					fixGQLConfigEntry("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntry("test.secret.key", "secretValue", util.BoolPtr(true)),
				},
			},
		},
		Configuration: []*gqlschema.ConfigEntry{
			fixGQLConfigEntry("global.config.key", "globalValue", util.BoolPtr(false)),
			fixGQLConfigEntry("global.config.key2", "globalValue2", util.BoolPtr(false)),
			fixGQLConfigEntry("global.secret.key", "globalSecretValue", util.BoolPtr(true)),
		},
	}
}

func fixGQLConfigEntry(key, val string, secret *bool) *gqlschema.ConfigEntry {
	return &gqlschema.ConfigEntry{
		Key:    key,
		Value:  val,
		Secret: secret,
	}
}

func fixKymaConfig() model.KymaConfig {
	return model.KymaConfig{
		ID:                  "id",
		Release:             fixKymaRelease(),
		Components:          fixKymaComponents(),
		GlobalConfiguration: fixGlobalConfig(),
		ClusterID:           "runtimeID",
	}
}

func fixGlobalConfig() model.Configuration {
	return model.Configuration{
		ConfigEntries: []model.ConfigEntry{
			model.NewConfigEntry("global.config.key", "globalValue", false),
			model.NewConfigEntry("global.config.key2", "globalValue2", false),
			model.NewConfigEntry("global.secret.key", "globalSecretValue", true),
		},
	}
}

func fixKymaComponents() []model.KymaComponentConfig {
	return []model.KymaComponentConfig{
		{
			ID:            "id",
			KymaConfigID:  "id",
			Component:     clusterEssentialsComponent,
			Namespace:     kymaSystemNamespace,
			Configuration: model.Configuration{ConfigEntries: make([]model.ConfigEntry, 0, 0)},
		},
		{
			ID:           "id",
			KymaConfigID: "id",
			Component:    coreComponent,
			Namespace:    kymaSystemNamespace,
			Configuration: model.Configuration{
				ConfigEntries: []model.ConfigEntry{
					model.NewConfigEntry("test.config.key", "value", false),
					model.NewConfigEntry("test.config.key2", "value2", false),
				},
			},
		},
		{
			ID:           "id",
			KymaConfigID: "id",
			Component:    applicationConnectorComponent,
			Namespace:    kymaIntegrationNamespace,
			Configuration: model.Configuration{
				ConfigEntries: []model.ConfigEntry{
					model.NewConfigEntry("test.config.key", "value", false),
					model.NewConfigEntry("test.secret.key", "secretValue", true),
				},
			},
		},
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
