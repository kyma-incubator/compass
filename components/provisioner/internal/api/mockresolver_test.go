package api

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ProvisionRuntime(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)
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
		resolver.cache.Flush()
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()
		input := &gqlschema.ProvisionRuntimeInput{}

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
		}

		resolver.cache.Set("123", operation, 0)

		//when
		emptyId, e := resolver.ProvisionRuntime(ctx, runtimeID, input)

		//then
		require.Error(t, e)
		require.Empty(t, emptyId)

		//cleanup
		resolver.cache.Flush()
	})
}

func TestResolver_ReconnectRuntimeAgent(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)
	runtimeID := "1234"

	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			operationID:   operationID,
		}

		resolver.cache.Set(operationID, provision, 0)

		//when
		id, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   operationID,
		}

		resolver.cache.Set(provisionID, provision, 0)
		resolver.cache.Set(operationID, operation, 0)

		//when
		emptyID, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return error when runtime has been deprovisioned", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   operationID,
		}

		resolver.cache.Set(provisionID, provision, 0)
		resolver.cache.Set(operationID, operation, 0)

		//when
		emptyID, e := resolver.ReconnectRuntimeAgent(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		resolver.cache.Flush()
	})
}

func TestResolver_UpgradeRuntime(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)

	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"
	runtimeID := "1234"

	input := &gqlschema.UpgradeRuntimeInput{}

	t.Run("Should return OperationID when runtime upgrade starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
			operationID:   provisionID,
		}

		resolver.cache.Set(provisionID, provision, 0)

		//when
		id, e := resolver.UpgradeRuntime(ctx, runtimeID, input)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		resolver.cache.Set(provisionID, provision, 0)
		resolver.cache.Set(operationID, operation, 0)

		//when
		emptyId, e := resolver.UpgradeRuntime(ctx, runtimeID, input)

		//then
		require.Error(t, e)
		require.Empty(t, emptyId)

		//cleanup
		resolver.cache.Flush()
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)

	runtimeID := "1234"
	provisionID := "51015a1a-3719-4e24-ba89-4971bc689e86"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		resolver.cache.Set(provisionID, provision, 0)

		//when
		id, e := resolver.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.NoError(t, e)
		require.NotEmpty(t, id)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return error when another operation is in progress", func(t *testing.T) {
		//given
		ctx := context.Background()

		provision := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			runtimeID:     runtimeID,
			startTime:     time.Now(),
		}

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			runtimeID:     runtimeID,
		}

		resolver.cache.Set(provisionID, provision, 0)
		resolver.cache.Set(operationID, operation, 0)

		//when
		emptyID, e := resolver.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.Error(t, e)
		require.Empty(t, emptyID)

		//cleanup
		resolver.cache.Flush()
	})
}

func TestResolver_RuntimeOperationStatus(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)

	runtimeID := "1234"
	operationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		ctx := context.Background()

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     runtimeID,
		}

		resolver.cache.Set(operationID, operation, 0)

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateInProgress)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return status of previous operation", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"
		secondOperationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     id,
		}

		secondOperation := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   secondOperationID,
			runtimeID:     id,
		}

		resolver.cache.Set(operationID, operation, 0)
		resolver.cache.Set(secondOperationID, secondOperation, 0)

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, secondOperationID)

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeProvision)
		assert.Equal(t, status.State, gqlschema.OperationStateSucceeded)

		//cleanup
		resolver.cache.Flush()
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

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
			runtimeID:     id,
		}

		resolver.cache.Set(operationID, operation, 0)

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
		resolver.cache.Flush()
	})
}

func TestResolver_RuntimeStatus(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)

	t.Run("should return last operation status", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"
		secondOperationID := "51015a1a-3719-4e24-ba89-4971bc762ef9"
		thirdOperationID := "51015a1a-3719-4e24-ba89-4971bc762agh3"

		operation := runtimeOperation{
			operationType: gqlschema.OperationTypeProvision,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   operationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		secondOperation := runtimeOperation{
			operationType: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateSucceeded,
			operationID:   secondOperationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		thirdOperation := runtimeOperation{
			operationType: gqlschema.OperationTypeUpgrade,
			status:        gqlschema.OperationStateInProgress,
			operationID:   thirdOperationID,
			runtimeID:     id,
			startTime:     time.Now(),
		}

		resolver.cache.Set(operationID, operation, 0)
		resolver.cache.Set(secondOperationID, secondOperation, 0)
		resolver.cache.Set(thirdOperationID, thirdOperation, 0)

		//when
		status, e := resolver.RuntimeStatus(ctx, id)

		//then
		require.NoError(t, e)
		assert.Equal(t, gqlschema.OperationTypeUpgrade, status.LastOperationStatus.Operation)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.LastOperationStatus.State)

		//cleanup
		resolver.cache.Flush()
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
