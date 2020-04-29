package provisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/logger"
	provisioningAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/stretchr/testify/assert"
)

const (
	iasInstanceID   = "cebd62ee-a32d-4dad-ad19-89dd12b0730e"
	iasClentID      = "1234id"
	iasClientSecret = "4567secret"
)

func TestIASRegistration_Run(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	grafanaOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_ID",
			Value:  iasClentID,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET",
			Value:  iasClientSecret,
			Secret: ptr.Bool(true),
		},
	}
	bundles := map[string]*automock.Bundle{
		"dex":     &automock.Bundle{},
		"grafana": &automock.Bundle{},
	}

	bundleBuilder := &automock.BundleBuilder{}
	defer bundleBuilder.AssertExpectations(t)

	for inputID, bundle := range bundles {
		defer bundle.AssertExpectations(t)
		bundle.On("ServiceProviderName").Return(inputID)
		bundle.On("FetchServiceProviderData").Return(nil).Once()
		bundle.On("ServiceProviderExist").Return(false).Once()
		bundle.On("CreateServiceProvider").Return(nil).Once()
		bundle.On("ConfigureServiceProvider").Return(nil).Once()
		bundleBuilder.On("NewBundle", iasInstanceID, inputID).Return(bundle, nil).Once()
	}
	bundles["grafana"].On("GetProvisioningOverrides").Return("monitoring", grafanaOverrides).Once()
	bundles["dex"].On("GetProvisioningOverrides").Return("", nil).Once()

	inputCreatorMock := &provisioningAutomock.ProvisionInputCreator{}
	defer inputCreatorMock.AssertExpectations(t)
	inputCreatorMock.On("AppendOverrides", "monitoring", grafanaOverrides).Return(nil).Once()
	operation := internal.ProvisioningOperation{
		Operation: internal.Operation{
			InstanceID: iasInstanceID,
		},
		InputCreator: inputCreatorMock,
	}

	step := NewIASRegistrationStep(memoryStorage.Operations(), bundleBuilder)

	// when
	_, repeat, err := step.Run(operation, logger.NewLogDummy())

	// then
	assert.Equal(t, time.Duration(0), repeat)
	assert.NoError(t, err)
}
