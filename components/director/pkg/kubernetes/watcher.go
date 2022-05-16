package kubernetes

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Watcher is a K8S watcher that listen for resources changes and receive events based on that
//go:generate mockery --name=Watcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type Watcher interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type processWatchEventsFunc func(ctx context.Context, events <-chan watch.Event)

// K8SWatcher is a structure containing dependencies for the watch mechanism
type K8SWatcher struct {
	ctx                    context.Context
	k8sWatcher             Watcher
	processWatchEventsFunc processWatchEventsFunc
	reconnectInterval      time.Duration
	resourceName           string
	correlationID          string
}

// NewWatcher return initialized K8SWatcher structure
func NewWatcher(ctx context.Context, k8sWatcher Watcher, processWatchEventsFunc processWatchEventsFunc, reconnectInterval time.Duration, resourceName, correlationID string) *K8SWatcher {
	return &K8SWatcher{
		ctx:                    ctx,
		k8sWatcher:             k8sWatcher,
		processWatchEventsFunc: processWatchEventsFunc,
		reconnectInterval:      reconnectInterval,
		resourceName:           resourceName,
		correlationID:          correlationID,
	}
}

// Run starts goroutine that uses kubernetes watch mechanism to listen for resource changes
func (w *K8SWatcher) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, w.correlationID)
	ctx = log.ContextWithLogger(ctx, entry)

	w.startKubeWatch(ctx)
}

func (w *K8SWatcher) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Infof("Context cancelled, stopping kubernetes watcher for resource with name: %s...", w.resourceName)
			return
		default:
		}
		log.C(ctx).Info("Starting kubernetes watcher for changes...")
		wr, err := w.k8sWatcher.Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + w.resourceName,
			Watch:         true,
		})
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Could not initialize watcher for resource with name: %s. Sleep for %s and try again... %v", w.resourceName, w.reconnectInterval.String(), err)
			time.Sleep(w.reconnectInterval)
			continue
		}
		log.C(ctx).Infof("Waiting for kubernetes watcher events for resource with name: %s...", w.resourceName)

		w.processWatchEventsFunc(ctx, wr.ResultChan())

		log.C(ctx).Infof("Processed kubernetes watcher events for resource with name: %s", w.resourceName)

		// Cleanup any allocated resources
		wr.Stop()
		time.Sleep(w.reconnectInterval)
	}
}
