package retry

import (
	"github.com/avast/retry-go"
	"time"
)

func DefaultRetryOptions() []retry.Option {
	return []retry.Option{
		retry.Attempts(2),
		retry.DelayType(retry.FixedDelay),
		retry.Delay(100 * time.Millisecond),
	}
}
