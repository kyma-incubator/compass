package consumer

import (
	"context"
	"errors"
)

const consumerKey = "consumer"

var NoConsumerError = errors.New("cannot read consumer from context")

func LoadFromContext(ctx context.Context) (Consumer, error) {
	value := ctx.Value(consumerKey)

	consumer, ok := value.(Consumer)

	if !ok {
		return Consumer{}, NoConsumerError
	}

	return consumer, nil
}

func SaveToContext(ctx context.Context, consumer Consumer) context.Context {
	return context.WithValue(ctx, consumerKey, consumer)
}
