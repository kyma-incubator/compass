package healthz

import (
	"context"
)

// DBIndicatorName missing godoc
const DBIndicatorName = "database"

// Pinger missing godoc
//go:generate mockery --name=Pinger --output=automock --outpkg=automock --case=underscore --disable-version-string
type Pinger interface {
	PingContext(ctx context.Context) error
}

// NewDBIndicatorFunc missing godoc
func NewDBIndicatorFunc(p Pinger) IndicatorFunc {
	return func(ctx context.Context) Status {
		if err := p.PingContext(ctx); err != nil {
			return &status{
				error:   err,
				details: "Error pinging database",
			}
		}

		return &status{
			error:   nil,
			details: "",
		}
	}
}
