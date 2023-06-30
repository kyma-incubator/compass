package consumertest

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/stretchr/testify/mock"
)

// CtxWithRuntimeConsumerMatcher matches contexts with consumer of type runtime
func CtxWithRuntimeConsumerMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		consumerFromCtx, err := consumer.LoadFromContext(ctx)
		return err == nil && consumerFromCtx.ConsumerType == consumer.Runtime
	})
}

// CtxWithConsumerMatcher matches contexts with consumer
func CtxWithConsumerMatcher() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		_, err := consumer.LoadFromContext(ctx)
		return err == nil
	})
}
