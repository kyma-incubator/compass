package consumer

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

type contextKey string

// ConsumerKey missing godoc
const ConsumerKey contextKey = "consumer"

// NoConsumerError missing godoc
var NoConsumerError = apperrors.NewInternalError("cannot read consumer from context")

// LoadFromContext missing godoc
func LoadFromContext(ctx context.Context) (Consumer, error) {
	value := ctx.Value(ConsumerKey)

	consumer, ok := value.(Consumer)

	if !ok {
		return Consumer{}, NoConsumerError
	}

	return consumer, nil
}

// SaveToContext missing godoc
func SaveToContext(ctx context.Context, consumer Consumer) context.Context {
	return context.WithValue(ctx, ConsumerKey, consumer)
}
