package queue

import (
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

//go:generate mockery -name=OperationQueue
type OperationQueue interface {
	Add(processId string)
	Run(stop <-chan struct{})
}

const (
	workersAmount = 5
)

type Executor interface {
	Execute(operationID string) operations.ProcessingResult
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

func (q *Queue) Add(operationId string) {
	q.queue.Add(operationId)
}

func (q *Queue) Run(stop <-chan struct{}) {
	var waitGroup sync.WaitGroup

	for i := 0; i < workersAmount; i++ {
		createWorker(q.queue, q.executor.Execute, stop, &waitGroup)
	}
}

func createWorker(queue workqueue.RateLimitingInterface, process func(id string) operations.ProcessingResult, stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(worker(queue, process), time.Second, stopCh)
		waitGroup.Done()
	}()
}

func worker(queue workqueue.RateLimitingInterface, process func(key string) operations.ProcessingResult) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				logrus.Debugf("Processing operation: %s", key)

				if quit {
					return true
				}
				defer func() {
					if err := recover(); err != nil {
						logrus.Errorf("panic error while processing key %s: %s", key, err)
					}
					queue.Done(key)
				}()

				result := process(key.(string))
				if result.Requeue {
					queue.AddAfter(key, result.Delay)
					return false
				}

				queue.Forget(key)
				return false
			}()
		}
	}
}
