package revocation

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const revocationListLoaderCorrelationID = "revocation-list-loader"

//go:generate mockery --name=Manager --output=automock --outpkg=automock --case=underscore
type Manager interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

//go:generate mockery --name=RevokedCertificatesCache --output=automock --outpkg=automock --case=underscore
type RevokedCertificatesCache interface {
	Put(data map[string]string)
}

type Loader interface {
	Run(ctx context.Context)
}

type revokedCertificatesLoader struct {
	revokedCertsCache RevokedCertificatesCache
	configMapManager  Manager
	configMapName     string
	reconnectInterval time.Duration
}

func NewRevokedCertificatesLoader(
	revokedCertsCache RevokedCertificatesCache,
	configMapManager Manager,
	configMapName string,
	reconnectInterval time.Duration,
) Loader {
	return &revokedCertificatesLoader{
		revokedCertsCache: revokedCertsCache,
		configMapManager:  configMapManager,
		configMapName:     configMapName,
		reconnectInterval: reconnectInterval,
	}
}

func (rl *revokedCertificatesLoader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, revocationListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	rl.startKubeWatch(ctx)
}

func (rl *revokedCertificatesLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping revocation config map watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting watcher for revocation list configmap changes...")
		watcher, err := rl.configMapManager.Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + rl.configMapName,
			Watch:         true,
		})
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Could not initialize watcher. Sleep for %s and try again... %v", rl.reconnectInterval.String(), err)
			time.Sleep(rl.reconnectInterval)
			continue
		}
		log.C(ctx).Info("Waiting for revocation list configmap events...")

		rl.processEvents(ctx, watcher.ResultChan())

		// Cleanup any allocated resources
		watcher.Stop()
		time.Sleep(rl.reconnectInterval)
	}
}

func (rl *revokedCertificatesLoader) processEvents(ctx context.Context, events <-chan watch.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			switch ev.Type {
			case watch.Added:
				fallthrough
			case watch.Modified:
				log.C(ctx).Info("Revocation list updated")
				config, ok := ev.Object.(*v1.ConfigMap)
				if !ok {
					log.C(ctx).Error("Unexpected error: object is not configmap. Try again")
					continue
				}
				rl.revokedCertsCache.Put(config.Data)
				log.C(ctx).Debugf("New configmap is: %s", config.Data)
			case watch.Deleted:
				log.C(ctx).Info("Revocation list deleted")
				rl.revokedCertsCache.Put(make(map[string]string))
			case watch.Error:
				log.C(ctx).Error("Error event is received, stop revocation list configmap watcher and try again...")
				return
			}
		}
	}
}
