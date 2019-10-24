package assertions

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertOperationSucceed(t *testing.T, expectedType gqlschema.OperationType, expectedRuntimeId string, operation gqlschema.OperationStatus) {
	AssertOperation(t, gqlschema.OperationStateSucceeded, expectedType, expectedRuntimeId, operation)
}

func AssertOperation(t *testing.T, expectedState gqlschema.OperationState, expectedType gqlschema.OperationType, expectedRuntimeId string, operation gqlschema.OperationStatus) {
	require.NotNil(t, operation.ID)
	require.NotNil(t, operation.Message)

	logrus.Infof("Assering operation %s is in %s state.", *operation.ID, expectedState)
	logrus.Infof("Operation message: %s", *operation.Message)
	require.Equal(t, expectedState, operation.State)
	assert.Equal(t, expectedType, operation.Operation)
	AssertNotNillAndEqualString(t, expectedRuntimeId, operation.RuntimeID)
}
