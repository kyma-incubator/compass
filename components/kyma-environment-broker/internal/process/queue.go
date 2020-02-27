package process

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

const (
	workersAmount = 5
)

type Executor interface {
	Execute(operationID string) (time.Duration, error)
}

type Queue struct {
	queue    workqueue.RateLimitingInterface
	executor Executor
}

func NewQueue(executor Executor) *Queue {
	return &Queue{
		queue:    workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "operations"),
		executor: executor,
	}
}

func (q *Queue) Add(processId string) {
	q.queue.Add(processId)
}

func (q *Queue) Run(stop <-chan struct{}) {
	var waitGroup sync.WaitGroup

	for i := 0; i < workersAmount; i++ {
		createWorker(q.queue, q.executor.Execute, stop, &waitGroup)
	}
}

func createWorker(queue workqueue.RateLimitingInterface, process func(id string) (time.Duration, error), stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(worker(queue, process), time.Second, stopCh)
		waitGroup.Done()
	}()
}

func worker(queue workqueue.RateLimitingInterface, process func(key string) (time.Duration, error)) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				if quit {
					return true
				}
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("panic error from process: %s", err)
					}
					queue.Done(key)
				}()

				when, err := process(key.(string))
				if err == nil && when != 0 {
					logrus.Infof("Adding %q item after %s", key.(string), when)
					queue.AddAfter(key, when)
					return false
				}
				if err != nil {
					logrus.Errorf("Error from process: %s", err)
				}

				queue.Forget(key)
				return false
			}()
		}
	}
}
