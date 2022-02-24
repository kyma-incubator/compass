package consumer_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsumerContext(t *testing.T) {
	t.Run("load returns consumer previously saved in context", func(t *testing.T) {
		// GIVEN
		id := "223da628-3756-4bef-ab48-fb0061a4eae4"
		givenConsumer := consumer.Consumer{ConsumerID: id, ConsumerType: consumer.Runtime}
		ctx := consumer.SaveToContext(context.TODO(), givenConsumer)

		// WHEN
		actual, err := consumer.LoadFromContext(ctx)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenConsumer, actual)
	})
	t.Run("load returns error if consumer not found in ctx", func(t *testing.T) {
		// WHEN
		_, err := consumer.LoadFromContext(context.TODO())
		// THEN
		assert.Equal(t, consumer.NoConsumerError, err)
	})

	t.Run("cannot override consumer accidentally", func(t *testing.T) {
		// GIVEN
		id := "223da628-3756-4bef-ab48-fb0061a4eae4"
		givenConsumer := consumer.Consumer{ConsumerID: id, ConsumerType: consumer.Runtime}
		ctx := consumer.SaveToContext(context.TODO(), givenConsumer)
		ctx = context.WithValue(ctx, 0, "some random value")

		// WHEN
		actual, err := consumer.LoadFromContext(ctx)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenConsumer, actual)
	})
}
