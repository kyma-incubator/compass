package provisioning

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/edp"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	edpName        = "cd4b333c-97fb-4894-bb20-7874f5833e8d"
	edpEnvironment = "test"
	edpRegion      = "cf-eu10"
)

func TestEDPRegistration_Run(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	client := &automock.EDPClient{}
	client.On("CreateDataTenant", edp.DataTenantPayload{
		Name:        edpName,
		Environment: edpEnvironment,
		// it is copy of body `generateSecret` method in `EDPRegistrationStep` step
		Secret: base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s", edpName, edpEnvironment))),
	}).Return(nil).Once()
	client.On("CreateMetadataTenant", edpName, edpEnvironment, edp.MetadataTenantPayload{
		Key:   edp.MaasConsumerEnvironmentKey,
		Value: "KUBERNETES",
	}).Return(nil).Once()
	client.On("CreateMetadataTenant", edpName, edpEnvironment, edp.MetadataTenantPayload{
		Key:   edp.MaasConsumerRegionKey,
		Value: edpRegion,
	}).Return(nil).Once()
	client.On("CreateMetadataTenant", edpName, edpEnvironment, edp.MetadataTenantPayload{
		Key:   edp.MaasConsumerSubAccountKey,
		Value: edpName,
	}).Return(nil).Once()
	defer client.AssertExpectations(t)

	step := NewEDPRegistrationStep(memoryStorage.Operations(), client, edp.Config{
		Environment: edpEnvironment,
		Required:    true,
	})

	// when
	_, repeat, err := step.Run(internal.ProvisioningOperation{
		ProvisioningParameters: `{"platform_region":"` + edpRegion + `", "ers_context":{"subaccount_id":"` + edpName + `"}}`,
	}, logrus.New())

	// then
	assert.Equal(t, 0*time.Second, repeat)
	assert.NoError(t, err)
}
