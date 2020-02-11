package hyperscaler

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var defaultTenant = "default-tenant"

func TestGardenerSecretNamePreAssigned(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		TargetSecret: "pre-assigned-secret",
	}

	secretName, err := accountProvider.GardenerSecretName(configInput, defaultTenant)

	assert.Equal(t, "pre-assigned-secret", secretName)
	assert.Nil(t, err)
}

func TestCompassSecretNamePreAssigned(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(pool, nil)

	input := &gqlschema.ProvisionRuntimeInput{
		Credentials: &gqlschema.CredentialsInput{
			SecretName: "pre-assigned-secret",
		},
	}

	secretName, err := accountProvider.CompassSecretName(input, defaultTenant)

	assert.Equal(t, "pre-assigned-secret", secretName)
	assert.Nil(t, err)
}

func TestGardenerSecretNamePool(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		Provider:     "AWS",
		TargetSecret: "",
	}

	secretName, err := accountProvider.GardenerSecretName(configInput, defaultTenant)

	assert.Equal(t, "secret5", secretName)
	assert.Nil(t, err)
}

func TestGardenerSecretNameError(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		Provider:     "bogus",
		TargetSecret: "",
	}

	_, err := accountProvider.GardenerSecretName(configInput, defaultTenant)

	require.Error(t, err)

	assert.Equal(t, "Unknown Hyperscaler provider type: bogus", err.Error())
}

func TestCompassSecretNameError(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(pool, nil)

	input := &gqlschema.ProvisionRuntimeInput{
		Credentials: &gqlschema.CredentialsInput{
			SecretName: "",
		},
	}

	_, err := accountProvider.CompassSecretName(input, defaultTenant)

	require.Error(t, err)

	assert.Contains(t, err.Error(), "Can't determine hyperscaler type")
}

func TestGardenerSecretNameNotFound(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		Provider:     "azure",
		TargetSecret: "",
	}

	_, err := accountProvider.GardenerSecretName(configInput, defaultTenant)

	require.Error(t, err)

	assert.Equal(t, "AccountPool failed to find unassigned secret for hyperscalerType: azure", err.Error())
}

func TestHyperscalerTypeFromProvisionInput(t *testing.T) {

	input := &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{},
		},
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	assert.Equal(t, hyperscalerType, GCP)
	assert.Nil(t, err)
}

func TestHyperscalerTypeFromProvisionInputGardenerGCP(t *testing.T) {

	input := &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Provider: "GCP",
			},
		},
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	assert.Equal(t, hyperscalerType, GCP)
	assert.Nil(t, err)
}

func TestHyperscalerTypeFromProvisionInputGardenerAWS(t *testing.T) {

	input := &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Provider: "AWS",
			},
		},
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	assert.Equal(t, hyperscalerType, AWS)
	assert.Nil(t, err)
}

func TestHyperscalerTypeFromProvisionInputGardenerAZURE(t *testing.T) {

	input := &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Provider: "AZURE",
			},
		},
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)
	assert.Equal(t, hyperscalerType, Azure)
	assert.Nil(t, err)
}

func TestHyperscalerTypeFromProvisionInputGardenerError(t *testing.T) {

	input := &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				Provider: "bogus",
			},
		},
	}

	hyperscalerType, err := HyperscalerTypeFromProvisionInput(input)

	require.Error(t, err)
	assert.Empty(t, hyperscalerType)
	assert.Equal(t, "Unknown Hyperscaler provider type: bogus", err.Error())
}

func TestHyperscalerTypeFromProvisionInputError(t *testing.T) {

	_, err := HyperscalerTypeFromProvisionInput(nil)

	require.Error(t, err)

	assert.Equal(t, err.Error(), "Can't determine hyperscaler type because ProvisionRuntimeInput not specified (was nil)")

	input := &gqlschema.ProvisionRuntimeInput{}

	_, err = HyperscalerTypeFromProvisionInput(input)

	require.Error(t, err)

	assert.Equal(t, err.Error(), "Can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig not specified (was nil)")

	input = &gqlschema.ProvisionRuntimeInput{
		ClusterConfig: &gqlschema.ClusterConfigInput{},
	}

	_, err = HyperscalerTypeFromProvisionInput(input)

	require.Error(t, err)

	assert.Equal(t, err.Error(), "Can't determine hyperscaler type because ProvisionRuntimeInput.ClusterConfig hyperscaler config not specified")
}
