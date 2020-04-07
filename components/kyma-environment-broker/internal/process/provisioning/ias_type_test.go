package provisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias/automock"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	iasTypeInstanceID   = "1180670b-9de4-421b-8f76-919faeb34249"
	iasTypeURLDashboard = "http://examplle.com"
)

func TestIASType_ConfigureType(t *testing.T) {
	// given
	bundle := &automock.Bundle{}
	defer bundle.AssertExpectations(t)
	bundle.On("FetchServiceProviderData").Return(nil).Once()
	bundle.On("ConfigureServiceProviderType", iasTypeURLDashboard).Return(nil).Once()

	bundleBuilder := &automock.BundleBuilder{}
	bundleBuilder.On("NewBundle", iasTypeInstanceID).Return(bundle).Once()
	defer bundleBuilder.AssertExpectations(t)

	step := NewIASType(bundleBuilder, false)

	// when
	repeat, err := step.ConfigureType(internal.ProvisioningOperation{
		Operation: internal.Operation{
			InstanceID: iasTypeInstanceID,
		},
	}, iasTypeURLDashboard, logrus.New())

	// then
	assert.Equal(t, time.Duration(0), repeat)
	assert.NoError(t, err)
}
