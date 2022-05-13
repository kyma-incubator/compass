package healthz

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

// Indicator missing godoc
//go:generate mockery --name=Indicator --output=automock --outpkg=automock --case=underscore --disable-version-string
type Indicator interface {
	Name() string
	Configure(IndicatorConfig)
	Run(ctx context.Context)
	Status() Status
}

// Status missing godoc
//go:generate mockery --name=Status --output=automock --outpkg=automock --case=underscore --disable-version-string
type Status interface {
	Error() error
	Details() string
}

// IndicatorFunc missing godoc
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

// NewIndicator returns new indicator with the provided name and IndicatorFunc
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
func (i *indicator) Run(ctx context.Context) {
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
					i.failureCount++
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

			if (i.failureCount > i.threshold || i.failureCount == 0) && i.status.Error() != currentStatus.Error() {
				log.C(ctx).Infof("Changing indicator %s state to %+v", i.Name(), currentStatus)
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
}

// Status reports the last calculated status of the indicator
func (i *indicator) Status() Status {
	i.statusLock.Lock()
	defer i.statusLock.Unlock()
	return i.status
}
