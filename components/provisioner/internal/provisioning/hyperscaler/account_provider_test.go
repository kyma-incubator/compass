package hyperscaler

import (
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGardenerSecretNamePreAssigned(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		TargetSecret: "pre-assigned-secret",
	}

	secretName, err := accountProvider.GardenerSecretName(configInput)

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

	secretName, err := accountProvider.CompassSecretName(input)

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

	secretName, err := accountProvider.GardenerSecretName(configInput)

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

	_, err := accountProvider.GardenerSecretName(configInput)

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

	_, err := accountProvider.CompassSecretName(input)

	assert.Equal(t, "Unknown Hyperscaler provider type: TBD", err.Error())
}

func TestGardenerSecretNameNotFound(t *testing.T) {

	pool := newTestAccountPool()

	accountProvider := NewAccountProvider(nil, pool)

	configInput := &gqlschema.GardenerConfigInput{
		Provider:     "azure",
		TargetSecret: "",
	}

	_, err := accountProvider.GardenerSecretName(configInput)

	assert.Equal(t, "AccountPool failed to find unassigned secret for hyperscalerType: azure", err.Error())
}
