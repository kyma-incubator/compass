package scope_test

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/kyma-incubator/compass/components/director/internal/scope"
	"github.com/kyma-incubator/compass/components/director/internal/scope/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"
)

func TestReloader(t *testing.T) {
	t.Run("file changed twice cause two reloads", func(t *testing.T) {
		// GIVEN
		mockLoader := &automock.Loader{}
		defer mockLoader.AssertExpectations(t)

		mockLoader.On("Load").Return(nil).Twice()

		evChan := make(chan fsnotify.Event)
		errChan := make(chan error)
		mockFileWatcher := configureFileWatcher(evChan, errChan)
		defer mockFileWatcher.AssertExpectations(t)

		reloader := scope.NewReloader("file.data", mockLoader, mockFileWatcher)

		ctx, cancelFunc := context.WithCancel(context.TODO())

		done := make(chan struct{})
		go func(t *testing.T) {
			// WHEN
			err := reloader.Watch(ctx)
			// THEN
			require.Equal(t, context.Canceled, err)
			done <- struct{}{}
		}(t)

		evChan <- fixWriteEvent()
		evChan <- fixWriteEvent()
		cancelFunc()
		<-done
	})

	t.Run("returns error if reload failed", func(t *testing.T) {
		// GIVEN
		mockLoader := &automock.Loader{}
		defer mockLoader.AssertExpectations(t)

		mockLoader.On("Load").Return(fixGivenError()).Once()
		evChan := make(chan fsnotify.Event)
		errChan := make(chan error)
		mockFileWatcher := configureFileWatcher(evChan, errChan)
		defer mockFileWatcher.AssertExpectations(t)

		reloader := scope.NewReloader("file.data", mockLoader, mockFileWatcher)

		done := make(chan struct{})
		go func(t *testing.T) {
			// WHEN
			err := reloader.Watch(context.TODO())
			// THEN
			require.Error(t, err, "some error")
			done <- struct{}{}
		}(t)

		evChan <- fixWriteEvent()
		<-done
	})

	t.Run("returns error if file watch failed", func(t *testing.T) {
		// GIVEN
		evChan := make(chan fsnotify.Event)
		errChan := make(chan error)
		mockFileWatcher := configureFileWatcher(evChan, errChan)
		defer mockFileWatcher.AssertExpectations(t)

		reloader := scope.NewReloader("file.data", nil, mockFileWatcher)

		done := make(chan struct{})
		go func(t *testing.T) {
			// WHEN
			err := reloader.Watch(context.TODO())
			// THEN
			require.Error(t, err, "some error")
			done <- struct{}{}
		}(t)

		errChan <- fixGivenError()
		<-done
	})

	t.Run("returns error if adding file to watch failed", func(t *testing.T) {
		// GIVEN
		mockFileWatcher := &automock.FileWatcher{}
		mockFileWatcher.On("Add", "file.data").Return(fixGivenError())
		mockFileWatcher.On("Close").Return(nil)
		mockFileWatcher.On("FileChangeEventsChannel").Return(make(chan fsnotify.Event)).Once()
		mockFileWatcher.On("ErrorsChannel").Return(make(chan error)).Once()

		defer mockFileWatcher.AssertExpectations(t)

		reloader := scope.NewReloader("file.data", nil, mockFileWatcher)

		done := make(chan struct{})
		go func(t *testing.T) {
			// WHEN
			err := reloader.Watch(context.TODO())
			// THEN
			require.Error(t, err, "some error")
			done <- struct{}{}
		}(t)
		<-done
	})
}

func TestReloaderWithFileWatcherAdapter(t *testing.T) {
	watchedFile := "testdata/watched_file.txt"

	loadExecuted := make(chan struct{})
	dummyLoader := dummyLoader{
		loaded: loadExecuted,
	}

	watcher, err := scope.NewFileWatcher()
	require.NoError(t, err)
	reloader := scope.NewReloader(watchedFile, &dummyLoader, watcher)
	done := make(chan struct{})

	ctx, cancelFunc := context.WithCancel(context.TODO())
	go func(t *testing.T) {
		// WHEN
		err := reloader.Watch(ctx)
		// THEN
		fmt.Println("watch got error",err)
		require.Equal(t,context.Canceled, err)
		done <- struct{}{}
	}(t)

	writeToFile(t, watchedFile, time.Now().String())
	<-loadExecuted
	cancelFunc()
	<-done
	//writeToFile(t, watchedFile, "default content")

}

type dummyLoader struct {
	loaded chan struct{}
}

func (d *dummyLoader) Load() error {
	go func() {
		d.loaded <- struct{}{}
	}()
	return nil
}

func writeToFile(t *testing.T, fileName, content string) {
	err := ioutil.WriteFile(fileName, []byte(content), 0660)
	require.NoError(t, err)
}

func fixGivenError() error {
	return errors.New("some error")
}

func fixWriteEvent() fsnotify.Event {
	return fsnotify.Event{Op: fsnotify.Write}
}

func configureFileWatcher(evChan chan fsnotify.Event, errChan chan error) *automock.FileWatcher {
	mockFileWatcher := &automock.FileWatcher{}
	mockFileWatcher.On("Add", "file.data").Return(nil)
	mockFileWatcher.On("Close").Return(nil)
	mockFileWatcher.On("FileChangeEventsChannel").Return(evChan).Once()
	mockFileWatcher.On("ErrorsChannel").Return(errChan).Once()
	return mockFileWatcher
}
