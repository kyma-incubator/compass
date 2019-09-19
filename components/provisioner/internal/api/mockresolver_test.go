package api

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ProvisionRuntime(t *testing.T) {
	repository := make(map[string]RuntimeOperation)

	resolver := NewMockResolver(repository)
	runtimeID := "1234"

	t.Run("Should return OperationID when runtime provisioning starts", func(t *testing.T) {
		//given
		ctx := context.Background()
		input := &gqlschema.ProvisionRuntimeInput{}

		//when
		id, e := resolver.ProvisionRuntime(ctx, runtimeID, input)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()
		input := &gqlschema.ProvisionRuntimeInput{}

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
		}

		resolver.repository[runtimeID] = operation

		//when
		emptyId, e := resolver.ProvisionRuntime(ctx, runtimeID, input)

		//then
		require.Error(t, e)
		require.Empty(t, emptyId)

		//cleanup
		flushRepository(resolver)
	})
}

func TestResolver_ReconnectRuntimeAgent(t *testing.T) {
	repository := make(map[string]RuntimeOperation)

	resolver := NewMockResolver(repository)
	runtimeID := "1234"

	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			operationID:   operationID,
		}

		resolver.repository[operationID] = provision

		//when
		id, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   operationID,
		}

		resolver.repository[provisionID] = provision
		resolver.repository[provisionID] = operation

		//when
		emptyID, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when runtime has been deprovisioned", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   operationID,
		}

		resolver.repository[provisionID] = provision
		resolver.repository[operationID] = operation

		//when
		emptyID, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		flushRepository(resolver)
	})
}

func TestResolver_UpgradeRuntime(t *testing.T) {
	repository := make(map[string]RuntimeOperation, 0)

	resolver := NewMockResolver(repository)

	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"
	runtimeID := "1234"

	input := &gqlschema.UpgradeRuntimeInput{}

	t.Run("Should return OperationID when runtime upgrade starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		resolver.repository[provisionID] = provision

		//when
		id, e := resolver.UpgradeRuntime(ctx, runtimeID, input)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		resolver.repository[provisionID] = provision
		resolver.repository[operationID] = operation

		//when
		emptyId, e := resolver.UpgradeRuntime(ctx, runtimeID, input)

		//then
		require.Error(t, e)
		require.Empty(t, emptyId)

		//cleanup
		flushRepository(resolver)
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	repository := make(map[string]RuntimeOperation, 0)

	resolver := NewMockResolver(repository)

	runtimeID := "1234"
	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		resolver.repository[provisionID] = provision

		//when
		id, e := resolver.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
		}

		resolver.repository[provisionID] = provision
		resolver.repository[operationID] = operation

		//when
		emptyID, e := resolver.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		flushRepository(resolver)
	})
}

func TestResolver_RuntimeOperationStatus(t *testing.T) {
	repository := make(map[string]RuntimeOperation, 0)

	resolver := NewMockResolver(repository)

	runtimeID := "1234"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		ctx := context.Background()

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     runtimeID,
		}

		resolver.repository[operationID] = operation

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateInProgress)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return status of previous operation", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"
		secondOperationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     id,
		}

		secondOperation := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   secondOperationID,
			runtimeID:     id,
		}

		resolver.repository[operationID] = operation
		resolver.repository[secondOperationID] = secondOperation

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, secondOperationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeProvision)
		assert.Equal(t, status.State, gqlschema.OperationStateSucceeded)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("Should return error when OperationID is not correct", func(t *testing.T) {
		//given
		ctx := context.Background()
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, operationID)

		//then
		require.Error(t, e)
		require.Empty(t, status)
	})

	t.Run("Should return operation status Succeeded after second call", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     id,
		}

		resolver.repository[operationID] = operation

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateInProgress)

		//when
		status, e = resolver.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateSucceeded)

		//cleanup
		flushRepository(resolver)
	})
}

func TestResolver_RuntimeStatus(t *testing.T) {
	repository := make(map[string]RuntimeOperation, 0)

	resolver := NewMockResolver(repository)

	t.Run("should return last operation status", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"
		secondOperationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"
		thirdOperationID := "51015a1a-3719-4e24-ba89-4971bc762agh3"

		operation := RuntimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   operationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		secondOperation := RuntimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   secondOperationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		thirdOperation := RuntimeOperation{
			operationType: gqlschema.OperationTypeUpgrade,
			status:        gqlschema.OperationStateInProgress,
			operationID:   thirdOperationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		resolver.repository[operationID] = operation
		resolver.repository[secondOperationID] = secondOperation
		resolver.repository[thirdOperationID] = thirdOperation

		//when
		status, e := resolver.RuntimeStatus(ctx, id)

		//then
		require.NoError(t, e)
		assert.Equal(t, gqlschema.OperationTypeUpgrade, status.LastOperationStatus.Operation)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.LastOperationStatus.State)

		//cleanup
		flushRepository(resolver)
	})

	t.Run("should return error when runtime does not exists", func(t *testing.T) {
		ctx := context.Background()
		id := "1234"

		//when
		_, e := resolver.RuntimeStatus(ctx, id)

		//then
		require.Error(t, e)
	})
}

func flushRepository(resolver *MockResolver) {
	resolver.repository = make(map[string]RuntimeOperation)
}
