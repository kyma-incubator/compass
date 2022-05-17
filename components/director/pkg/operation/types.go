package operation

import (
	"context"
)

// Scheduler is responsible for scheduling any provided Operation entity for later processing
//go:generate mockery --name=Scheduler --output=automock --outpkg=automock --case=underscore --disable-version-string
type Scheduler interface {
	Schedule(ctx context.Context, op *Operation) (string, error)
}
