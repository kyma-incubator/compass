package certloader

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const certsListLoaderCorrelationID = "certs-list-loader"

type Loader interface {
	Run(ctx context.Context)
}

//go:generate mockery --name=Manager
type Manager interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type certificatesLoader struct {
	certsCache        Cache
	secretManager     Manager
	secretName        string
	reconnectInterval time.Duration
}

func NewCertificatesLoader(certsCache Cache, secretManager Manager, secretName string, reconnectInterval time.Duration) Loader {
	return &certificatesLoader{
		certsCache:        certsCache,
		secretManager:     secretManager,
		secretName:        secretName,
		reconnectInterval: reconnectInterval,
	}
}

func (cl *certificatesLoader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	cl.startKubeWatch(ctx)
}

func (cl *certificatesLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping certificate watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting certificate watcher for secret changes...")
		watcher, err := cl.secretManager.Watch(ctx, metav1.ListOptions{
			FieldSelector: "metadata.name=" + cl.secretName,
			Watch:         true,
		})
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Could not initialize watcher. Sleep for %s and try again... %v", cl.reconnectInterval.String(), err)
			time.Sleep(cl.reconnectInterval)
			continue
		}
		log.C(ctx).Info("Waiting for certificate secret events...")

		cl.processEvents(ctx, watcher.ResultChan())

		// Cleanup any allocated resources
		watcher.Stop()
		time.Sleep(cl.reconnectInterval)
	}
}

func (cl *certificatesLoader) processEvents(ctx context.Context, events <-chan watch.Event) {
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
				log.C(ctx).Info("Certificate secret is updated")
				secret, ok := ev.Object.(*v1.Secret)
				if !ok {
					log.C(ctx).Error("Unexpected error: object is not secret. Try again")
					continue
				}
				cl.certsCache.Put(secret.Data)
			case watch.Deleted:
				log.C(ctx).Info("Certificate secret is deleted")
				cl.certsCache.Put(make(map[string][]byte))
			case watch.Error:
				log.C(ctx).Error("Error event is received, stop certificate secret watcher and try again...")
				return
			}
		}
	}
}
