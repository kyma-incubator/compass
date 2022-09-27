package certloader

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const certsListLoaderCorrelationID = "cert-loader-id"

// Loader provide mechanism to load certificate data into in-memory storage
type Loader interface {
	Run(ctx context.Context)
}

// Manager is a kubernetes secret manager that has methods to work with secret resources
//go:generate mockery --name=Manager --output=automock --outpkg=automock --case=underscore --disable-version-string
type Manager interface {
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

type certificateLoader struct {
	config            Config
	certCache         *certificateCache
	secretManagers    []Manager
	secretNames       []string
	reconnectInterval time.Duration
}

// NewCertificateLoader creates new certificate loader which is responsible to watch a secret containing client certificate
// and update in-memory cache with that certificate if there is any change
func NewCertificateLoader(config Config, certCache *certificateCache, secretManagers []Manager, secretNames []string, reconnectInterval time.Duration) Loader {
	return &certificateLoader{
		config:            config,
		certCache:         certCache,
		secretManagers:    secretManagers,
		secretNames:       secretNames,
		reconnectInterval: reconnectInterval,
	}
}

// StartCertLoader prepares and run certificate loader goroutine
func StartCertLoader(ctx context.Context, certLoaderConfig Config) (Cache, error) {
	parsedCertSecret, err := namespacedname.Parse(certLoaderConfig.ExternalClientCertSecret)
	if err != nil {
		return nil, err
	}

	parsedExtSvcCertSecret, err := namespacedname.Parse(certLoaderConfig.ExtSvcClientCertSecret)
	if err != nil {
		return nil, err
	}

	kubeConfig := kubernetes.Config{}
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, kubeConfig.PollInterval, kubeConfig.PollTimeout, kubeConfig.Timeout)
	if err != nil {
		return nil, err
	}

	certCache := NewCertificateCache()
	secretManagers := []Manager{k8sClientSet.CoreV1().Secrets(parsedCertSecret.Namespace), k8sClientSet.CoreV1().Secrets(parsedExtSvcCertSecret.Namespace)}
	secretNames := []string{parsedCertSecret.Name, parsedExtSvcCertSecret.Name}

	certLoader := NewCertificateLoader(certLoaderConfig, certCache, secretManagers, secretNames, time.Second)
	go certLoader.Run(ctx)

	return certCache, nil
}

// Run uses kubernetes watch mechanism to listen for resource changes and update certificate cache
func (cl *certificateLoader) Run(ctx context.Context) {
	entry := log.C(ctx)
	entry = entry.WithField(log.FieldRequestID, certsListLoaderCorrelationID)
	ctx = log.ContextWithLogger(ctx, entry)

	cl.startKubeWatch(ctx)
}

func (cl *certificateLoader) startKubeWatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.C(ctx).Info("Context cancelled, stopping certificate watcher...")
			return
		default:
		}
		log.C(ctx).Info("Starting certificate watchers for secret changes...")

		wg := &sync.WaitGroup{}
		for idx, manager := range cl.secretManagers {
			wg.Add(1)

			go func(manager Manager, idx int) {
				defer wg.Done()

				watcher, err := manager.Watch(ctx, metav1.ListOptions{
					FieldSelector: "metadata.name=" + cl.secretNames[idx],
					Watch:         true,
				})

				if err != nil {
					log.C(ctx).WithError(err).Errorf("Could not initialize watcher. Sleep for %s and try again... %v", cl.reconnectInterval.String(), err)
					time.Sleep(cl.reconnectInterval)
					return
				}
				log.C(ctx).Info("Waiting for certificate secret events...")

				cl.processEvents(ctx, watcher.ResultChan(), cl.secretNames[idx])

				// Cleanup any allocated resources
				watcher.Stop()
				time.Sleep(cl.reconnectInterval)
			}(manager, idx)
		}

		wg.Wait()
	}
}

func (cl *certificateLoader) processEvents(ctx context.Context, events <-chan watch.Event, secretName string) {
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
				log.C(ctx).Info("Updating the cache with certificate data...")
				secret, ok := ev.Object.(*v1.Secret)
				if !ok {
					log.C(ctx).Error("Unexpected error: object is not secret. Try again")
					continue
				}
				tlsCert, err := parseCertificate(ctx, secret.Data, cl.config)
				if err != nil {
					log.C(ctx).WithError(err).Error("Fail during certificate parsing")
				}
				cl.certCache.put(secretName, tlsCert)
			case watch.Deleted:
				log.C(ctx).Info("Removing certificate secret data from cache...")
				cl.certCache.put(secretName, nil)
			case watch.Error:
				log.C(ctx).Error("Error event is received, stop certificate secret watcher and try again...")
				return
			}
		}
	}
}

func parseCertificate(ctx context.Context, secretData map[string][]byte, config Config) (*tls.Certificate, error) {
	log.C(ctx).Info("Parsing provided certificate data...")
	certChainBytes, existsCertKey := secretData[config.ExternalClientCertCertKey]
	privateKeyBytes, existsKeyKey := secretData[config.ExternalClientCertKeyKey]

	if existsCertKey && existsKeyKey {
		return cert.ParseCertificateBytes(certChainBytes, privateKeyBytes)
	}

	extSvcCertChainBytes, existsExtSvcCertKey := secretData[config.ExtSvcClientCertCertKey]
	extSvcPrivateKeyBytes, existsExtSvcKeyKey := secretData[config.ExtSvcClientCertKeyKey]

	if existsExtSvcCertKey && existsExtSvcKeyKey {
		return cert.ParseCertificateBytes(extSvcCertChainBytes, extSvcPrivateKeyBytes)
	}

	return nil, errors.New("There is no certificate data provided")
}
