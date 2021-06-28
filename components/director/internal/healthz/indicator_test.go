package healthz_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/healthz/automock"

	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/stretchr/testify/require"
)

func TestNewIndicator(t *testing.T) {
	t.Run("should return not nil indicator", func(t *testing.T) {
		// GIVEN
		indicatorFunc := dummyIndicatorFunc(nil)

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)

		// THEN
		require.NotNil(t, indicator)
		require.Equal(t, "test", indicator.Name())
		require.NotNil(t, indicator.Status())
		require.NoError(t, indicator.Status().Error())
		require.Equal(t, "", indicator.Status().Details())

	})
}

func TestRun(t *testing.T) {
	t.Run("should return error if interval is <= 0", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		indicatorFunc := dummyIndicatorFunc(nil)
		invalidInterval := healthz.IndicatorConfig{
			Interval: 0,
		}

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)
		indicator.Configure(invalidInterval)
		err := indicator.Run(ctx)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "interval")
		require.NotNil(t, indicator)
	})
	t.Run("should return error if timeout is <= 0", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		indicatorFunc := dummyIndicatorFunc(nil)
		invalidTimeout := healthz.IndicatorConfig{
			Interval: time.Second,
			Timeout:  0,
		}

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)
		indicator.Configure(invalidTimeout)
		err := indicator.Run(ctx)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "timeout")
		require.NotNil(t, indicator)
	})
	t.Run("should return error if initialDelay is < 0", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		indicatorFunc := dummyIndicatorFunc(nil)
		invalidDelay := healthz.IndicatorConfig{
			Interval:     time.Second,
			Timeout:      time.Second,
			InitialDelay: -1,
		}

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)
		indicator.Configure(invalidDelay)
		err := indicator.Run(ctx)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "initial delay")
		require.NotNil(t, indicator)
	})
	t.Run("should return error if threshold is < 0", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()
		indicatorFunc := dummyIndicatorFunc(nil)
		invalidThreshold := healthz.IndicatorConfig{
			Interval:     time.Second,
			Timeout:      time.Second,
			InitialDelay: time.Second,
			Threshold:    -1,
		}

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)
		indicator.Configure(invalidThreshold)
		err := indicator.Run(ctx)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "threshold")
		require.NotNil(t, indicator)
	})
	t.Run("should return context timeout when timeout is reached", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		indicatorFunc := timeOutIndicatorFunc()
		cfg := healthz.IndicatorConfig{
			Name:         "First",
			Interval:     time.Minute,
			Timeout:      time.Nanosecond,
			InitialDelay: 0,
			Threshold:    0,
		}

		// WHEN
		indicator := healthz.NewIndicator("test", indicatorFunc)
		indicator.Configure(cfg)
		err := indicator.Run(ctx)

		// THEN
		require.Eventually(t, func() bool {
			return indicator.Status().Error() != nil
		}, time.Second, time.Second/2)
		require.NoError(t, err)
		require.NotNil(t, indicator)
		require.NotNil(t, indicator.Status())
		require.Error(t, indicator.Status().Error())
		require.Contains(t, indicator.Status().Error().Error(), "timeout")
		require.Contains(t, indicator.Status().Details(), "timeout")
	})
	t.Run("should call function on interval time", func(t *testing.T) {
		// GIVEN
		var counter uint64

		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cfg := healthz.IndicatorConfig{
			Name:         "First",
			Interval:     time.Millisecond,
			Timeout:      time.Second,
			InitialDelay: 0,
			Threshold:    0,
		}
		status := &automock.Status{}
		status.On("Error").Return(nil)

		// WHEN
		indicator := healthz.NewIndicator("test", func(ctx context.Context) healthz.Status {
			atomic.AddUint64(&counter, 1)
			return status

		})
		indicator.Configure(cfg)
		err := indicator.Run(ctx)

		// THEN
		require.Eventually(t, func() bool {
			fmt.Println(atomic.LoadUint64(&counter))
			return atomic.LoadUint64(&counter) >= 4
		}, 5*time.Millisecond, time.Millisecond)
		require.NoError(t, err)
		require.NotNil(t, indicator)
	})
	t.Run("should respect the threshold", func(t *testing.T) {
		// GIVEN
		ctx, cancel := context.WithCancel(context.TODO())
		defer cancel()

		cfg := healthz.IndicatorConfig{
			Name:         "First",
			Interval:     time.Millisecond,
			Timeout:      time.Second,
			InitialDelay: 0,
			Threshold:    3,
		}
		status := &automock.Status{}
		status.On("Error").Return(errors.New("some error"))
		status.On("Details").Return("some details")
		// WHEN
		indicator := healthz.NewIndicator("test", func(ctx context.Context) healthz.Status {
			return status
		})
		indicator.Configure(cfg)
		err := indicator.Run(ctx)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, indicator)

		require.NoError(t, indicator.Status().Error())
		require.Eventually(t, func() bool {
			return indicator.Status().Error() != nil
		}, 5*time.Millisecond, time.Millisecond)
	})
}

func dummyIndicatorFunc(status *automock.Status) func(ctx context.Context) healthz.Status {
	return func(ctx context.Context) healthz.Status {
		return status
	}
}

func timeOutIndicatorFunc() func(ctx context.Context) healthz.Status {
	status := &automock.Status{}
	status.On("Error").Return(errors.New("timeout")).Times(5)
	status.On("Details").Return("some timeout details").Times(2)
	return func(ctx context.Context) healthz.Status {
		select {
		case <-ctx.Done():
			return status
		case <-time.After(time.Second):
		}
		return nil
	}
}
