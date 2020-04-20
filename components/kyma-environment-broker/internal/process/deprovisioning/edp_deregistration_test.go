package deprovisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/edp"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/deprovisioning/automock"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	edpName        = "f88401ba-c601-45bb-bec0-a2156c07c9a6"
	edpEnvironment = "test"
)

func TestEDPDeregistration_Run(t *testing.T) {
	// given
	client := &automock.EDPClient{}
	client.On("DeleteMetadataTenant", edpName, edpEnvironment, edp.MaasConsumerSubAccountKey).
		Return(nil).Once()
	client.On("DeleteMetadataTenant", edpName, edpEnvironment, edp.MaasConsumerRegionKey).
		Return(nil).Once()
	client.On("DeleteMetadataTenant", edpName, edpEnvironment, edp.MaasConsumerEnvironmentKey).
		Return(nil).Once()
	client.On("DeleteDataTenant", edpName, edpEnvironment).
		Return(nil).Once()
	defer client.AssertExpectations(t)

	step := NewEDPDeregistration(client, edp.Config{
		Environment: edpEnvironment,
	})

	// when
	_, repeat, err := step.Run(internal.DeprovisioningOperation{SubAccountID: edpName}, logrus.New())

	// then
	assert.Equal(t, 0*time.Second, repeat)
	assert.NoError(t, err)
}
