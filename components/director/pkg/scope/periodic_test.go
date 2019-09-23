package scope_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/scope/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestPeriodic(t *testing.T) {
	t.Run("watch calls Load twice", func(t *testing.T) {
		// GIVEN
		mockLoader := &automock.Loader{}
		defer mockLoader.AssertExpectations(t)
		mockLoader.On("Load").Return(nil).Twice()

		mockPrinter := &automock.InfoPrinter{}
		defer mockPrinter.AssertExpectations(t)
		mockPrinter.On("Infof", "Successfully reloaded scopes configuration").Twice()

		dummyTicker := NewDummyTicker()
		p := scope.NewPeriodicReloader(mockLoader, mockPrinter, dummyTicker)

		ctx, cancelFunc := context.WithCancel(context.TODO())
		done := make(chan struct{})
		go func() {
			// WHEN
			err := p.Watch(ctx)
			// THEN
			require.NoError(t, err)
			done <- struct{}{}
		}()

		dummyTicker.ticks <- time.Now()
		dummyTicker.ticks <- time.Now()
		cancelFunc()
		<-done
		assert.True(t, dummyTicker.stopped)
	})

	t.Run("watch returns error if Load failed", func(t *testing.T) {
		// GIVEN
		mockLoader := &automock.Loader{}
		defer mockLoader.AssertExpectations(t)
		mockLoader.On("Load").Return(fixGivenError()).Once()

		dummyTicker := NewDummyTicker()
		p := scope.NewPeriodicReloader(mockLoader, nil, dummyTicker)

		done := make(chan struct{})
		go func() {
			// WHEN
			err := p.Watch(context.TODO())
			// THEN
			require.EqualError(t,err,"while loading: some error")
			done <- struct{}{}
		}()

		dummyTicker.ticks <- time.Now()
		<-done
		assert.True(t, dummyTicker.stopped)
	})

}

func fixGivenError() error {
	return errors.New("some error")
}

func NewDummyTicker() *dummyTicker {
	return &dummyTicker{
		ticks: make(chan time.Time),
	}
}

type dummyTicker struct {
	ticks   chan time.Time
	stopped bool
}

func (d *dummyTicker) Stop() {
	d.stopped = true
}

func (d *dummyTicker) Ticks() <-chan time.Time {
	return d.ticks
}
