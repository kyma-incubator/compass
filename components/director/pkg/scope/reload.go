package scope

import (
	"context"
	"fmt"
	"io"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
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

func NewReloader(fileName string, loader Loader, fw FileWatcher) (*reloader, error) {
	err := fw.Add(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "while adding file %s to watch", fileName)
	}
	return &reloader{
		fileName:    fileName,
		loader:      loader,
		fileWatcher: fw,
	}, nil
}

func (r *reloader) Watch(ctx context.Context) error {
	defer func() {
		if err := r.fileWatcher.Close(); err != nil {
			panic(err)
		}
	}()

	evChan := r.fileWatcher.FileChangeEventsChannel()
	errChan := r.fileWatcher.ErrorsChannel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-evChan:
			if (e.Op & fsnotify.Write) == fsnotify.Write {
				fmt.Println("WRITE")
				err := r.loader.Load()
				if err != nil {
					return err
				}
			}
		case err := <-errChan:
			return err
		}
	}

}
