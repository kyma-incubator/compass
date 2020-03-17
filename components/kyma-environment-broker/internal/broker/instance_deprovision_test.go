package broker

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provisioner"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeprovisionEndpoint_DeprovisionNotExistingInstance(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	instStorage := memoryStorage.Instances()
	provisionerClient := provisioner.NewFakeClient()

	svc := NewDeprovision(instStorage, provisionerClient, logrus.StandardLogger())

	// when
	_, err := svc.Deprovision(context.TODO(), "inst-0001", domain.DeprovisionDetails{}, true)

	// then
	assert.Equal(t, apiresponses.ErrInstanceDoesNotExist, err)
}

func TestDeprovisionEndpoint_DeprovisionExistingInstance(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	instStorage := memoryStorage.Instances()
	instStorage.Insert(internal.Instance{
		InstanceID: "instance-001",
	})
	provisionerClient := provisioner.NewFakeClient()

	svc := NewDeprovision(instStorage, provisionerClient, logrus.StandardLogger())

	// when
	_, err := svc.Deprovision(context.TODO(), "instance-001", domain.DeprovisionDetails{}, true)

	// then
	require.NoError(t, err)

	// the instance must be removed
	_, err = instStorage.GetByID("instance-001")
	assert.True(t, dberr.IsNotFound(err))
}
