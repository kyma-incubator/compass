package revocation_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/hydrator/internal/revocation"
	"github.com/kyma-incubator/compass/components/hydrator/internal/revocation/automock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

const configMapName = "revokedCertificates"

func prep(ctx context.Context, number int) (revocation.Cache, *testWatch, *automock.Manager) {
	cache := revocation.NewCache()
	watcher := &testWatch{
		events: make(chan watch.Event, 100),
	}
	configListManagerMock := &automock.Manager{}
	configListManagerMock.
		On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).
		Return(watcher, nil).
		Times(number)
	loader := revocation.NewRevokedCertificatesLoader(cache, configListManagerMock, configMapName, time.Millisecond)

	go loader.Run(ctx)
	return cache, watcher, configListManagerMock
}

func Test_revokedCertificatesLoader(t *testing.T) {

	t.Run("should load configmap on add event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 1)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.ConfigMap{
				Data: map[string]string{
					"hash": "hash",
				},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			return cache.Get()["hash"] == "hash"
		}, time.Second*2, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

	t.Run("should load configmap on modify event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 1)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Modified,
			Object: &v1.ConfigMap{
				Data: map[string]string{
					"modified": "modified",
				},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			return cache.Get()["modified"] == "modified"
		}, time.Second*2, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

	t.Run("should not load configmap if event object is not configmap", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 1)

		// when
		watcher.putEvent(watch.Event{
			Type:   watch.Added,
			Object: &runtime.Unknown{},
		})

		// then
		assert.Eventually(t, func() bool {
			return len(cache.Get()) == 0
		}, time.Second*2, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

	t.Run("should return empty cache after delete event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 1)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.ConfigMap{
				Data: map[string]string{
					"added": "added",
				},
			},
		})

		assert.Eventually(t, func() bool {
			return cache.Get()["added"] == "added"
		}, time.Second*2, time.Millisecond*100)

		watcher.putEvent(watch.Event{
			Type: watch.Deleted,
		})

		// then
		assert.Eventually(t, func() bool {
			return len(cache.Get()) == 0
		}, time.Second*2, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

	t.Run("should try reconnect when there is error event", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 2)

		// when
		watcher.putEvent(watch.Event{
			Type: watch.Error,
		})

		watcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.ConfigMap{
				Data: map[string]string{
					"hash2": "hash2",
				},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			return cache.Get()["hash2"] == "hash2"
		}, time.Second, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second*2, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

	t.Run("should try reconnect when event channel is closed", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache, watcher, managerMock := prep(ctx, 1)

		// when
		watcher.close()
		newWatcher := &testWatch{
			events: make(chan watch.Event, 100),
		}
		managerMock.On("Watch", mock.Anything, mock.AnythingOfType("v1.ListOptions")).
			Return(newWatcher, nil).Once()

		newWatcher.putEvent(watch.Event{
			Type: watch.Added,
			Object: &v1.ConfigMap{
				Data: map[string]string{
					"hash3": "hash3",
				},
			},
		})

		// then
		assert.Eventually(t, func() bool {
			return cache.Get()["hash3"] == "hash3"
		}, time.Second, time.Millisecond*100)
		cancel()
		assert.Eventually(t, func() bool {
			<-ctx.Done()
			return true
		}, time.Second*2, time.Millisecond*100)
		managerMock.AssertExpectations(t)
	})

}

type testWatch struct {
	events chan watch.Event
}

func (tw *testWatch) close() {
	close(tw.events)
}

func (tw *testWatch) putEvent(ev watch.Event) {
	tw.events <- ev
}

func (tw *testWatch) Stop() {}
func (tw *testWatch) ResultChan() <-chan watch.Event {
	return tw.events
}
