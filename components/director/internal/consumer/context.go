package consumer

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const ConsumerKey = "consumer"

var NoConsumerError = apperrors.NewInternalError("cannot read consumer from context")

func LoadFromContext(ctx context.Context) (Consumer, error) {
	value := ctx.Value(ConsumerKey)

	consumer, ok := value.(Consumer)

	if !ok {
		return Consumer{}, NoConsumerError
	}

	return consumer, nil
}

func SaveToContext(ctx context.Context, consumer Consumer) context.Context {
	return context.WithValue(ctx, ConsumerKey, consumer)
}
