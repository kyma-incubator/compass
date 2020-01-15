package broker

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"

	"github.com/stretchr/testify/assert"
)

func TestGCP(t *testing.T) {
	// given
	b := newProvisioningParamsBuilder(&gcpInputProvider{})

	// when
	b.ApplyParameters(&internal.ProvisioningParametersDTO{
		Name: "gcp-cluster",
	})
	input := b.ClusterConfigInput()

	// then
	assert.Equal(t, "gcp", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "gcp-cluster", input.ClusterConfig.GardenerConfig.Name)
}

func TestAzure(t *testing.T) {
	// given
	b := newProvisioningParamsBuilder(&azureInputProvider{})

	// when
	b.ApplyParameters(&internal.ProvisioningParametersDTO{
		Name: "azure-cluster",
	})
	input := b.ClusterConfigInput()

	// then
	assert.Equal(t, "azure", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "azure-cluster", input.ClusterConfig.GardenerConfig.Name)
}

func TestAWS(t *testing.T) {
	// given
	b := newProvisioningParamsBuilder(&awsInputProvider{})

	// when
	b.ApplyParameters(&internal.ProvisioningParametersDTO{
		Name: "aws-cluster",
	})
	input := b.ClusterConfigInput()

	// then
	assert.Equal(t, "aws", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "aws-cluster", input.ClusterConfig.GardenerConfig.Name)
}
