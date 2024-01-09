package retry

import (
	"context"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	gcli "github.com/machinebox/graphql"
)

func GQLRun(run func(context.Context, *gcli.Request, interface{}) error,
	ctx context.Context,
	req *gcli.Request,
	resp interface{}) error {

	return GQLRunWithOptions(run, ctx, req, resp, defaultOptions())
}

func GQLRunWithOptions(run func(context.Context, *gcli.Request, interface{}) error,
	ctx context.Context,
	req *gcli.Request,
	resp interface{},
	options []retry.Option) error {

	return retry.Do(
		func() error {
			return run(ctx, req, resp)
		},
		options...,
	)
}

func defaultOptions() []retry.Option {
	return []retry.Option{
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(100 * time.Millisecond),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "connection refused") ||
				strings.Contains(err.Error(), "connection reset by peer")
		}),
	}
}
