package viewer_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/viewer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_Viewer(t *testing.T) {
	cons := consumer.Consumer{
		ConsumerID:   uuid.New().String(),
		ConsumerType: consumer.Application,
	}
	expectedViewer := graphql.Viewer{
		ID:   cons.ConsumerID,
		Type: graphql.ViewerTypeApplication,
	}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		ctx = consumer.SaveToContext(ctx, cons)
		resolver := viewer.NewViewerResolver()

		//WHEN
		vwr, err := resolver.Viewer(ctx)
		//THEN
		require.NoError(t, err)
		require.NotNil(t, vwr)
		assert.Equal(t, expectedViewer, *vwr)
	})

	t.Run("Error while converting", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		invalidConsumer := consumer.Consumer{
			ConsumerID:   uuid.New().String(),
			ConsumerType: "",
		}
		ctx = consumer.SaveToContext(ctx, invalidConsumer)
		resolver := viewer.NewViewerResolver()
		//WHEN
		_, err := resolver.Viewer(ctx)
		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "viewer does not exist")
	})

	t.Run("No consumer in ctx", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		resolver := viewer.NewViewerResolver()

		//WHEN
		_, err := resolver.Viewer(ctx)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while getting viewer from context: cannot read consumer from context")
	})
}
