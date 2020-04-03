package operation

import (
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
	"sync"
	"time"
)

type OperationQueue interface {
	Add(processId string)
	Run(stop <-chan struct{})
}

type ProcessingResult struct {
	Requeue bool
	Delay   time.Duration
}

const (
	workersAmount = 5
)

type Executor interface {
	Execute(operationID string) ProcessingResult
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

func createWorker(queue workqueue.RateLimitingInterface, process func(id string) ProcessingResult, stopCh <-chan struct{}, waitGroup *sync.WaitGroup) {
	waitGroup.Add(1)
	go func() {
		wait.Until(worker(queue, process), time.Second, stopCh)
		waitGroup.Done()
	}()
}

func worker(queue workqueue.RateLimitingInterface, process func(key string) ProcessingResult) func() {
	return func() {
		exit := false
		for !exit {
			exit = func() bool {
				key, quit := queue.Get()
				logrus.Infof("PROCESSING KEY: %s", key)

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
				//if err != nil {
				//	recoverableErr := RecoverableError{}
				//	if errors.As(err, &recoverableErr) {
				//		// TODO: log
				//		queue.AddAfter(key, result.Delay)
				//		return false
				//	}
				//
				//	queue.Forget(key)
				//	return false
				//	// TODO: log fail
				//}

				if result.Requeue {
					queue.AddAfter(key, result.Delay)
					return false
				}

				//if err == nil && when != 0 {
				//	logrus.Infof("Adding %q item after %s", key.(string), when)
				//	queue.AddAfter(key, when)
				//	return false
				//}
				//if err != nil {
				//	logrus.Errorf("Error from process: %s", err)
				//}

				queue.Forget(key)
				return false
			}()
		}
	}
}
