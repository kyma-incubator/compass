package provisioning

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/edp"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/logger"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage"

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
		Value: "CF",
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
	}, logger.NewLogDummy())

	// then
	assert.Equal(t, 0*time.Second, repeat)
	assert.NoError(t, err)
}

func TestEDPRegistrationStep_selectEnvironmentKey(t *testing.T) {
	for name, tc := range map[string]struct {
		region   string
		expected string
	}{
		"kubernetes region": {
			region:   "k8s-as34",
			expected: "KUBERNETES",
		},
		"cf region": {
			region:   "cf-eu10",
			expected: "CF",
		},
		"neo region": {
			region:   "neo-us13",
			expected: "NEO",
		},
		"default region": {
			region:   "undefined",
			expected: "CF",
		},
		"empty region": {
			region:   "",
			expected: "CF",
		},
	} {
		t.Run(name, func(t *testing.T) {
			// given
			step := NewEDPRegistrationStep(nil, nil, edp.Config{})

			// when
			envKey := step.selectEnvironmentKey(tc.region, logger.NewLogDummy())

			// then
			assert.Equal(t, tc.expected, envKey)
		})
	}
}
