package mock

import (
	"context"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolver_ProvisionRuntime(t *testing.T) {
	noExpireCache := cache.New(0, 0)

	resolver := NewMockResolver(*noExpireCache)

	t.Run("Should return OperationID when runtime provisioning starts", func(t *testing.T) {
		//given
		ctx := context.Background()
		runtimeID := &gqlschema.RuntimeIDInput{ID: "1234"}
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
		id := "1234"
		runtimeID := &gqlschema.RuntimeIDInput{ID: id}
		input := &gqlschema.ProvisionRuntimeInput{}

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
		}

		resolver.cache.Set(id, operation, 0)

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

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()
		runtimeID := &gqlschema.RuntimeIDInput{ID: "1234"}

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
		id := "1234"
		runtimeID := &gqlschema.RuntimeIDInput{ID: id}

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
		}

		resolver.cache.Set(id, operation, 0)

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

	t.Run("Should return OperationID when runtime provisioning starts", func(t *testing.T) {
		//given
		ctx := context.Background()
		runtimeID := &gqlschema.RuntimeIDInput{ID: "1234"}
		input := &gqlschema.UpgradeRuntimeInput{}

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
		id := "1234"
		runtimeID := &gqlschema.RuntimeIDInput{ID: id}
		input := &gqlschema.UpgradeRuntimeInput{}

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeDeprovision,
			status:        gqlschema.OperationStateInProgress,
		}

		resolver.cache.Set(id, operation, 0)

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

	t.Run("Should return OperationID when runtime reconnect starts", func(t *testing.T) {
		//given
		ctx := context.Background()
		runtimeID := &gqlschema.RuntimeIDInput{ID: "1234"}

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
		id := "1234"
		runtimeID := &gqlschema.RuntimeIDInput{ID: id}

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
		}

		resolver.cache.Set(id, operation, 0)

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

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
		}

		resolver.cache.Set(id, operation, 0)

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, &gqlschema.AsyncOperationIDInput{ID: operationID})

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateInProgress)

		//cleanup
		resolver.cache.Flush()
	})

	t.Run("Should return error when OperationID is not correct", func(t *testing.T) {
		//given
		ctx := context.Background()
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, &gqlschema.AsyncOperationIDInput{ID: operationID})

		//then
		require.Error(t, e)
		require.Empty(t, status)
	})

	t.Run("Should return operation status Succeeded after second call", func(t *testing.T) {
		//given
		ctx := context.Background()
		id := "1234"
		operationID := "51015a1a-3719-4e24-ba89-4971bc689e86"

		operation := currentOperation{
			lastOperation: gqlschema.OperationTypeReconnectRuntime,
			status:        gqlschema.OperationStateInProgress,
			operationID:   operationID,
		}

		resolver.cache.Set(id, operation, 0)

		//when
		status, e := resolver.RuntimeOperationStatus(ctx, &gqlschema.AsyncOperationIDInput{ID: operationID})

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateInProgress)

		//when
		status, e = resolver.RuntimeOperationStatus(ctx, &gqlschema.AsyncOperationIDInput{ID: operationID})

		//then
		require.NoError(t, e)
		assert.Equal(t, status.Operation, gqlschema.OperationTypeReconnectRuntime)
		assert.Equal(t, status.State, gqlschema.OperationStateSucceeded)

		//cleanup
		resolver.cache.Flush()
	})
}
