package revocation

import (
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
	watcher := startWatch(cl.manager, cl.configMapName)
	ch := watcher.ResultChan()
	go func() {
		for {
			log.Println("Waiting for revocation list configmap events...")
			ev, ok := <-ch
			if !ok {
				log.Println("revocation list events channel is closed")
				// Cleanup any allocated resources
				watcher.Stop()
				time.Sleep(reconnectInternal)

				watcher = startWatch(cl.manager, cl.configMapName)
				ch = watcher.ResultChan()
				continue
			}
			switch ev.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				log.Println("revocation list updated")
				config, ok := ev.Object.(*v1.ConfigMap)
				if !ok {
					log.Println("Unexpected error: object is not configmap. Try again")
					continue
				}
				cl.revocationListCache.Put(config.Data)
				log.Debug("New configmap is:", config.Data)
			case watch.Deleted:
				log.Println("revocation list deleted")
				cl.revocationListCache.Put(make(map[string]string))
			case watch.Error:
				log.Println("Error event is received, stop revocation list configmap watcher...")
				watcher.Stop()
				watcher = startWatch(cl.manager, cl.configMapName)
				ch = watcher.ResultChan()
			}
		}
	}()
}

func startWatch(manager Manager, configMapName string) watch.Interface {
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
