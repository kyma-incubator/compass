package healthz

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

//go:generate mockery --name=Indicator --output=automock --outpkg=automock --case=underscore
type Indicator interface {
	Name() string
	Configure(IndicatorConfig)
	Run(ctx context.Context) error
	Status() Status
}

//go:generate mockery --name=Status --output=automock --outpkg=automock --case=underscore
type Status interface {
	Error() error
	Details() string
}

type IndicatorFunc func(ctx context.Context) Status

// Implements Status interface
type status struct {
	error   error
	details string
}

// Error returns status error
func (s *status) Error() error {
	return s.error
}

// Details returns status details
func (s *status) Details() string {
	return s.details
}

// Implements Indicator interface
type indicator struct {
	name string

	interval     time.Duration
	timeout      time.Duration
	initialDelay time.Duration
	threshold    int

	indicatorFunc IndicatorFunc
	status        Status
	statusLock    sync.Mutex
	failureCount  int
}

// NewIndicator returns new indicator from the provided IndicatorConfig and IndicatorFunc
func NewIndicator(name string, indicatorFunc IndicatorFunc) Indicator {
	return &indicator{
		name:          name,
		indicatorFunc: indicatorFunc,
		status:        &status{details: ""},
		statusLock:    sync.Mutex{},
		failureCount:  0,
	}
}

// Name returns indicator name
func (i *indicator) Name() string {
	return i.name
}

// Configure sets indicator config based on IndicatorConfig
func (i *indicator) Configure(cfg IndicatorConfig) {
	i.interval = cfg.Interval
	i.timeout = cfg.Timeout
	i.initialDelay = cfg.InitialDelay
	i.threshold = cfg.Threshold
}

// Run starts the periodic indicator checks
func (i *indicator) Run(ctx context.Context) error {
	if i.interval <= 0 {
		return errors.New("indicator interval could not be <= 0")
	}
	if i.timeout <= 0 {
		return errors.New("indicator timeout could not be <= 0")
	}
	if i.initialDelay < 0 {
		return errors.New("indicator initial delay could not be < 0")
	}
	if i.threshold < 0 {
		return errors.New("indicator threshold could not be < 0")
	}

	go func() {
		<-time.After(i.initialDelay)

		ticker := time.NewTicker(i.interval)
		for {
			timeoutCtx, cancel := context.WithTimeout(ctx, i.timeout)
			currentStatus := i.indicatorFunc(timeoutCtx)
			cancel()

			i.statusLock.Lock()
			if currentStatus.Error() != nil {
				// escape overflow
				if i.failureCount < math.MaxInt32 {
					i.failureCount += 1
				}
				log.C(ctx).Warnf("Threshold for indicator %s is %d, current failure count is : %d, current error is: %s, details are: %s",
					i.Name(),
					i.threshold,
					i.failureCount,
					currentStatus.Error(),
					currentStatus.Details())
			} else {
				i.failureCount = 0
			}

			if i.failureCount > i.threshold || i.failureCount == 0 {
				log.C(ctx).Debugf("Changing indicator %s state to %+v", i.Name(), currentStatus)
				i.status = currentStatus
			}
			i.statusLock.Unlock()

			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
			}
		}
	}()
	return nil
}

// Status reports the last calculated status of the indicator
func (i *indicator) Status() Status {
	i.statusLock.Lock()
	defer i.statusLock.Unlock()
	return i.status
}
