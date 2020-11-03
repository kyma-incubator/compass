package scheduler

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

type Scheduler struct {
	wg          *sync.WaitGroup
	concurrency chan struct{}
	errChan     chan error
	ctx         context.Context
	cancelFunc  context.CancelFunc
}

func New(ctx context.Context, maxConcurrency int) *Scheduler {
	childContext, cancel := context.WithCancel(ctx)
	return &Scheduler{wg: &sync.WaitGroup{}, concurrency: make(chan struct{}, maxConcurrency), ctx: childContext, cancelFunc: cancel}
}

func (s Scheduler) Schedule(f func(ctx context.Context) error) {
	select {
	case <-s.ctx.Done():
		return
	case s.concurrency <- struct{}{}:
	}
	s.wg.Add(1)
	go func() {
		defer func() {
			<-s.concurrency
			s.wg.Done()
		}()
		err := f(s.ctx)
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			case s.errChan <- err:
				return
			}
		}
	}()
}

func (s Scheduler) Wait() error {
	success := make(chan interface{})
	go func() {
		s.wg.Wait()
		close(success)
	}()

	select {
	case <-success:
		return nil
	case err := <-s.errChan:
		s.cancelFunc()
		return errors.Wrap(err, "while fetching packages for apps")
	}
}
