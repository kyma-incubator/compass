package provisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/logger"

	"github.com/stretchr/testify/assert"
)

const (
	iasTypeInstanceID   = "1180670b-9de4-421b-8f76-919faeb34249"
	iasTypeURLDashboard = "http://example.com"
)

func TestIASType_ConfigureType(t *testing.T) {
	// given
	bundles := map[string]*automock.Bundle{
		"dex":     &automock.Bundle{},
		"grafana": &automock.Bundle{},
	}
	bundleBuilder := &automock.BundleBuilder{}
	defer bundleBuilder.AssertExpectations(t)
	for inputID, bundle := range bundles {
		defer bundle.AssertExpectations(t)
		bundle.On("FetchServiceProviderData").Return(nil).Once()
		bundle.On("ServiceProviderName").Return(inputID)
		bundle.On("ConfigureServiceProviderType", iasTypeURLDashboard).Return(nil).Once()
		bundleBuilder.On("NewBundle", iasTypeInstanceID, inputID).Return(bundle, nil).Once()
	}

	step := NewIASType(bundleBuilder, false)

	// when
	repeat, err := step.ConfigureType(internal.ProvisioningOperation{
		Operation: internal.Operation{
			InstanceID: iasTypeInstanceID,
		},
	}, iasTypeURLDashboard, logger.NewLogDummy())

	// then
	assert.Equal(t, time.Duration(0), repeat)
	assert.NoError(t, err)
}
