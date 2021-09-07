package healthz

import (
	"context"
)

const DBIndicatorName = "database"

//go:generate mockery --name=Pinger --output=automock --outpkg=automock --case=underscore
type Pinger interface {
	PingContext(ctx context.Context) error
}

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
