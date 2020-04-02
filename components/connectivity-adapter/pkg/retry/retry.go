package retry

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	gcli "github.com/machinebox/graphql"
)

func GQLRun(run func(context.Context, *gcli.Request, interface{}) error,
	ctx context.Context,
	req *gcli.Request,
	resp interface{}) error {
	return retry.Do(func() error {
		return run(ctx, req, resp)
	}, defaultRetryOptions()...)
}

func defaultRetryOptions() []retry.Option {
	return []retry.Option{
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(100 * time.Millisecond),
	}
}
