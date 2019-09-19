package scope

import (
	"context"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
)

//go:generate mockery -name=Loader -output=automock -outpkg=automock -case=underscore
type Loader interface {
	Load() error
}

//go:generate mockery -name=FileWatcher -output=automock -outpkg=automock -case=underscore
type FileWatcher interface {
	io.Closer
	Add(fileName string) error
	FileChangeEventsChannel() chan fsnotify.Event
	ErrorsChannel() chan error
}

type fileWatcherAdapter struct {
	fsnotify.Watcher
}

func (a *fileWatcherAdapter) FileChangeEventsChannel() chan fsnotify.Event {
	return a.Events
}

func (a *fileWatcherAdapter) ErrorsChannel() chan error {
	return a.Errors
}

func NewFileWatcher() (*fileWatcherAdapter, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &fileWatcherAdapter{Watcher: *watcher}, nil
}

type reloader struct {
	fileName    string
	loader      Loader
	fileWatcher FileWatcher
}

func NewReloader(fileName string, loader Loader, fw FileWatcher) *reloader {
	return &reloader{
		fileName:    fileName,
		loader:      loader,
		fileWatcher: fw,
	}
}

func (r *reloader) Watch(ctx context.Context) error {
	defer r.fileWatcher.Close()

	evChan := r.fileWatcher.FileChangeEventsChannel()
	errChan := r.fileWatcher.ErrorsChannel()
	result := make(chan error)

	addFailed := make(chan struct{})
	go func() {
		defer fmt.Println("goroutine ended")
		for {
			select {
			case <-addFailed:
				fmt.Println("add failed")
				return
			case <-ctx.Done():
				fmt.Println("ctx done")
				result <- ctx.Err()
				return
			case e := <-evChan:
				if (e.Op & fsnotify.Write) == fsnotify.Write {
					err := r.loader.Load()
					fmt.Println("load err", err)
					if err != nil {
						result <- err
						return
					}
				}
			case err := <-errChan:
				fmt.Println("errchan")
				result <- err
				return
			}
		}
	}()

	if err := r.fileWatcher.Add(r.fileName); err != nil {
		addFailed <- struct{}{}
		return err
	}

	return <-result
}
