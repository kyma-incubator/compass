package revocation

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

type Loader interface {
	Run(ctx context.Context)
}

type revocationListLoader struct {
	revocationListCache Cache
	manager             Manager
	configMapName       string
	reconnectInternal   time.Duration
}

func NewRevocationListLoader(revocationListCache Cache,
	manager Manager,
	configMapName string,
	reconnectInternal time.Duration,
) Loader {
	return &revocationListLoader{
		revocationListCache: revocationListCache,
		manager:             manager,
		configMapName:       configMapName,
		reconnectInternal:   reconnectInternal,
	}
}

func (rl *revocationListLoader) Run(ctx context.Context) {
	rl.startKubeWatch(ctx)
}

func (rl *revocationListLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled, stop revocation config map watcher")
			return
		default:
		}
		watcher, err := connectWatch(rl.manager, rl.configMapName)
		if err != nil {
			log.Printf("Could not initialize watcher: %s. Sleep for %s and try again...", err, rl.reconnectInternal.String())
			time.Sleep(rl.reconnectInternal)
			continue
		}
		log.Println("Waiting for revocation list configmap events...")
		for {
			if !rl.processEvents(ctx, watcher.ResultChan()) {
				// Cleanup any allocated resources
				log.Printf("Stopping revocation configmap watcher")
				watcher.Stop()
				time.Sleep(rl.reconnectInternal)
				break
			}
		}
	}
}

func (rl *revocationListLoader) processEvents(ctx context.Context, events <-chan watch.Event) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case ev, ok := <-events:
			if !ok {
				return false
			}
			switch ev.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				log.Println("revocation list updated")
				config, ok := ev.Object.(*v1.ConfigMap)
				if !ok {
					log.Println("Unexpected error: object is not configmap. Try again")
					return true
				}
				rl.revocationListCache.Put(config.Data)
				log.Debug("New configmap is:", config.Data)
			case watch.Deleted:
				log.Println("revocation list deleted")
				rl.revocationListCache.Put(make(map[string]string))
			case watch.Error:
				log.Println("Error event is received, stop revocation list configmap watcher and try again...")
				return false
			}
		}
	}
}

func connectWatch(manager Manager, configMapName string) (watch.Interface, error) {
	log.Println("Starting watcher for revocation list configmap changes...")
	for {
		watcher, err := manager.Watch(metav1.ListOptions{
			FieldSelector: "metadata.name=" + configMapName,
			Watch:         true,
		})
		return watcher, err
	}
}
