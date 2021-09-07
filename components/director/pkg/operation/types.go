package operation

import (
	"context"
)

// Scheduler missing godoc
// Scheduler is responsible for scheduling any provided Operation entity for later processing
//go:generate mockery --name=Scheduler --output=automock --outpkg=automock --case=underscore
type Scheduler interface {
	Schedule(ctx context.Context, op *Operation) (string, error)
}
