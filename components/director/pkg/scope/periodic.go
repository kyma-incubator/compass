package scope

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

//go:generate mockery -name=InfoPrinter -output=automock -outpkg=automock -case=underscore
type InfoPrinter interface {
	Infof(format string, args ...interface{})
}

//go:generate mockery -name=Ticker -output=automock -outpkg=automock -case=underscore
type Ticker interface {
	Stop()
	Ticks() <-chan time.Time
}

//go:generate mockery -name=Loader -output=automock -outpkg=automock -case=underscore
type Loader interface {
	Load() error
}

func NewTicker(d time.Duration) Ticker {
	return &tickerWrapper{
		internal: time.NewTicker(d),
	}
}

type tickerWrapper struct {
	internal *time.Ticker
}

func (t *tickerWrapper) Stop() {
	t.internal.Stop()
}

func (t *tickerWrapper) Ticks() <-chan time.Time {
	return t.internal.C

}

type periodic struct {
	period time.Duration
	load   Loader
	logger InfoPrinter
	ticker Ticker
}

func NewPeriodicReloader(load Loader, logger InfoPrinter, ticker Ticker) *periodic {
	return &periodic{
		load:   load,
		logger: logger,
		ticker: ticker,
	}
}

func (p *periodic) Watch(ctx context.Context) error {
	ticks := p.ticker.Ticks()
	defer p.ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticks:
			err := p.load.Load()
			if err != nil {
				return errors.Wrap(err, "while loading")
			}
			p.logger.Infof("Successfully reloaded scopes configuration")

		}

	}
}
