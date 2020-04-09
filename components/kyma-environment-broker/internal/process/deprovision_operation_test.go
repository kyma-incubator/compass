package process

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

func Test_Deprovision_RetryOperationOnce(t *testing.T) {
	// given
	memory := storage.NewMemoryStorage()
	operations := memory.Operations()
	opManager := NewDeprovisionOperationManager(operations)
	op := internal.DeprovisioningOperation{}
	op.UpdatedAt = time.Now()
	retryInterval := time.Hour
	errorMessage := fmt.Sprintf("ups ... ")

	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := operations.InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// then - first call
	op, when, err := opManager.RetryOperationOnce(op, errorMessage, retryInterval, fixLogger())

	// when - first retry
	assert.True(t, when > 0)
	assert.Nil(t, err)

	// then - second call
	t.Log(op.UpdatedAt.String())
	op.UpdatedAt = op.UpdatedAt.Add(-retryInterval - time.Second) // simulate wait of first retry
	t.Log(op.UpdatedAt.String())
	op, when, err = opManager.RetryOperationOnce(op, errorMessage, retryInterval, fixLogger())

	// when - second call => no retry
	assert.True(t, when == 0)
	assert.NotNil(t, err)
}

func Test_Deprovision_RetryOperationWithoutFail(t *testing.T) {
	// given
	memory := storage.NewMemoryStorage()
	operations := memory.Operations()
	opManager := NewDeprovisionOperationManager(operations)
	op := internal.DeprovisioningOperation{}
	op.UpdatedAt = time.Now()
	retryInterval := time.Hour
	errorMessage := fmt.Sprintf("ups ... ")

	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := operations.InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// then - first call
	op, when, err := opManager.RetryOperationWithoutFail(op, errorMessage, retryInterval, retryInterval+1, fixLogger())

	// when - first retry
	assert.True(t, when > 0)
	assert.Nil(t, err)

	// then - second call
	t.Log(op.UpdatedAt.String())
	op.UpdatedAt = op.UpdatedAt.Add(-retryInterval - time.Second) // simulate wait of first retry
	t.Log(op.UpdatedAt.String())
	op, when, err = opManager.RetryOperationWithoutFail(op, errorMessage, retryInterval, retryInterval+1, fixLogger())

	// when - second call => no retry
	assert.True(t, when == 0)
	assert.NoError(t, err)
}

func Test_Deprovision_RetryOperation(t *testing.T) {
	// given
	memory := storage.NewMemoryStorage()
	operations := memory.Operations()
	opManager := NewDeprovisionOperationManager(operations)
	op := internal.DeprovisioningOperation{}
	op.UpdatedAt = time.Now()
	retryInterval := time.Hour
	errorMessage := fmt.Sprintf("ups ... ")
	maxtime := time.Hour * 3 // allow 2 retries

	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := operations.InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// then - first call
	op, when, err := opManager.RetryOperation(op, errorMessage, retryInterval, maxtime, fixLogger())

	// when - first retry
	assert.True(t, when > 0)
	assert.Nil(t, err)

	// then - second call
	t.Log(op.UpdatedAt.String())
	op.UpdatedAt = op.UpdatedAt.Add(-retryInterval - time.Second) // simulate wait of first retry
	t.Log(op.UpdatedAt.String())
	op, when, err = opManager.RetryOperation(op, errorMessage, retryInterval, maxtime, fixLogger())

	// when - second call => retry
	assert.True(t, when > 0)
	assert.Nil(t, err)
}

func fixLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}
