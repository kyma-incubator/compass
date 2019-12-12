package authenticator

import (
	"context"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
)

const ConsumerKey = "consumer type"

var NoConsumerError = errors.New("cannot read consumer from context")

func LoadFromContext(ctx context.Context) (tenantmapping.Consumer, error) {
	value := ctx.Value(ConsumerKey)

	consumer, ok := value.(tenantmapping.Consumer)

	if !ok {
		return tenantmapping.Consumer{}, NoConsumerError
	}

	return consumer, nil
}

func SaveToContext(ctx context.Context, consumer tenantmapping.Consumer) context.Context {
	return context.WithValue(ctx, ConsumerKey, consumer)
}
