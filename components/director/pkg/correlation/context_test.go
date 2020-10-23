package correlation_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationContext_IDFromContext(t *testing.T) {
	t.Run("when the value is not a string an empty string is returned", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), correlation.ContextField, 123)

		actual := correlation.IDFromContext(ctx)
		assert.Empty(t, actual)
	})
}
