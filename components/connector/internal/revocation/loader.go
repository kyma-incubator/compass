package revocation

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const reconnectInternal = time.Second * 5

type Loader interface {
	Run()
}

type revocationListLoader struct {
	revocationListCache Cache
	manager             Manager
	configMapName       string
}

func NewRevocationListLoader(revocationListCache Cache,
	manager Manager,
	configMapName string,
) Loader {
	return &revocationListLoader{
		revocationListCache: revocationListCache,
		manager:             manager,
		configMapName:       configMapName,
	}
}

func (cl *revocationListLoader) Run() {
	cl.startKubeWatch()
}

func (cl *revocationListLoader) startKubeWatch() {
	for {
		watcher := connectWatch(cl.manager, cl.configMapName)
		log.Println("Waiting for revocation list configmap events...")
		if err := cl.processEvents(watcher.ResultChan()); err != nil {
			// Cleanup any allocated resources
			watcher.Stop()
			time.Sleep(reconnectInternal)
		}
	}
}

func (cl *revocationListLoader) processEvents(ch <-chan watch.Event) error {
	for {
		ev, ok := <-ch
		if !ok {
			return errors.New("reconnect watcher")
		}
		if err := cl.processEvenType(ev); err != nil {
			return err
		}
	}
}

func (cl *revocationListLoader) processEvenType(ev watch.Event) error {
	switch ev.Type {
	case watch.Added:
		fallthrough
	case watch.Modified:
		log.Println("revocation list updated")
		config, ok := ev.Object.(*v1.ConfigMap)
		if !ok {
			log.Println("Unexpected error: object is not configmap. Try again")
			return nil
		}
		cl.revocationListCache.Put(config.Data)
		log.Debug("New configmap is:", config.Data)
	case watch.Deleted:
		log.Println("revocation list deleted")
		cl.revocationListCache.Put(make(map[string]string))
	case watch.Error:
		log.Println("Error event is received, stop revocation list configmap watcher...")
		return errors.New("reconnect watcher")
	}
	return nil
}

func connectWatch(manager Manager, configMapName string) watch.Interface {
	log.Println("Starting watcher for revocation list configmap changes...")
	for {
		watcher, err := manager.Watch(metav1.ListOptions{
			FieldSelector: "metadata.name=" + configMapName,
			Watch:         true,
		})
		if err != nil {
			log.Printf("Could not initialize watcher: %s. Sleep for %s and try again...", err, reconnectInternal.String())
			time.Sleep(reconnectInternal)
		}
		return watcher
	}
}
