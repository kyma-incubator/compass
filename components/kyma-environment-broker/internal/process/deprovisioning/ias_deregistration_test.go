package deprovisioning

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ias/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const iasInstanceID = "9b130e29-7f1c-4778-8f0a-b9110304cf27"

func TestIASDeregistration_Run(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()

	bundle := &automock.Bundle{}
	defer bundle.AssertExpectations(t)
	bundle.On("DeleteServiceProvider").Return(nil).Once()

	bundleBuilder := &automock.BundleBuilder{}
	defer bundleBuilder.AssertExpectations(t)
	bundleBuilder.On("NewBundle", iasInstanceID).Return(bundle).Once()

	operation := internal.DeprovisioningOperation{
		Operation: internal.Operation{
			InstanceID: iasInstanceID,
		},
	}

	step := NewIASDeregistration(memoryStorage.Operations(), bundleBuilder)

	// when
	_, repeat, err := step.Run(operation, logrus.New())

	// then
	assert.Equal(t, time.Duration(0), repeat)
	assert.NoError(t, err)
}
